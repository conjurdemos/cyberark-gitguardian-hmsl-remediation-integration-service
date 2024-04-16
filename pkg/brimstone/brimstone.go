package brimstone

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"

	"gorm.io/gorm"

	//"gorm.io/driver/sqlite"
	gg "github.com/davidh-cyberark/brimstone/pkg/gitguardian"
	hmsl "github.com/davidh-cyberark/brimstone/pkg/hasmysecretleaked"
	pam "github.com/davidh-cyberark/brimstone/pkg/privilegeaccessmanager"

	"github.com/davidh-cyberark/brimstone/pkg/utils"
	"github.com/labstack/echo/v4"
	//"github.com/labstack/echo/v4/middleware"
)

// current, then current-1 and current-2, so, 3 total
const MAX_HASH_COUNT = 3

type BaseConfig struct {
	IdTenantUrl string `env:"ID_TENANT_URL,required"`
	PcloudUrl   string `env:"PCLOUD_URL,required"`
	SafeName    string `env:"SAFE_NAME,required"`
	PlatformID  string `env:"PLATFORM_ID,required"`
	PamUser     string `env:"PAM_USER,required"`
	PamPass     string `env:"PAM_PASS,required,unset"`

	TlsSkipVerify bool `env:"TLS_SKIP_VERIFY" envDefault:"false"`
}

type Brimstone struct {
	Db         *gorm.DB
	HMSLClient *hmsl.ClientWithResponses
	PAMConfig  *pam.Config
}

// PAM Vault Safe list of Hashes with accountid and accountname (de-normalized table)
type SafeHash struct {
	gorm.Model
	Safename string
	Hash     string
	Name     string
}

type SendHashesStats struct {
	Status    int                    `json:"status"`
	SendCount int                    `json:"sendcount"`
	SafeStats *[]SafeHashStats       `json:"safestats,omitempty"`
	LeakStats *[]hmsl.SecretResponse `json:"leakstats,omitempty"`
}

type SafeHashStats struct {
	Safename      string    `json:"safename"`
	TotalCount    int       `json:"sentcount"`
	InvalidCount  int       `json:"invalidcount"`
	DistinctCount int       `json:"distinctcount"`
	Message       *[]string `json:"message,omitempty"`
}

type ErrorMessagesResponse struct {
	Error
	Errors []Error `json:"errors"`
}

// InitializeDb calls auto-migrate to create tables, if needed
func (b Brimstone) InitializeDb() error {
	errAutoMigrate := b.Db.AutoMigrate(&SafeHash{})
	if errAutoMigrate != nil {
		return errAutoMigrate
	}
	return nil
}

// sendBrimstoneError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendBrimstoneError(ctx echo.Context, code int, message string) error {
	brimstoneErr := Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, brimstoneErr)
	return err
}

func (b Brimstone) SaveHashBatch(hashbatch HashBatch) error {
	db := b.Db

	var hashes []SafeHash
	result := db.Limit(1).Where(&SafeHash{Safename: hashbatch.Safename}).Find(&hashes)
	if result.RowsAffected != 0 {
		return b.SaveExistingSafeHashes(hashbatch)
	}

	// new safe with new hashes
	var newsafehashes []SafeHash
	for i := 0; i < len(hashbatch.Hashes); i++ {
		newhash := SafeHash{
			Safename: hashbatch.Safename,
			Name:     hashbatch.Hashes[i].Name,
			Hash:     hashbatch.Hashes[i].Hash,
		}
		newsafehashes = append(newsafehashes, newhash)
	}
	result = db.CreateInBatches(newsafehashes, 100)

	return result.Error
}

// HashesPut - PUT /v1/hashes
func (b Brimstone) HashesPut(ctx echo.Context) error {

	var hashbatch HashBatch
	err := ctx.Bind(&hashbatch)
	if err != nil {
		return sendBrimstoneError(ctx, http.StatusBadRequest, "Invalid format for HashBatch")
	}

	return b.SaveHashBatch(hashbatch)
}

