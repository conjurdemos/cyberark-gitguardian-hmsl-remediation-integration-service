package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	allaccounts, err := client.FetchAccounts()
	var accounts []pam.Account

	// filter out everything except the safename we're interested in
	for i := 0; i < len(allaccounts); i++ {
		if allaccounts[i].SafeName == *safename {
			accounts = append(accounts, allaccounts[i])
		}
	}

	accountid := "1010_1"
	if len(accounts) > 0 {
		accountid = accounts[0].ID
	}
	curpassword, err := client.FetchAccountPassword(accountid)
	if err != nil {
		log.Printf("ERROR: failed to fetch password for account id, %s: %s\n", accountid, err.Error())
	}
	ActionChangeAccountPassword(&client, accountid)
	newpassword, err := client.FetchAccountPassword(accountid)
	if err != nil {
		log.Printf("ERROR: failed to fetch password after change for account id, %s: %s\n", accountid, err.Error())
	}
	if curpassword == newpassword {
		log.Printf("INFO: passwords are the same: %s=%s\n", curpassword, newpassword)
	}

	list := "abcdefghijklmnopqrstuvwxyz01234567890"
	seq := []rune(list)
	newseq := utils.RandSeq(seq, 6)

	newacctname := fmt.Sprintf("pamclient-%s", newseq)
	newacctsecret := fmt.Sprintf("myleakedsecret-%s", newseq)

	newaccountid, err := ActionAddAccount(&client, newacctname, *safename, newacctsecret, "password", "https://dashboard.gitguardian.com/workspace/000000/incident/00000000")
	log.Printf("New account id: %s\n", newaccountid)

	// Try the cycle with the new account id
	curpassword, err = client.FetchAccountPassword(newaccountid)
	if err != nil {
		log.Printf("ERROR: failed to fetch password for account id, %s: %s\n", newaccountid, err.Error())
	}
	ActionChangeAccountPassword(&client, newaccountid)
	newpassword, err = client.FetchAccountPassword(newaccountid)
	if err != nil {
		log.Printf("ERROR: failed to fetch password after change for account id, %s: %s\n", newaccountid, err.Error())
	}
	if curpassword == newpassword {
		log.Printf("INFO: passwords are the same: %s=%s\n", curpassword, newpassword)
	}
}

func ActionChangeAccountPassword(client *pam.Client, accountid string) {
	code, err := client.ChangePasswordImmediately(accountid)
	if err != nil {
		log.Printf("ERROR: (status code:%d) failed to change password for acct id, %s: %s\n", code, accountid, err.Error())
	}

}

func ActionAddAccount(client *pam.Client, acctname string, safename string, secret string, secrettype string, ggincidenturl string) (string, error) {

	addreq := pam.PostAddAccountRequest{
		Name:                      acctname,
		Address:                   ggincidenturl, // Ex: "https://dashboard.gitguardian.com/workspace/000000/incident/00000000"
		UserName:                  "gitguardian",
		SafeName:                  safename,
		PlatformID:                "DummyPlatform", // PCloud
		Secret:                    secret,
		SecretType:                secrettype,
		PlatformAccountProperties: pam.PlatformAccountProperties{},
	}
	newaccount, _, err := client.AddAccount(addreq)
	if err != nil {
		return "", err
	}
	return newaccount.ID, nil
}
