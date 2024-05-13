package privilegeaccessmanager

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/utils"
)

type Config struct {
	IDTenantURL     string
	PCloudURL       string
	SafeName        string
	User            string
	Pass            string
	PlatformID      string
	TLS_SKIP_VERIFY bool
}

type Session struct {
	Token      string
	TokenType  string
	Expiration time.Time
}

// PAMClient contains the data necessary for requests to pass successfully
type Client struct {
	BaseURL  string
	AuthType string
	Session  Session
	Config   Config
}

type Safe struct {
	SafeNumber                int64   `json:"safeNumber,omitempty"`
	Location                  string  `json:"location,omitempty"`
	Creator                   Creator `json:"creator,omitempty"`
	OlacEnabled               bool    `json:"olacEnabled,omitempty"`
	NumberOfVersionsRetention int     `json:"numberOfVersionsRetention,omitempty"`
	NumberOfDaysRetention     int     `json:"numberOfDaysRetention,omitempty"`
	AutoPurgeEnabled          bool    `json:"autoPurgeEnabled,omitempty"`
	CreationTime              int64   `json:"creationTime,omitempty"`
	LastModificationTime      int64   `json:"lastModificationTime,omitempty"`
	SafeUrlId                 string  `json:"safeUrlId,omitempty"`
	SafeName                  string  `json:"safeName,omitempty"`
	Description               string  `json:"description,omitempty"`
	ManagingCPM               string  `json:"managingCPM,omitempty"`
	IsExpiredMember           bool    `json:"isExpiredMember,omitempty"`
}

type Creator struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Account struct {
	ID                        string                    `json:"id,omitempty"`
	Name                      string                    `json:"name,omitempty"`
	CategoryModificationTime  int64                     `json:"categoryModificationTime,omitempty"`
	PlatformId                string                    `json:"platformId,omitempty"`
	SafeName                  string                    `json:"safeName,omitempty"`
	Address                   string                    `json:"address,omitempty"`
	UserName                  string                    `json:"userName,omitempty"`
	SecretType                string                    `json:"secretType,omitempty"`
	Secret                    string                    `json:"secret,omitempty"`
	CreatedTime               int64                     `json:"createdTime,omitempty"`
	DeletionTime              int64                     `json:"deletionTime,omitempty"`
	PlatformAccountProperties PlatformAccountProperties `json:"platformAccountProperties,omitempty"`
	SecretManagement          SecretManagement          `json:"secretManagement,omitempty"`
	RemoteMachinesAccess      RemoteMachinesAccess      `json:"remoteMachinesAccess,omitempty"`
}

type PostAddAccountRequest struct {
	Name                      string                    `json:"name,omitempty"`
	Address                   string                    `json:"address,omitempty"`
	UserName                  string                    `json:"userName,omitempty"`
	PlatformID                string                    `json:"platformId,omitempty"` // Required
	SafeName                  string                    `json:"safeName,omitempty"`   // Required
	SecretType                string                    `json:"secretType,omitempty"`
	Secret                    string                    `json:"secret,omitempty"`
	PlatformAccountProperties PlatformAccountProperties `json:"platformAccountProperties,omitempty"`
	SecretManagement          SecretManagementRequest   `json:"secretManagement,omitempty"`
	RemoteMachinesAccess      RemoteMachinesAccess      `json:"remoteMachinesAccess,omitempty"`
}

type PlatformAccountProperties struct {
	IncidentDetails string `json:"IncidentDetails,omitempty"`
}

type SecretManagementRequest struct {
	AutomaticManagementEnabled bool   `json:"automaticManagementEnabled,omitempty"`
	ManualManagementReason     string `json:"manualManagementReason,omitempty"`
}

type RemoteMachinesAccess struct {
	RemoteMachines                   string `json:"remoteMachines,omitempty"`
	AccessRestrictedToRemoteMachines bool   `json:"accessRestrictedToRemoteMachines,omitempty"`
}