// SaveExistingSafeHashes - Safe already exists, save versions of hashes
func (b Brimstone) SaveExistingSafeHashes(batch HashBatch) error {
	db := b.Db

	// create a lookup dictionary from the name/hashes in the safe
	var hashes []SafeHash
	db.Where(&SafeHash{Safename: batch.Safename}).Find(&hashes)
	lookup := make(map[string]string)
	for i := 0; i < len(hashes); i++ {
		lookup[hashes[i].Name] = hashes[i].Hash
	}

	var newhashes []SafeHash
	var existinghashes []SafeHash
	for i := 0; i < len(batch.Hashes); i++ {
		h := SafeHash{
			Safename: batch.Safename,
			Name:     batch.Hashes[i].Name,
			Hash:     batch.Hashes[i].Hash,
		}

		// split out new and existing objects;
		// an object exists in the db if it's name is already associated with the safe
		if hash, ok := lookup[batch.Hashes[i].Name]; ok {
			// ignore if batch request hash value is same as db hash value
			if batch.Hashes[i].Hash != hash {
				existinghashes = append(existinghashes, h)
			}
		} else {
			newhashes = append(newhashes, h)
		}
	}

	// new name in the safe, so, just need to create them
	if len(newhashes) > 0 {
		db.CreateInBatches(newhashes, 100)
	}

	if len(existinghashes) > 0 {
		for i := 0; i < len(existinghashes); i++ {
			h := existinghashes[i]
			_ = db.Create(&h)

			// clean out extra hash records since we keep only MAX_HASH_COUNT records per safename-name

			var found []SafeHash
			result := db.Where(&SafeHash{Safename: h.Safename, Name: h.Name}).Find(&found)

			if result.RowsAffected > MAX_HASH_COUNT {
				// newest records first
				sort.Slice(found, func(i, j int) bool {
					return found[i].CreatedAt.After(found[j].CreatedAt)
				})

				// remove first 3 items so we don't delete them from the db
				found = slices.Delete(found, 0, 3)
				db.Unscoped().Delete(found)
			}
		}
	}

	// err := ctx.JSON()

	return nil
}

// SendHashPrefixesGet - GET /v1/hashes/sendprefixes
func (b Brimstone) SendHashPrefixesGet(ctx echo.Context) error {
	db := b.Db
	hmslclient := b.HMSLClient
	contextctx := context.TODO()

	// Send batches of hashes grouped by safename
	var safenames []SafeHash
	db.Distinct("safename").Order("safename").Find(&safenames)

	// start gathering the response stats
	stats := SendHashesStats{
		Status:    0,
		SendCount: 0,
		SafeStats: &[]SafeHashStats{},
	}
	batchsize := 5

	// send batches of hashes to HMSL
	// TODO: (low priority) use context.Context to store stats
	for i := 0; i < len(safenames); i++ {

		prefixes, err := b.FetchPrefixes(safenames[i].Safename, &stats)
		if err != nil {
			return err
		}

		// if we have valid prefixes, then send them to HMSL
		if stats.SendCount > 0 {
			var batch []string
			for i := 0; i < len((*prefixes)); i += batchsize {
				batch = (*prefixes)[i:utils.MinInt(i+batchsize, len((*prefixes)))]
				fmt.Println(batch)
				prefixquery := hmsl.PrefixesQuery{Prefixes: &batch}
				respHashes, respHashesErr := hmslclient.BatchPrefixesV1PrefixesPostWithResponse(contextctx, prefixquery)
				if respHashesErr != nil {
					return respHashesErr
				}

				HandlePrefixes(batch, respHashes)
			}
		}
	}

	err := ctx.JSON(200, stats)
	return err
}

