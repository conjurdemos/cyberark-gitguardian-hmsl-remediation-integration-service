package main

// Hailstone will fetch passwords from PAM, compute hashes, and add the
// hashes to the database

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davidh-cyberark/brimstone/pkg/brimstone"
	hmsl "github.com/davidh-cyberark/brimstone/pkg/hasmysecretleaked"
	pam "github.com/davidh-cyberark/brimstone/pkg/privilegeaccessmanager"
	"github.com/davidh-cyberark/brimstone/pkg/utils"

	"github.com/caarlos0/env/v10"
)

var (
	version = "dev"
	DEBUG   = false
)

type config struct {
	BrimstoneUrl    string `env:"BRIMSTONE_URL" envDefault:"http://127.0.0.1:9191"`
	BrimstoneApiKey string `env:"BRIMSTONE_API_KEY,unset"`

	utils.BaseConfig
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse config: %+v\n", err)
	}

	debug := flag.Bool("d", false, "Enable debug settings")
	ver := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *ver {
		log.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	DEBUG = *debug

	pamconfig := pam.NewConfig(cfg.IdTenantUrl, cfg.PcloudUrl, cfg.SafeName, cfg.PlatformID, cfg.PamUser, cfg.PamPass, cfg.TlsSkipVerify)
	client := pam.NewClient(pamconfig.PCloudURL, pamconfig)
	err := client.RefreshSessionToken()
	if err != nil {
		log.Fatalf("failed to fetch session token: %s", err.Error())
	}

	accounts, err := client.FetchAccounts()
	if err != nil {
		log.Fatalf("failed to fetch all safes: %s", err.Error())
	}

	requests := make(map[string]brimstone.HashBatch)
	for a := range accounts {
		p, e := client.FetchAccountPassword(accounts[a].ID)
		if e != nil {
			log.Printf("error fetching password for account id, %s: %s\n", accounts[a].ID, err.Error())
		}

		hmslhash, hhErr := hmsl.ComputeHash(p)
		if hhErr != nil {
			log.Printf("Error computing HMSL hash, skipping one: %s", err)
			continue
		}
		newhash := brimstone.Hash{
			Hash: hmslhash,
			Name: accounts[a].ID,
		}

		req, ok := requests[accounts[a].SafeName]
		if !ok {
			req = brimstone.HashBatch{
				Safename: accounts[a].SafeName,
			}
		}
		req.Hashes = append(req.Hashes, newhash)
		requests[accounts[a].SafeName] = req
	}

	// Send to brimstone
	for k := range requests {
		err = SendKeysToBrimstone(requests[k], cfg)
		if err != nil {
			log.Printf("Unable to send keys to brimstone: %s\n", err.Error())
		}
	}
}

func SendKeysToBrimstone(keys brimstone.HashBatch, cfg config) error {
	client := utils.GetHTTPClient(time.Second*30, cfg.TlsSkipVerify)

	content, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("failed to serialize keys: %s", err.Error())
	}

	brimstoneEndpoint := fmt.Sprintf("%s/v1/hashes", cfg.BrimstoneUrl)
	req, err := http.NewRequest(http.MethodPut, brimstoneEndpoint, bytes.NewReader(content))
	if err != nil {
		log.Fatalf("error in request to get platform token: %s", err.Error())
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.BrimstoneApiKey))
	req.Header.Add("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return err
	}

	body, e := io.ReadAll(response.Body)
	if e != nil {
		return fmt.Errorf("failed to read response body: %s", err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return fmt.Errorf("failed to send to brimstone: (%s) %s", response.Status, body)
	}
	return nil
}
