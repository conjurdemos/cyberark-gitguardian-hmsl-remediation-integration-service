package main

import (
	"flag"
	"log"
	"os"

	"github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/credentialprovider"
)

var (
	version         = "dev"
	TLS_SKIP_VERIFY = false

	DEBUG = false
)

func main() {
	safe := flag.String("safename", "", "safe name from which to fetch integration host settings")
	hostobjname := flag.String("hostobjname", "", "Integration Host Account object name")
	dbobjname := flag.String("dbobjname", "", "Integration Database Account object name")
	appid := flag.String("appid", "", "Integration Host Application ID")

	debug := flag.Bool("d", false, "Enable debug settings")
	ver := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *ver {
		log.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	DEBUG = *debug

	client := credentialprovider.Client{}

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
	hostprops := credentialprovider.NewProperties("PASSWORD", *safe, *appid, *hostobjname, hostattrs)
	hostpropvals, err := client.FetchProperties(hostprops)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	for k, v := range hostpropvals.Attributes {
		log.Printf("Prop: %s, Val: %s (%s)\n", k, v, hostpropvals.Attributes[k])
	}

	dbattrs := []string{
		"PassProps.DSN",
	}
	dbprops := credentialprovider.NewProperties("PASSWORD", *safe, *appid, *dbobjname, dbattrs)
	dbpropvals, err := client.FetchProperties(dbprops)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	for k, v := range dbpropvals.Attributes {
		log.Printf("Prop: %s, Val: %s (%s)\n", k, v, dbpropvals.Attributes[k])
	}
}