// SendFullHashesGet - GET /v1/hashes/sendhashes
func (b Brimstone) SendFullHashesGet(ctx echo.Context) error {
	db := b.Db
	hmslclient := b.HMSLClient
	contextctx := context.TODO()

	// Send batches of hashes grouped by safename
	var safenames []SafeHash
	db.Distinct("safename").Order("safename").Find(&safenames)

	// start gathering the response stats
	stats := SendHashesStats{
		Status:    0,
		SendCount: 0,
		SafeStats: &[]SafeHashStats{},
		LeakStats: &[]hmsl.SecretResponse{},
	}
	batchsize := 1000

	// send batches of hashes to HMSL
	// TODO: (low priority) use context.Context to store stats
	var hashesleaked []hmsl.SecretResponse
	var hashesleakedvalidationerrors []error
	for i := 0; i < len(safenames); i++ {

		hashvals, err := b.FetchHashes(safenames[i].Safename, &stats)
		if err != nil {
			return err
		}

		// send hashes batch size limit to 1000; hmsl restriction.
		if stats.SendCount > 0 {
			var batch []string
			for i := 0; i < len((*hashvals)); i += batchsize {
				hashbody := hmsl.BatchHashesV1HashesPostJSONRequestBody{}
				batch = (*hashvals)[i:utils.MinInt(i+batchsize, len((*hashvals)))]
				fmt.Println(batch)

				hashbody.Hashes = &batch
				respHashes, respHashesErr := hmslclient.BatchHashesV1HashesPostWithResponse(contextctx, hashbody)
				if respHashesErr != nil {
					return respHashesErr
				}

				secretresponses, validationerrs := HandleHasheBatchResponses(batch, respHashes)
				if len(validationerrs) > 0 {
					hashesleakedvalidationerrors = append(hashesleakedvalidationerrors, validationerrs...)
				}
				hashesleaked = append(hashesleaked, secretresponses...)
			}
		}
	}

	for i := 0; i < len(hashesleaked); i++ {
		err := b.ChangePasswordFromHash(ctx, hashesleaked[i].Hash)
		msg := "succeded"
		if err != nil {
			msg = fmt.Sprintf("failed with error: %s", err.Error())
		}
		if hashesleaked[i].Location == nil {
			hashesleaked[i].Location = &hmsl.APILocation{
				U: "Location=UNK",
			}
		}
		hashesleaked[i].Location.U = fmt.Sprintf("Password request %s; Location=%s", msg, hashesleaked[i].Location.U)
	}

	var rsp = []struct {
		Responses        []hmsl.SecretResponse
		ValidationErrors []error
	}{
		{
			Responses:        hashesleaked,
			ValidationErrors: hashesleakedvalidationerrors,
		},
	}
	err := ctx.JSON(200, rsp)
	return err
}

