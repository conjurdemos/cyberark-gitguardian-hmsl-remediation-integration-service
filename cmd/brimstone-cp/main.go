package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	bs "github.com/davidh-cyberark/brimstone/pkg/brimstone"
	cp "github.com/davidh-cyberark/brimstone/pkg/credentialprovider"
	gg "github.com/davidh-cyberark/brimstone/pkg/gitguardian"
	hmsl "github.com/davidh-cyberark/brimstone/pkg/hasmysecretleaked"
	pam "github.com/davidh-cyberark/brimstone/pkg/privilegeaccessmanager"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	GG_HEADER = "Gitguardian-Signature"
)

var (
	version         = "dev"
	TLS_SKIP_VERIFY = false
	DEBUG           = false

	// postgres://<username>:<password>@<host>:<port>/<database>?<parameters>
	LOCAL_DATABASE_URL = "postgresql://root@localhost:26257/brimstone?sslmode=disable&application_name=brimstone"
)

func main() {
	safe := flag.String("safename", "", "Safe name from which to fetch integration host settings")
	hostobjname := flag.String("hostobjname", "", "Integration Host Account object name")
	dbobjname := flag.String("dbobjname", "", "Integration Database Account object name")
	appid := flag.String("appid", "", "Integration Host Application ID")

	url := flag.String("hmslurl", "https://api.hasmysecretleaked.com", "HMSL url where to send hashes (Used as audience when sending JWT request)")
	audiencetype := flag.String("hmslaudtype", "hmsl", "Audience type for HMSL JWT request")

	tlsskipverify := flag.Bool("tls-skip-verify", false, "Skip TLS Verify when calling pam (for self-signed cert)")

	ver := flag.Bool("version", false, "Print version")
	debug := flag.Bool("d", false, "Enable debug settings")
	flag.Parse()

	TLS_SKIP_VERIFY = *tlsskipverify
	DEBUG = *debug

	// Fetch the Integration Host settings from the Credential Provider
	cpclient := cp.Client{}
	hostattrs := []string{
		"PassProps.APIKey",
		"PassProps.Port",
		"PassProps.GitGuardianAPIURL",
		"PassProps.GitGuardianAPIToken",
		"PassProps.GitGuardianWebhookToken",
		"PassProps.IDTenantURL",
		"PassProps.PCloudURL",
		"PassProps.PAMUser",
		"PassProps.PAMPassword",
		"PassProps.PendingSafename",
	}
	hostprops := cp.NewProperties("PASSWORD", *safe, *appid, *hostobjname, hostattrs)
	hostpropvals, err := cpclient.FetchProperties(hostprops)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	if DEBUG {
		for k, v := range hostpropvals.Attributes {
			log.Printf("Prop: %s, Val: %s (%s)\n", k, v, hostpropvals.Attributes[k])
		}
	}

	dbattrs := []string{
		"PassProps.DSN",
	}
	dbprops := cp.NewProperties("PASSWORD", *safe, *appid, *dbobjname, dbattrs)
	dbpropvals, err := cpclient.FetchProperties(dbprops)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	for k, v := range dbpropvals.Attributes {
		log.Printf("Prop: %s, Val: %s (%s)\n", k, v, dbpropvals.Attributes[k])
	}

	databaseurl := dbpropvals.Attributes["PassProps.DSN"]

	// Using the settings from the CP, config the PAM vault
	pamconfig := pam.Config{
		IDTenantURL:     hostpropvals.Attributes["PassProps.IDTenantURL"],
		PCloudURL:       hostpropvals.Attributes["PassProps.PCloudURL"],
		SafeName:        hostpropvals.Attributes["PassProps.PendingSafename"],
		PlatformID:      "DummyPlatform", // PlatformID for new accounts
		User:            hostpropvals.Attributes["PassProps.PAMUser"],
		Pass:            hostpropvals.Attributes["PassProps.PAMPassword"],
		TLS_SKIP_VERIFY: TLS_SKIP_VERIFY,
	}

	e := echo.New()

	// Log all requests
	e.Use(middleware.Logger())
	if *ver {
		e.Logger.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		e.Logger.Printf("REQUEST BODY BEGIN:\n%s\nREQUEST BODY END.\n", string(reqBody))
	}))

	ggapitoken := hostpropvals.Attributes["PassProps.GitGuardianAPIToken"]
	if len(ggapitoken) == 0 {
		e.Logger.Fatalf("Missing GG API token")
	}
	ggapiurl := hostpropvals.Attributes["PassProps.GitGuardianAPIURL"]
	if len(ggapiurl) == 0 {
		e.Logger.Fatalf("Missing GG API URL")
	}
	ggwebhooktoken := hostpropvals.Attributes["PassProps.GitGuardianWebhookToken"]
	if len(ggwebhooktoken) == 0 {
		e.Logger.Fatalf("Missing GG Webhook Token")
	}

	// Configure GG authentication
	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup:  "header:" + GG_HEADER,
		AuthScheme: "",
		Skipper: func(c echo.Context) bool {
			// skip if we do not have a gg sig header
			ggsig := c.Request().Header.Get(GG_HEADER)
			return len(ggsig) == 0
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			return GGValidator(key, c, ggwebhooktoken)
		},
	}))

	// Configure the api key authentication
	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup:  "header:" + echo.HeaderAuthorization,
		AuthScheme: "Bearer",
		Skipper: func(c echo.Context) bool {
			// skip if we have a gg sig header
			ggsig := c.Request().Header.Get(GG_HEADER)
			return len(ggsig) > 0
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			brimstoneapikey, ok := hostpropvals.Attributes["PassProps.APIKey"]
			return (ok && key == brimstoneapikey), nil
		},
	}))

	// sqlite
	// db, err := gorm.Open(sqlite.Open("brimstone.db"), &gorm.Config{})

	// https://www.cockroachlabs.com/docs/v23.1/connection-parameters
	db, err := gorm.Open(postgres.Open(databaseurl), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	ctx := context.TODO()

	clientWithResponses, errClient := hmsl.NewClientAuthenticateWithGitGuardian(ctx, url, audiencetype, &ggapiurl, &ggapitoken)
	if errClient != nil {
		e.Logger.Fatalf("failed to create HMSL client: %s", errClient)
	}

	br := bs.Brimstone{
		Db:         db,
		HMSLClient: clientWithResponses,
		PAMConfig:  &pamconfig,
	}

	bs.RegisterHandlers(e, br)

	if initdbErr := br.InitializeDb(); initdbErr != nil {
		e.Logger.Fatalf("failed to initialize database: %s", initdbErr)
	}

	e.Logger.Fatal(e.Start(net.JoinHostPort("0.0.0.0", hostpropvals.Attributes["PassProps.Port"])))
}

func GGValidator(ggsig string, c echo.Context, webhooktoken string) (bool, error) {
	ggts := c.Request().Header.Get("timestamp")
	if !strings.HasPrefix(ggsig, "sha256=") {
		return false, fmt.Errorf("bad signature")
	}
	bodyBytes, bbErr := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // need to put the bytes reader back on the Body
	if bbErr != nil {
		return false, fmt.Errorf("unable to read request body")
	}

	if ok := gg.ValidateGGPayload(ggsig, ggts, webhooktoken, bodyBytes); !ok {
		return false, fmt.Errorf("bad gg request")
	}
	return true, nil
}
