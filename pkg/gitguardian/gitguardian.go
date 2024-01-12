package gitguardian

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type IncidentEvent struct {
	Source     string    `json:"source,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	Action     string    `json:"action,omitempty"`
	Message    string    `json:"message,omitempty"`
	TargetUser string    `json:"target_user,omitempty"`
	Incident   Incident  `json:"incident,omitempty"`
	Occurrence VcsOccurrence `json:"occurrence,omitempty"`
}

// Python impl: <https://docs.gitguardian.com/platform/monitor-perimeter/notifiers-integrations/custom-webhook#how-to-verify-the-payload-signature>
func ValidateGGPayload(ggheader string, timestamp string, signature_token string, msg []byte) bool {

	if !strings.HasPrefix(ggheader, "sha256=") {
		return false
	}

	// Using the GG header hash
	hash := strings.TrimPrefix(ggheader, "sha256=")
	sig, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	// HMAC sign and sum the post body
	key := []byte(fmt.Sprintf("%s%s", timestamp, signature_token))
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	macsum := mac.Sum(nil)

	// If postbody sum matches the header hash, then this is a valid post
	return hmac.Equal(sig, macsum)
}
