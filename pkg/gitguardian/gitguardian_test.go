package gitguardian

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGGPayloadTrue(t *testing.T) {
	signature := "sha256=172fe3d694b734aa53dc892fd3b8d62163fc240064de570ba006900bb54a0fc2"
	timestamp := "0"
	signature_token := "foo"
	payload := []byte("bar")

	a := ValidateGGPayload(signature, timestamp, signature_token, payload)
	assert.True(t, a)
}
func TestValidateGGPayloadFalse(t *testing.T) {
	signature := "sha256=172fe3d694b734aa53dc892fd3b8d62163fc240064de570ba006900bb54a0fc2"
	timestamp := "1"
	signature_token := "foo"
	payload := []byte("bar")

	a := ValidateGGPayload(signature, timestamp, signature_token, payload)
	assert.False(t, a)
}
