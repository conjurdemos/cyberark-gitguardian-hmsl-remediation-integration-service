package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	gg "github.com/davidh-cyberark/brimstone/pkg/gitguardian"
)

var (
	version = "dev"
	DEBUG   bool
)

func main() {
	ver := flag.Bool("version", false, "Print version")
	ggapiurl := flag.String("ggapiurl", "https://api.gitguardian.com", "GG API URL, default: https://api.gitguardian.com")
	ggapitoken := flag.String("ggapitoken", "", "API Token from GitGuardian Dashboard")

	flag.Parse()

	if *ver {
		log.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	if len(*ggapitoken) == 0 {
		log.Fatalf("No GG token passed.")
	}

	c, err := gg.NewClient(*ggapiurl)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}
	req, err := gg.NewHealthCheckRequest(*ggapiurl)

	// Authorization: Token GG_API_TOKEN"
	apiToken := fmt.Sprintf("Token %s", *ggapitoken)
	req.Header.Add("Authorization", apiToken)

	resp, respErr := c.Client.Do(req)
	if respErr != nil {
		log.Fatalf("health check faild: %s", err)
	}

	log.Printf("Health Check Response: %s\n", resp.Status)
}
