package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidh-cyberark/brimstone/pkg/brimstone"
	bs "github.com/davidh-cyberark/brimstone/pkg/brimstone"
	gg "github.com/davidh-cyberark/brimstone/pkg/gitguardian"
	"github.com/davidh-cyberark/brimstone/pkg/hailstone"
	hmsl "github.com/davidh-cyberark/brimstone/pkg/hasmysecretleaked"
	pam "github.com/davidh-cyberark/brimstone/pkg/privilegeaccessmanager"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/caarlos0/env/v10"
)

const (
	GG_HEADER = "Gitguardian-Signature"
)

var (
	version             = "dev"
	LastRefreshAccounts = time.UnixMilli(0)
)

type config struct {
	HmslUrl        string `env:"HMSL_URL" envDefault:"https://api.hasmysecretleaked.com"`
	AudienceType   string `env:"HMSL_AUDIENCE_TYPE" envDefault:"hmsl"`
	GgApiUrl       string `env:"GG_API_URL" envDefault:"https://api.gitguardian.com"`
	GgApiToken     string `env:"GG_API_TOKEN,unset"`
	GgWebhookToken string `env:"GG_WEBHOOK_TOKEN,required,unset"`
	ApiKey         string `env:"BRIMSTONE_API_KEY,unset"`

	DbUrl string `env:"DB_URL,required,unset"`
	Port  uint16 `env:"PORT" envDefault:"9191"`

	RefreshAccountsEvery time.Duration `env:"REFRESH_ACCOUNTS_EVERY" envDefault:"1h"`

	brimstone.BaseConfig
}

type CustomContext struct {
	echo.Context
	Config        config
	ServerAddress string
	Brimstone     bs.Brimstone
}

func RefreshAccountsRegularly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		e := c.Echo()
		cc := c.(*CustomContext)
		e.Logger.Debugf("Last refresh of accounts: %s", time.Since(LastRefreshAccounts).String())
		if cc.Config.RefreshAccountsEvery > 0 && time.Since(LastRefreshAccounts) > cc.Config.RefreshAccountsEvery {
			e.Logger.Debug("Refreshing accounts")
			LastRefreshAccounts = time.Now()
			pamClient := pam.NewClient(cc.Brimstone.PAMConfig.PCloudURL, *cc.Brimstone.PAMConfig)
			batches, err := hailstone.LoadAllAccounts(&pamClient)
			if err != nil {
				e.Logger.Error(err)
			} else {
				for _, batch := range batches {
					cc.Brimstone.SaveHashBatch(batch)
				}
			}
		}
		return next(cc)
	}
}

func main() {
	e := echo.New()
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		e.Logger.Fatalf("failed to parse config: %+v\n", err)
	}

	ver := flag.Bool("version", false, "Print version")
	debug := flag.Bool("d", false, "Enable debug settings")
	flag.Parse()

	if debug != nil && *debug {
		e.Debug = true
	}

	pamconfig := pam.Config{
		IDTenantURL:     cfg.IdTenantUrl,
		PCloudURL:       cfg.PcloudUrl,
		SafeName:        cfg.SafeName,
		PlatformID:      cfg.PlatformID,
		User:            cfg.PamUser,
		Pass:            cfg.PamPass,
		TLS_SKIP_VERIFY: cfg.TlsSkipVerify,
	}

	// Connect to the database
	var db *gorm.DB
	{
		var err error
		if strings.HasPrefix(cfg.DbUrl, "postgres://") {
			// https://www.cockroachlabs.com/docs/v23.1/connection-parameters
			db, err = gorm.Open(postgres.Open(cfg.DbUrl), &gorm.Config{})
		} else if strings.HasPrefix(cfg.DbUrl, "sqlite://") {
			dbFilename := strings.TrimPrefix(cfg.DbUrl, "sqlite://")
			db, err = gorm.Open(sqlite.Open(dbFilename), &gorm.Config{})
		} else {
			e.Logger.Fatalf("unsupported database url: %s", cfg.DbUrl)
		}
		if err != nil {
			e.Logger.Fatal("failed to connect database")
		}
	}

	ctx := context.TODO()
	clientWithResponses, errClient := hmsl.NewClientAuthenticateWithGitGuardian(ctx, &cfg.HmslUrl, &cfg.AudienceType, &cfg.GgApiUrl, &cfg.GgApiToken)
	if errClient != nil {
		e.Logger.Fatalf("failed to create HMSL client: %s", errClient)
	}

	br := bs.Brimstone{
		Db:         db,
		HMSLClient: clientWithResponses,
		PAMConfig:  &pamconfig,
	}

	server_addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(int(cfg.Port)))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c, cfg, server_addr, br}
			return next(cc)
		}
	})

	// Log all requests
	e.Use(middleware.Logger())
	if *ver {
		e.Logger.Printf("Version: %s\n", version)
		os.Exit(0)
	}
	e.Use(middleware.Recover())
	e.Use(RefreshAccountsRegularly)

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
			cc := c.(*CustomContext)
			return key == cc.Config.ApiKey, nil
		},
	}))

	bs.RegisterHandlers(e, br)

	if initdbErr := br.InitializeDb(); initdbErr != nil {
		e.Logger.Fatalf("failed to initialize database: %s", initdbErr)
	}

	e.Logger.Fatal(e.Start(server_addr))
}

func GGValidator(ggsig string, c echo.Context) (bool, error) {
	cc := c.(*CustomContext)
	ggts := c.Request().Header.Get("timestamp")
	if !strings.HasPrefix(ggsig, "sha256=") {
		return false, fmt.Errorf("bad signature")
	}
	bodyBytes, bbErr := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // need to put the bytes reader back on the Body
	if bbErr != nil {
		return false, fmt.Errorf("unable to read request body")
	}
	if ok := gg.ValidateGGPayload(ggsig, ggts, cc.Config.GgWebhookToken, bodyBytes); !ok {
		return false, fmt.Errorf("bad gg request")
	}
	return true, nil
}
