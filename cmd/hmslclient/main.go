package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	hmsl "github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/hasmysecretleaked"
	"github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/utils"
)

var (
	version = "dev"
	DEBUG   bool
)

func main() {
	url := flag.String("url", "https://api.hasmysecretleaked.com", "HMSL url where to send hashes (Also used as audience when sending JWT request)")
	audiencetype := flag.String("audtype", "hmsl", "Audience type for HMSL JWT request")
	ggapiurl := flag.String("ggapiurl", "https://api.gitguardian.com", "GG API URL, default: https://api.gitguardian.com")
	ggapitoken := flag.String("ggapitoken", "", "API Token from GitGuardian Dashboard (HMSL uses GG for JWT auth)")
	ver := flag.Bool("version", false, "Print version")
	computehash := flag.String("computehash", "", "Compute hash print it out and exit")
	flag.Parse()

	if *ver {
		log.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	if len(*computehash) > 0 {
		hmslhash, err := hmsl.ComputeHash(*computehash)
		if err != nil {
			log.Fatalf("failed to compute HMSL hash: %s\n", err)
		}

		log.Printf("Secret: %s\nHash: %s\n", *computehash, hmslhash)
		os.Exit(0)
	}

	ctx := context.TODO()
	var clientWithResponses *hmsl.ClientWithResponses
	var errClient error

	if len(*ggapitoken) == 0 {
		log.Printf("No GG token passed, using free tier")
		clientWithResponses, errClient = hmsl.NewClientWithResponses(*url)
	} else {
		clientWithResponses, errClient = hmsl.NewClientAuthenticateWithGitGuardian(ctx, url, audiencetype, ggapiurl, ggapitoken)
	}
	if errClient != nil {
		log.Fatalf("failed to create HMSL client: %s", errClient)
	}

	respHealth, respHealthErr := clientWithResponses.HealthzHealthzGetWithResponse(ctx)
	if respHealthErr != nil {
		log.Fatalf("failed call to healthz: %s", respHealthErr)
	}
	log.Printf("Response: %s\n", respHealth.Status())

	respReady, respReadyErr := clientWithResponses.ReadinessReadinessGetWithResponse(ctx)
	if respReadyErr != nil {
		log.Fatalf("failed call to readiness: %s", respReadyErr)
	}
	log.Printf("Readiness Status: (%d) %s\n", respReady.StatusCode(), respReady.Status())

	prefixes := hmsl.BatchPrefixesV1PrefixesPostJSONRequestBody{
		Prefixes: &[]string{"faa11", "bad22", "baac3"},
	}

	respHashes, respHashesErr := clientWithResponses.BatchPrefixesV1PrefixesPostWithResponse(ctx, prefixes)
	if respHashesErr != nil {
		log.Fatalf("failed call to readiness: %s", respHashesErr)
	}
	b, err := json.MarshalIndent(prefixes, "", "  ")
	if err != nil {
		log.Println(err)
	}
	log.Print(string(b))
	log.Printf("batch prefixes v1: (%d) %s\n%s\n", respHashes.StatusCode(), respHashes.Status(), string(respHashes.Body))

	hmslexample(&ctx, clientWithResponses)
}

func hmslexample(ctx *context.Context, client *hmsl.ClientWithResponses) {
	log.Printf("Starting HMSL Example\n")

	// This is copy/pasted from the HMSL docs
	mysecret := "hjshnk5ex5u34565d4654HJKGjhz545d89sjkjak" // #gitleaks:allow

	// 1. We compute the hash as explained above, for example with the following function:
	hmslhash, err := hmsl.ComputeHash(mysecret)
	if err != nil {
		log.Fatalf("failed to compute HMSL hash: %s\n", err)
	}

	log.Printf("Secret: %s\nHash: %s\n", mysecret, hmslhash)

	// 2. I now have 408a5b05c35bb4d230e31da1f9afa0e8881050cb72775e925d6bc7cb945b4f39 as the hash of my secret, so my prefix is 408a5
	if hmslhash != "408a5b05c35bb4d230e31da1f9afa0e8881050cb72775e925d6bc7cb945b4f39" {
		log.Printf("Error: Hash example does not match")
		os.Exit(1)
	}
	log.Printf("Hash matches the example doc hash")

	prefix := utils.FirstN(hmslhash, 5)

	// 3. I use the /v1/prefix/{prefix} endpoint to query HasMySecretLeaked, so I sent a request to https://api.hasmysecretleaked.com/v1/prefix/408a5
	respHashes, respHashesErr := client.GetSecretV1PrefixPrefixGetWithResponse(*ctx, prefix)
	if respHashesErr != nil {
		log.Fatalf("failed call to readiness: %s", respHashesErr)
	}

	log.Printf("Example prefix responses v1: (%d) %s\n%s\n", respHashes.StatusCode(), respHashes.Status(), string(respHashes.Body))

	// hashsum of hmslhash
	myhint, myerr := hmsl.ComputeHint(hmslhash)
	if myerr != nil {
		log.Printf("Error computing hint: %s\n", myerr.Error())
	}
	if myhint != "5de9f935ee515de855551e1786d71f1d7f4c3805083f57dc49863ad624f5ba42" {
		log.Printf("Error computing hint, does not match doc example\n")
	}
	log.Printf("Myhint from mysecret: %s\n", myhint)
	for i := 0; i < len(respHashes.JSON200.Matches); i++ {
		if respHashes.JSON200.Matches[i].Hint == myhint {

			log.Printf("Match Hint: %s\n", respHashes.JSON200.Matches[i].Hint)
			pbytes, pbErr := respHashes.JSON200.Matches[i].Payload.Bytes()
			if pbErr != nil {
				log.Printf("Error converting to bytes: %s\n", pbErr.Error())
			}
			// pstr := string(pbytes)
			msg1, err1 := hmsl.DecryptPayload(pbytes, hmslhash)
			if err1 != nil {
				log.Printf("Error decrypt payload failed with hmslhash: %s\n", err1.Error())
			}

			log.Printf("HMSLHASH: %s\n", msg1)
		}
	}

}
