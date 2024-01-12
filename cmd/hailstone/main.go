package main

// Hailstone will fetch passwords from PAM, compute hashes, and add the
// hashes to the database

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davidh-cyberark/brimstone/pkg/brimstone"
	hmsl "github.com/davidh-cyberark/brimstone/pkg/hasmysecretleaked"
	pam "github.com/davidh-cyberark/brimstone/pkg/privilegeaccessmanager"
	"github.com/davidh-cyberark/brimstone/pkg/utils"
)

var (
	version         = "dev"
	TLS_SKIP_VERIFY = false

	DEBUG = false
)

func main() {
	idtenanturl := flag.String("idtenanturl", "", "PAM config ID tenant URL")
	pcloudurl := flag.String("pcloudurl", "", "PAM config Privilege Cloud URL")
	pamuser := flag.String("pamuser", "", "PAM config PAM User")
	pampass := flag.String("pampass", "", "PAM config PAM Pass")
	safename := flag.String("safename", "", "PAM config PAM Safe Name")

	brimstoneurl := flag.String("brimstoneurl", "http://127.0.0.1:9191/v1/hashes", "Brimstone api url")
	brimstoneapikey := flag.String("brimstoneapikey", "", "Brimstone api key")

	tlsskipverify := flag.Bool("tls-skip-verify", false, "Skip TLS Verify when calling pam (for self-signed cert)")

	debug := flag.Bool("d", false, "Enable debug settings")
	ver := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *ver {
		log.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	TLS_SKIP_VERIFY = *tlsskipverify
	DEBUG = *debug

	pamconfig := pam.NewConfig(*idtenanturl, *pcloudurl, *safename, "DummyPlatform", *pamuser, *pampass, TLS_SKIP_VERIFY)
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
		err = SendKeysToBrimstone(requests[k], *brimstoneurl, *brimstoneapikey)
		if err != nil {
			log.Printf("Unable to send keys to brimstone: %s\n", err.Error())
		}
	}
}

func SendKeysToBrimstone(keys brimstone.HashBatch, brimstoneurl string, brimstoneapikey string) error {
	client := utils.GetHTTPClient(time.Second*30, TLS_SKIP_VERIFY)

	content, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("failed to serialize keys: %s", err.Error())
	}

	req, err := http.NewRequest(http.MethodPut, brimstoneurl, bytes.NewReader(content))
	if err != nil {
		log.Fatalf("error in request to get platform token: %s", err.Error())
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", brimstoneapikey))
	req.Header.Add("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode >= 300 {
		return fmt.Errorf("failed to send to brimstone: (%d) %s", response.StatusCode, response.Status)
	}
	return nil
}