// GitGuardianEventPost - POST /v1/notify/ggevent
func (b Brimstone) GitGuardianEventPost(ctx echo.Context) error {
	// If we get here, then the GG header has already been validated
	pamconfig := b.PAMConfig
	client := pam.NewClient(pamconfig.PCloudURL, *pamconfig)

	var event gg.IncidentEvent
	err := ctx.Bind(&event)
	if err != nil {
		return sendBrimstoneError(ctx, http.StatusBadRequest, "Invalid format for Incident Event")
	}

	if event.Action != "incident_triggered" && event.Action != "new_occurrence" {
		return sendBrimstoneError(ctx, http.StatusBadRequest, "Incident already recorded")
	}

	// hmsl_hash only exists on the Incident
	if event.Incident.HmslHash == nil {
		return sendBrimstoneError(ctx, http.StatusNotFound, "HMSL hash not sent as a parameter")
	}

	accountids, err := b.FindAccountID(*event.Incident.HmslHash)
	if err != nil {
		log.Printf("Error finding hmsl hash: %s\n", err.Error())
		return sendBrimstoneError(ctx, http.StatusNotFound, "No matching hmsl hash")
	}
	err = client.RefreshSessionToken()
	if err != nil {
		log.Printf("Error refreshing PAM session token: %s\n", err.Error())
		return sendBrimstoneError(ctx, http.StatusBadGateway, "Unable to obtain PAM session token")
	}

	// Found a matching HMSL Hash, so, let's tell PAM to change the password
	if len(accountids) > 0 {
		var resp ErrorMessagesResponse
		for i := 0; i < len(accountids); i++ {
			log.Printf("Account ID: %s\n", accountids[i])
			code, err := client.ChangePasswordImmediately(accountids[i])
			if err != nil {
				newmsg := fmt.Sprintf("ERROR: failed to change password for acct id, %s: %s\n", accountids[i], err.Error())
				brimstoneErr := Error{
					Code:    int32(code),
					Message: newmsg,
				}
				resp.Errors = append(resp.Errors, brimstoneErr)
			}
		}
		code := http.StatusOK
		if len(resp.Errors) > 0 {
			code = http.StatusConflict
		}
		return ctx.JSON(code, resp)
		// return sendBrimstoneError(ctx, http.StatusConflict, "Unable to change PAM account password")

	} else {
		// NO matching HMSL Hash, so, let's add a new account to PAM
		acctname := "gitguardian"
		if event.Incident.GitguardianUrl != nil {
			u, err := url.Parse(*event.Incident.GitguardianUrl)
			if err == nil {
				acctname = strings.ReplaceAll(u.Path, "/", "")
			}
		}
		addreq := pam.PostAddAccountRequest{
			Name:                      acctname,
			Address:                   *event.Incident.GitguardianUrl, // "https://dashboard.gitguardian.com/workspace/00000/incident/00000",
			UserName:                  "gitguardian",
			SafeName:                  pamconfig.SafeName,
			PlatformID:                pamconfig.PlatformID, // PCloud
			Secret:                    *event.Incident.SecretHash,
			SecretType:                "password",
			PlatformAccountProperties: pam.PlatformAccountProperties{},
		}
		newaccount, code, err := client.AddAccount(addreq)
		if err != nil || newaccount.ID == "" {
			return sendBrimstoneError(ctx, code, "Unable to add PAM account from GG incident")
		}

		db := b.Db
		var newsafehashes []SafeHash
		newhash := SafeHash{
			Safename: newaccount.SafeName,
			Name:     newaccount.ID,
			Hash:     *event.Incident.HmslHash,
		}
		newsafehashes = append(newsafehashes, newhash)
		result := db.CreateInBatches(newsafehashes, 1)
		if result != nil && result.Error != nil {
			log.Printf("unable to save new account hmsl hash (%s, %s, %s): %s\n", newhash.Safename, newhash.Name, newhash.Hash, result.Error.Error())
		}

		msg := fmt.Sprintf("New account created, %s, for %s", newaccount.ID, *event.Incident.GitguardianUrl)
		return sendBrimstoneError(ctx, http.StatusOK, msg)
	}
}

// CyberArkPAMCPMEventPut receive CPM plugin request; CPM updated the password, this request is telling brimstone to update its database
func (b Brimstone) CyberArkPAMCPMEventPut(ctx echo.Context) error {
	db := b.Db
	pamconfig := b.PAMConfig

	var event HashBatch
	err := ctx.Bind(&event)
	if err != nil {
		return sendBrimstoneError(ctx, http.StatusBadRequest, "Invalid format for CPM Event")
	}

	if event.Safename == "" || len(event.Hashes) == 0 {
		return fmt.Errorf("no hashes passed in")
	}

	// CPM will usually send account name (not account id), so, we attempt to determine accountid by querying PAM
	client := pam.NewClient(pamconfig.PCloudURL, *pamconfig)
	err = client.RefreshSessionToken()
	if err != nil {
		log.Printf("Error refreshing PAM session token: %s\n", err.Error())
		return sendBrimstoneError(ctx, http.StatusBadGateway, "Unable to obtain PAM session token")
	}

	// Loookup acount id based on account name
	accountid, rescode, errFetchId := client.FetchAccountIdFromAccountName(event.Safename, event.Hashes[0].Name)
	if rescode == 404 {
		log.Printf("unable to determine accountid, safename: %s, account name: %s\n", event.Safename, event.Hashes[0].Name)
		return sendBrimstoneError(ctx, http.StatusNotFound, fmt.Sprintf("unable to determine accountid, safename: %s, account name: %s", event.Safename, event.Hashes[0].Name))
	}
	if rescode > 299 && errFetchId != nil {
		log.Printf("error with result code, %d: %s\n", rescode, errFetchId.Error())
		return sendBrimstoneError(ctx, rescode, fmt.Sprintf("error with result code, %d: %s", rescode, errFetchId.Error()))
	}
	// Adjust the event obj to use account id instead of name
	event.Hashes[0].Name = accountid
	log.Printf("CPM Event found Account ID: %s\n", accountid)

	var hashes []SafeHash
	result := db.Limit(1).Where("safename = ? AND name = ?", event.Safename, event.Hashes[0].Name).Find(&hashes)
	if result.RowsAffected != 0 {
		log.Printf("saving next version of hash, safename: %s, account id: %s, hash: %s\n", event.Safename, event.Hashes[0].Name, event.Hashes[0].Hash)
		return b.SaveExistingSafeHashes(event)
	}

	// safe with new hash (CPM will only send 1 hash when an account password is reset)
	var newsafehashes []SafeHash
	newhash := SafeHash{
		Safename: event.Safename,
		Name:     event.Hashes[0].Name,
		Hash:     event.Hashes[0].Hash,
	}
	newsafehashes = append(newsafehashes, newhash)
	log.Printf("saving new version of hash, safename: %s, account id: %s, hash: %s\n", event.Safename, event.Hashes[0].Name, event.Hashes[0].Hash)
	result = db.CreateInBatches(newsafehashes, 10)

	return result.Error
}

