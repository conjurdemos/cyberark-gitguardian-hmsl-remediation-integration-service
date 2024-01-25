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
	version                  = "dev"
	TLS_SKIP_VERIFY          = false
	DEBUG                    = false
	API_KEY_VARNAME          = "BRIMSTONE_API_KEY"
	GG_API_TOKEN_VARNAME     = "GG_API_TOKEN"
	GG_WEBHOOK_TOKEN_VARNAME = "GG_WEBHOOK_TOKEN"

	// postgres://<username>:<password>@<host>:<port>/<database>?<parameters>
	LOCAL_DATABASE_URL = "postgresql://root@localhost:26257/brimstone?sslmode=disable&application_name=brimstone"
)

func main() {
	url := flag.String("hmslurl", "https://api.hasmysecretleaked.com", "HMSL url where to send hashes (Used as audience when sending JWT request)")
	audiencetype := flag.String("hmslaudtype", "hmsl", "Audience type for HMSL JWT request")
	ggapiurl := flag.String("ggapiurl", "https://api.gitguardian.com", "GG API URL")
	ggapitokenvar := flag.String("ggapitokenvar", GG_API_TOKEN_VARNAME, "GG API Token env var contains GG API token to use (default: GG_API_TOKEN)")
	ggwebhooktokenvar := flag.String("ggwebhooktokenvar", GG_WEBHOOK_TOKEN_VARNAME, "GG API Token env var contains GG custom webhook token from settings (default: GG_WEBHOOK_TOKEN)")

	apikeyvarname := flag.String("keyvar", API_KEY_VARNAME, "Name of env var from which to set brimstone's api key (default: BRIMSTONE_API_KEY)")

	dburl := flag.String("dburl", LOCAL_DATABASE_URL, "Database URL")
	port := flag.String("port", "9191", "port whereon brimstone listens (default: 9191)")

	idtenanturl := flag.String("idtenanturl", "", "PAM config ID tenant URL")
	pcloudurl := flag.String("pcloudurl", "", "PAM config Privilege Cloud URL")
	pamuser := flag.String("pamuser", "", "PAM config PAM User")
	pampass := flag.String("pampass", "", "PAM config PAM Pass")
	safename := flag.String("safename", "", "PAM config PAM Safe Name")

	tlsskipverify := flag.Bool("tls-skip-verify", false, "Skip TLS Verify when calling pam (for self-signed cert)")

	ver := flag.Bool("version", false, "Print version")
	debug := flag.Bool("d", false, "Enable debug settings")
	flag.Parse()

	databaseurl := *dburl
	API_KEY_VARNAME = *apikeyvarname
	GG_API_TOKEN_VARNAME = *ggapitokenvar
	GG_WEBHOOK_TOKEN_VARNAME = *ggwebhooktokenvar

	TLS_SKIP_VERIFY = *tlsskipverify
	DEBUG = *debug

	pamconfig := pam.Config{
		IDTenantURL:     *idtenanturl,
		PCloudURL:       *pcloudurl,
		SafeName:        *safename,
		PlatformID:      "DummyPlatform", // not enough info from GG incident to make this more specific
		User:            *pamuser,
		Pass:            *pampass,
		TLS_SKIP_VERIFY: TLS_SKIP_VERIFY,
	}

	e := echo.New()

	// Log all requests
	e.Use(middleware.Logger())
	if *ver {
		e.Logger.Printf("Version: %s\n", version)
		os.Exit(0)
	}
	_, ok := GetAPIKey()
	if !ok {
		log.Fatalf("must set env var, %s, with api key", API_KEY_VARNAME)
	}

	ggapitoken, ggok := GetGGAPIToken()
	if !ggok {
		log.Fatalf("must set env var, %s, with GG api token", GG_API_TOKEN_VARNAME)
	}
	if _, ggok := GetGGWebhookToken(); !ggok {
		log.Fatalf("must set env var, %s, with GG webhook token", GG_WEBHOOK_TOKEN_VARNAME)
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
			return GGValidator(key, c)
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
			brimstoneapikey, ok := GetAPIKey()
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

	clientWithResponses, errClient := hmsl.NewClientAuthenticateWithGitGuardian(ctx, url, audiencetype, ggapiurl, &ggapitoken)
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

	e.Logger.Fatal(e.Start(net.JoinHostPort("0.0.0.0", *port)))
}

func GetAPIKey() (string, bool) {
	return os.LookupEnv(API_KEY_VARNAME)
}
func GetGGAPIToken() (string, bool) {
	return os.LookupEnv(GG_API_TOKEN_VARNAME)
}
func GetGGWebhookToken() (string, bool) {
	return os.LookupEnv(GG_WEBHOOK_TOKEN_VARNAME)
}

func GGValidator(ggsig string, c echo.Context) (bool, error) {
	ggts := c.Request().Header.Get("timestamp")
	if !strings.HasPrefix(ggsig, "sha256=") {
		return false, fmt.Errorf("bad signature")
	}
	bodyBytes, bbErr := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // need to put the bytes reader back on the Body
	if bbErr != nil {
		return false, fmt.Errorf("unable to read request body")
	}
	webhooktoken, ok := GetGGWebhookToken()
	if !ok {
		return false, fmt.Errorf("no webhook token set")
	}
	if ok := gg.ValidateGGPayload(ggsig, ggts, webhooktoken, bodyBytes); !ok {
		return false, fmt.Errorf("bad gg request")
	}
	return true, nil
}
