package hailstone

import (
	"errors"
	"testing"

	hmsl "github.com/davidh-cyberark/brimstone/pkg/hasmysecretleaked"
	pam "github.com/davidh-cyberark/brimstone/pkg/privilegeaccessmanager"

	"github.com/stretchr/testify/assert"
)

type MockPamClient struct{}

func (t *MockPamClient) RefreshSessionToken() error {
	return nil
}

func (t *MockPamClient) FetchAccounts() ([]pam.Account, error) {
	return []pam.Account{
		{
			ID:       "acct1",
			Name:     "acct1",
			SafeName: "safe1",
		},
		{
			ID:       "acct2",
			Name:     "acct2",
			SafeName: "safe1",
		},

		{
			ID:       "acct3",
			Name:     "acct3",
			SafeName: "safe2",
		},
	}, nil
}

func (t *MockPamClient) FetchAccountPassword(accountID string) (string, error) {
	switch accountID {
	case "acct1":
		return "password1", nil
	case "acct2":
		return "password2", nil
	case "acct3":
		return "password3", nil
	}
	return "", errors.New("account not found")
}

func TestLoadAllAccounts(t *testing.T) {
	pamClient := MockPamClient{}

	hash1, _ := hmsl.ComputeHash("password1")
	hash2, _ := hmsl.ComputeHash("password2")
	hash3, _ := hmsl.ComputeHash("password3")

	batches, err := LoadAllAccounts(&pamClient)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(batches))

	assert.Equal(t, "safe1", batches["safe1"].Safename)
	assert.Equal(t, 2, len(batches["safe1"].Hashes))
	assert.Equal(t, hash1, batches["safe1"].Hashes[0].Hash)
	assert.Equal(t, hash2, batches["safe1"].Hashes[1].Hash)

	assert.Equal(t, "safe2", batches["safe2"].Safename)
	assert.Equal(t, 1, len(batches["safe2"].Hashes))
	assert.Equal(t, hash3, batches["safe2"].Hashes[0].Hash)
}