// FindAccountID - given hmslhash return list of accountids from the db
func (b Brimstone) FindAccountID(hmslhash string) ([]string, error) {
	db := b.Db
	var hashes []SafeHash
	var ids []string
	result := db.Where(&SafeHash{Hash: hmslhash}).Find(&hashes)
	if result.RowsAffected != 0 {
		for i := 0; i < len(hashes); i++ {
			ids = append(ids, hashes[i].Name)
		}
	}
	return ids, nil
}

// FetchHashes - given safename return list of hmsl hashes from the db associated to that safe
func (b Brimstone) FetchHashes(safename string, stats *SendHashesStats) (*[]string, error) {
	db := b.Db

	safestats := SafeHashStats{
		Safename:      safename,
		TotalCount:    0,
		InvalidCount:  0,
		DistinctCount: 0,
	}

	var hashes []SafeHash
	var hashvals []string
	result := db.Where(&SafeHash{Safename: safename}).Find(&hashes)
	if result.RowsAffected != 0 {
		for i := 0; i < len(hashes); i++ {
			hashvals = append(hashvals, hashes[i].Hash)
		}
	}
	safestats.TotalCount = len(hashvals)

	// remove duplicates
	slices.Sort(hashvals)
	hashvals = slices.Compact(hashvals)
	safestats.DistinctCount = len(hashvals)
	stats.SendCount = len(hashvals) // number of valid prefixes to send

	return &hashvals, nil
}

// FetchPrefixes - given safename return list of hmsl hash prefixes from db
func (b Brimstone) FetchPrefixes(safename string, stats *SendHashesStats) (*[]string, error) {
	db := b.Db

	safestats := SafeHashStats{
		Safename:      safename,
		TotalCount:    0,
		InvalidCount:  0,
		DistinctCount: 0,
	}

	var hashes []SafeHash
	var prefixes []string
	result := db.Where(&SafeHash{Safename: safename}).Find(&hashes)
	if result.RowsAffected != 0 {
		for i := 0; i < len(hashes); i++ {
			prefixes = append(prefixes, hashes[i].Hash[0:5])
		}
	}
	safestats.TotalCount = len(prefixes)

	// remove duplicates
	slices.Sort(prefixes)
	prefixes = slices.Compact(prefixes)
	safestats.DistinctCount = len(prefixes)

	//
	var validprefixes []string
	var messages []string
	for i := 0; i < len(prefixes); i++ {
		str := strings.ToLower(prefixes[i])
		// {"pattern":"^[a-f0-9]{5}$"}}]}
		isMatch := regexp.MustCompile(`^[a-f0-9]{5}$`).MatchString(str)
		if !isMatch {
			msg := fmt.Sprintf("hash for safe, %s, is not valid: %s.  Skipping.", safename, prefixes[i])
			messages = append(messages, msg)
			continue
		}
		validprefixes = append(validprefixes, prefixes[i])
	}
	if len(messages) > 0 {
		safestats.Message = &messages
	}
	safestats.InvalidCount = len(prefixes) - len(validprefixes)
	x := append(*stats.SafeStats, safestats)
	stats.SafeStats = &x
	stats.SendCount = len(prefixes) // number of valid prefixes to send

	return &validprefixes, nil
}

