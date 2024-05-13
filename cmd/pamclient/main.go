package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	pam "github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/privilegeaccessmanager"
	"github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/utils"
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

	listAccountsFlag := flag.Bool("l", false, "List Accounts and exit")
	debug := flag.Bool("d", false, "Enable debug settings")
	ver := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *ver {
		log.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	TLS_SKIP_VERIFY = *tlsskipverify
	DEBUG = *debug
	listAccounts := *listAccountsFlag

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
		if listAccounts {
			log.Printf("ACCT: %+v\n", allaccounts[i])
		}
	}
	if listAccounts {
		os.Exit(0)
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

	newacct, err := ActionAddAccount(&client, newacctname, *safename, newacctsecret, "password", "https://dashboard.gitguardian.com/workspace/000000/incident/00000000")
	newaccountid := newacct.ID
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

	ActionFindAccountIdFromName(&client, *safename, newacctname)
}

func ActionChangeAccountPassword(client *pam.Client, accountid string) {
	code, err := client.ChangePasswordImmediately(accountid)
	if err != nil {
		log.Printf("ERROR: (status code:%d) failed to change password for acct id, %s: %s\n", code, accountid, err.Error())
	}

}

func ActionAddAccount(client *pam.Client, acctname string, safename string, secret string, secrettype string, ggincidenturl string) (pam.PostAddAccountResponse, error) {
	list := "abcdefghijklmnopqrstuvwxyz01234567890"
	seq := []rune(list)
	newseq := utils.RandSeq(seq, 6)

	uname := fmt.Sprintf("mjollnir-%s", newseq)
	addreq := pam.PostAddAccountRequest{
		Name:                      acctname,
		Address:                   ggincidenturl, // Ex: "https://dashboard.gitguardian.com/workspace/000000/incident/00000000"
		UserName:                  uname,
		SafeName:                  safename,
		PlatformID:                "DummyPlatform", // PCloud
		Secret:                    secret,
		SecretType:                secrettype,
		PlatformAccountProperties: pam.PlatformAccountProperties{},
	}
	newaccount, _, err := client.AddAccount(addreq)
	if err != nil {
		return pam.PostAddAccountResponse{}, err
	}
	return newaccount, nil
}

func ActionFindAccountIdFromName(client *pam.Client, safename string, accountname string) {
	err := client.RefreshSessionToken()
	if err != nil {
		log.Printf("Error: %s\n", err.Error())
		return
	}

	acctid, rescode, err := client.FetchAccountIdFromAccountName(safename, accountname)
	if rescode > 299 {
		log.Printf("Result Code=%d\n", rescode)
	}
	log.Printf("Account Name=%s, ID=%s\n", accountname, acctid)
}