type PostAddAccountResponse struct {
	ID                       string                   `json:"id"`
	SafeName                 string                   `json:"safeName,omitempty"`
	PlatformID               string                   `json:"platformId,omitempty"`
	Address                  string                   `json:"address,omitempty"`
	UserName                 string                   `json:"userName,omitempty"`
	Name                     string                   `json:"name,omitempty"`
	SecretType               string                   `json:"secretType,omitempty"`
	SecretManagement         SecretManagementResponse `json:"secretManagement,omitempty"`
	CreatedTime              int                      `json:"createdTime,omitempty"`
	CategoryModificationTime int                      `json:"categoryModificationTime,omitempty"`
}
type SecretManagementResponse struct {
	AutomaticManagementEnabled bool `json:"automaticManagementEnabled,omitempty"`
	LastModifiedTime           int  `json:"lastModifiedTime,omitempty"`
}

type IDTenantResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int64  `json:"expires_in,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type GetSafesResponse struct {
	Value []Safe `json:"value,omitempty"`
	Count int    `json:"count,omitempty"`
}

type FetchAccountsResponse struct {
	Value    []Account `json:"value,omitempty"`
	Count    int       `json:"count,omitempty"`
	NextLink string    `json:"nextLink,omitempty"`
}

type PostPasswordRetrieveRequest struct {
	Reason string `json:"Reason"`
}

type PostChangePasswordImmediatelyRequest struct {
	ChangeEntireGroup bool `json:"ChangeEntireGroup"`
}

type GetAccountsResponse struct {
	Value []Account `json:"value"`
	Count int       `json:"count"`
}
type SecretManagement struct {
	AutomaticManagementEnabled bool   `json:"automaticManagementEnabled,omitempty"`
	ManualManagementReason     string `json:"manualManagementReason,omitempty"`
	Status                     string `json:"status,omitempty"`
	LastModifiedTime           int    `json:"lastModifiedTime,omitempty"`
	LastReconciledTime         int    `json:"lastReconciledTime,omitempty"`
	LastVerifiedTime           int    `json:"lastVerifiedTime,omitempty"`
}

func NewSession(tok string, toktype string, exp time.Time) Session {
	session := Session{
		Token:      tok,
		TokenType:  toktype,
		Expiration: exp,
	}
	return session
}

func NewConfig(idtenanturl string, pcloudurl string, safename string, platformid string, u string, p string, skipverify bool) Config {
	config := Config{
		IDTenantURL:     idtenanturl, // Example: "https://EXAMPLE123.id.cyberark.cloud"
		PCloudURL:       pcloudurl,   // Example: "https://EXAMPLE123.privilegecloud.cyberark.cloud"
		SafeName:        safename,
		PlatformID:      platformid,
		User:            u,
		Pass:            p,
		TLS_SKIP_VERIFY: skipverify,
	}
	return config
}

// NewClient - create a client with reasonable defaults
func NewClient(baseurl string, config Config) Client {
	sess := NewSession("", "", time.Now().Add(5*time.Hour)) // default 5hours for sesion token
	client := Client{
		BaseURL:  baseurl,
		AuthType: "",
		Session:  sess,
		Config:   config,
	}
	return client
}

func (c *Client) FetchAccounts() ([]Account, error) {
	var accounts []Account
	apiurl := fmt.Sprintf("%s/PasswordVault/API/Accounts/", c.Config.PCloudURL) // Use PCloud OAuth

	client := utils.GetHTTPClient(time.Second*30, c.Config.TLS_SKIP_VERIFY)

	req, err := http.NewRequest(http.MethodGet, apiurl, nil)
	if err != nil {
		return accounts, err
	}
	// attach the header
	req.Header = make(http.Header)
	// if token is provided, add header Authorization
	if c.Session.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.Session.TokenType, c.Session.Token))
	}

	res, err := client.Do(req)
	if err != nil {
		return accounts, fmt.Errorf("failed to send request. %s", err)
	}

	// read response body
	body, error := io.ReadAll(res.Body)
	if error != nil {
		log.Println(error)
	}
	// close response body
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return accounts, fmt.Errorf("received non-200 status code '%d'", res.StatusCode)
	}

	FetchAccountsResponse := FetchAccountsResponse{}
	err = json.Unmarshal(body, &FetchAccountsResponse)

	if err == nil {
		for i := 0; i < len(FetchAccountsResponse.Value); i++ {
			accounts = append(accounts, FetchAccountsResponse.Value[i])
		}
	}

	return accounts, nil
}

func (c *Client) GetSessionToken() (string, string, error) {
	identurl := fmt.Sprintf("%s/oauth2/platformtoken", c.Config.IDTenantURL) // Use PCloud OAuth

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.Config.User)
	data.Set("client_secret", c.Config.Pass)
	encodedData := data.Encode()

	client := utils.GetHTTPClient(time.Second*30, c.Config.TLS_SKIP_VERIFY)

	req, err := http.NewRequest(http.MethodPost, identurl, strings.NewReader(encodedData))
	if err != nil {
		log.Fatalf("error in request to get session token: %s", err.Error())
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))
	response, err := client.Do(req)

	body, e := io.ReadAll(response.Body)
	if e != nil {
		log.Fatalf("error reading platform token response: %s", err.Error())
	}
	defer response.Body.Close()

	var idresp IDTenantResponse
	err = json.Unmarshal(body, &idresp)
	if err != nil {
		log.Fatalf("failed to parse json body for platform token: %s\n", err.Error())
	}

	return idresp.TokenType, idresp.AccessToken, nil
}
func (c *Client) RefreshSessionToken() error {
	var err error
	c.Session.TokenType, c.Session.Token, err = c.GetSessionToken()
	return err
}

// POST /api/Accounts/{accountId}/Password/Retrieve
func (c *Client) FetchAccountPassword(accountid string) (string, error) {
	var pass string
	apiurl := fmt.Sprintf("%s/PasswordVault/API/Accounts/%s/Password/Retrieve", c.Config.PCloudURL, accountid)

	client := utils.GetHTTPClient(time.Second*30, c.Config.TLS_SKIP_VERIFY)

	postbody := PostPasswordRetrieveRequest{
		Reason: "HMSL Hash",
	}

	jsonbody, err := json.Marshal(postbody)
	if err != nil {
		log.Fatalf("failed to parse json body for platform token: %s\n", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, apiurl, strings.NewReader(string(jsonbody)))
	if err != nil {
		return pass, err
	}
	// attach the header
	req.Header = make(http.Header)
	// if token is provided, add header Authorization
	if c.Session.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.Session.TokenType, c.Session.Token))
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return pass, fmt.Errorf("failed to send request. %s", err)
	}

	// read response body
	body, error := io.ReadAll(res.Body)
	if error != nil {
		log.Println(error)
	}
	// close response body
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return pass, fmt.Errorf("received non-200 status code '%d'", res.StatusCode)
	}

	// err = json.Unmarshal(body, &pass)
	pass = utils.TrimQuotes(string(body))
	return pass, nil
}

// ChangePasswordImmediately -- https://docs.cyberark.com/PrivCloud-SS/latest/en/Content/WebServices/Change-credentials-immediately.htm
func (c *Client) ChangePasswordImmediately(accountid string) (int, error) {

	// POST /PasswordVault/API/Accounts/<AccountID>/Change/
	apiurl := fmt.Sprintf("%s/PasswordVault/API/Accounts/%s/Change", c.Config.PCloudURL, accountid)

	client := utils.GetHTTPClient(time.Second*30, c.Config.TLS_SKIP_VERIFY)

	postbody := PostChangePasswordImmediatelyRequest{
		ChangeEntireGroup: true,
	}

	jsonbody, err := json.Marshal(postbody)
	if err != nil {
		log.Fatalf("failed to parse json body for platform token: %s\n", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, apiurl, strings.NewReader(string(jsonbody)))
	if err != nil {
		return http.StatusConflict, err
	}
	// attach the header
	req.Header = make(http.Header)
	// if token is provided, add header Authorization
	if c.Session.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.Session.TokenType, c.Session.Token))
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return http.StatusBadGateway, fmt.Errorf("failed to send request. %s", err)
	}

	// read response body
	body, error := io.ReadAll(res.Body)
	if error != nil {
		log.Println(error)
	}
	// close response body
	defer res.Body.Close()
	log.Printf("%s\n", string(body))

	if res.StatusCode >= 300 {
		return res.StatusCode, fmt.Errorf("received non-200 status (code=%d): %s", res.StatusCode, body)
	}

	return http.StatusOK, nil
}

// AddAccount -- https://docs.cyberark.com/PrivCloud-SS/latest/en/Content/WebServices/Add%20Account%20v10.htm
func (c *Client) AddAccount(postbody PostAddAccountRequest) (PostAddAccountResponse, int, error) {

	var newacct PostAddAccountResponse

	// POST /PasswordVault/API/Accounts/
	apiurl := fmt.Sprintf("%s/PasswordVault/API/Accounts/", c.Config.PCloudURL)
	client := utils.GetHTTPClient(time.Second*30, c.Config.TLS_SKIP_VERIFY)

	jsonbody, err := json.Marshal(postbody)
	if err != nil {
		log.Fatalf("failed to parse json body for platform token: %s\n", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, apiurl, strings.NewReader(string(jsonbody)))
	if err != nil {
		return newacct, http.StatusConflict, err
	}
	// attach the header
	req.Header = make(http.Header)
	// if token is provided, add header Authorization
	if c.Session.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", c.Session.TokenType, c.Session.Token))
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return newacct, http.StatusBadGateway, fmt.Errorf("failed to send request. %s", err)
	}

	// read response body
	body, error := io.ReadAll(res.Body)
	if error != nil {
		log.Println(error)
	}
	// close response body
	defer res.Body.Close()
	log.Printf("%s\n", string(body))

	err = json.Unmarshal(body, &newacct)
	if err != nil {
		if res.StatusCode >= 300 {
			return newacct, res.StatusCode, fmt.Errorf("received non-200 status code(%d): %s", res.StatusCode, err.Error())
		}
	}

	return newacct, http.StatusOK, nil
}

// FetchAccountIdFromAccountName fetch all accounts in the safe and iterate through list until accountname is found
func (c *Client) FetchAccountIdFromAccountName(safename string, accountname string) (string, int, error) {

	// https://docs.cyberark.com/PrivCloud-SS/Latest/en/Content/SDK/GetAccounts.htm
	filter := fmt.Sprintf("safeName eq %s", safename)
	apiurl := fmt.Sprintf("%s/PasswordVault/API/Accounts?filter=%s", c.Config.PCloudURL, url.QueryEscape(filter))

	client := utils.GetHTTPClient(time.Second*30, c.Config.TLS_SKIP_VERIFY)

	req, err := http.NewRequest(http.MethodGet, apiurl, nil)
	if err != nil {
		return "", http.StatusConflict, err
	}
	req.Header = make(http.Header)

	// if token is provided, add header Authorization
	authheader := "UNK"
	if c.Session.Token != "" {
		authheader = fmt.Sprintf("%s %s", c.Session.TokenType, c.Session.Token)
		req.Header.Add("Authorization", authheader)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", http.StatusBadGateway, fmt.Errorf("failed to send request. %s", err)
	}
	log.Printf("APIURL: %s\n", apiurl)
	// For debugging, emit an equivalent curl call, example...
	// log.Printf("curl -XGET -H 'Authorization: %s' -H 'Accept: application/json' -H 'Content-Type: application/json' '%s'\n", authheader, apiurl)

	body, error := io.ReadAll(res.Body)
	if error != nil {
		log.Println(error)
	}
	defer res.Body.Close()

	log.Printf("%s\n", string(body))
	var resp GetAccountsResponse
	err = json.Unmarshal(body, &resp)

	if res.StatusCode >= 300 {
		return "", res.StatusCode, fmt.Errorf("received non-200 status (code=%d): %s", res.StatusCode, body)
	}

	foundid := ""
	for i := 0; i < len(resp.Value); i++ {
		if resp.Value[i].Name == accountname {
			return resp.Value[i].ID, http.StatusOK, nil
		}
	}
	if len(foundid) == 0 {
		return "", http.StatusNotFound, nil
	}
	return foundid, http.StatusOK, nil
}