// ChangePasswordFromHash - given hmslhash lookup accountid and call to pam api to change password
func (b Brimstone) ChangePasswordFromHash(ctx echo.Context, hmslhash string) error {
	pamconfig := b.PAMConfig
	client := pam.NewClient(pamconfig.PCloudURL, *pamconfig)

	clientErr := client.RefreshSessionToken()
	if clientErr != nil {
		return fmt.Errorf("error refreshing PAM session token: %s", clientErr.Error())
	}

	accountids, err := b.FindAccountID(hmslhash)
	if len(accountids) > 0 {
		for i := 0; i < len(accountids); i++ {
			log.Printf("Account ID: %s\n", accountids[i])
			code, err := client.ChangePasswordImmediately(accountids[i])
			if err != nil {
				return fmt.Errorf("failed to change password for acct id, %s: (code=%d) %s", accountids[i], code, err.Error())
			}
		}
	}
	if err != nil {
		return fmt.Errorf("error finding hmsl hash: %s", err.Error())
	}

	return nil
}

// HandlePrefixes - TODO: (low priority) add vault remediation logic for prefix responses
func HandlePrefixes(hashes []string, responses *hmsl.BatchPrefixesV1PrefixesPostResponse) *hmsl.PrefixesResponse {
	var result hmsl.PrefixesResponse
	hints := make(map[string]string)

	for i := 0; i < len(hashes); i++ {
		h, e := hmsl.ComputeHint(hashes[i])
		if e == nil {
			hints[h] = hashes[i]
			continue
		}
	}

	// foreach of the batch items: find which ones have matches from the responses
	for i := 0; i < len(responses.JSON200.Matches); i++ {

		pbytes, pbErr := responses.JSON200.Matches[i].Payload.Bytes()
		if pbErr != nil {
			log.Printf("Error converting to bytes: %s\n", pbErr.Error())
		}

		msg1, err1 := hmsl.DecryptPayload(pbytes, hints[responses.JSON200.Matches[i].Hint])
		if err1 != nil {
			log.Printf("Error decrypt payload failed with hmslhash: %s\n", err1.Error())
		}

		log.Printf("HMSLHASH: %s\n", msg1)
	}

	return &result
}

// HandleHasheBatchResponses
func HandleHasheBatchResponses(hashes []string, responses *hmsl.BatchHashesV1HashesPostResponse) ([]hmsl.SecretResponse, []error) {
	var hashesleaked []hmsl.SecretResponse
	var validationerrors []error

	if responses == nil {
		e := fmt.Errorf("no responses to process")
		validationerrors = append(validationerrors, e)
		return hashesleaked, validationerrors
	}
	if responses.JSON200 != nil {
		for i := 0; i < len(responses.JSON200.Secrets); i++ {
			secret := responses.JSON200.Secrets[i]
			loc := "UNK"
			if secret.Location != nil {
				loc = secret.Location.U
			}
			log.Printf("INFO: Secret Result, count=%d, hash=%s, location=%s\n", secret.Count, secret.Hash, loc)
			hashesleaked = append(hashesleaked, secret)
		}
	}
	if responses.JSON422 != nil && responses.JSON422.Detail != nil {
		for i := 0; i < len(*responses.JSON422.Detail); i++ {
			detail := (*responses.JSON422.Detail)[i]
			m := fmt.Errorf("422 validation error (%s) %s", detail.Type, detail.Msg)
			validationerrors = append(validationerrors, m)
		}
	}
	return hashesleaked, validationerrors
}
