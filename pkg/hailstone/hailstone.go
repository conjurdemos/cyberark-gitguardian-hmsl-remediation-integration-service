package hailstone

import (
	"github.com/davidh-cyberark/brimstone/pkg/brimstone"
	hmsl "github.com/davidh-cyberark/brimstone/pkg/hasmysecretleaked"
	pam "github.com/davidh-cyberark/brimstone/pkg/privilegeaccessmanager"
)

type IPamClient interface {
	RefreshSessionToken() error
	FetchAccounts() ([]pam.Account, error)
	FetchAccountPassword(accountID string) (string, error)
}

func LoadAllAccounts(pamClient IPamClient) (map[string]brimstone.HashBatch, error) {
	err := pamClient.RefreshSessionToken()
	if err != nil {
		return nil, err
	}

	accounts, err := pamClient.FetchAccounts()
	if err != nil {
		return nil, err
	}

	batches := make(map[string]brimstone.HashBatch)
	for _, account := range accounts {
		password, err := pamClient.FetchAccountPassword(account.ID)
		if err != nil {
			return nil, err
		}

		hmslHash, err := hmsl.ComputeHash(password)
		if err != nil {
			return nil, err
		}
		newhash := brimstone.Hash{
			Hash: hmslHash,
			Name: account.ID,
		}

		batch, ok := batches[account.SafeName]
		if !ok {
			batch = brimstone.HashBatch{
				Safename: account.SafeName,
			}
		}
		batch.Hashes = append(batch.Hashes, newhash)
		batches[account.SafeName] = batch
	}

	return batches, nil
}
