package hasmysecretleaked

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

// From the docs <https://api.hasmysecretleaked.com/docs>

// Compute HMSL hash key
func ComputeHash(val string) (string, error) {
	pepper := sha256.Sum256([]byte("GitGuardian"))
	dk, err := scrypt.Key([]byte(val), []byte(pepper[:]), 2048, 8, 1, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(dk[:]), nil
}

func ComputeHint(val string) (string, error) {
	/* def make_hint(key_as_hex: str) -> str:
	 * return sha256(bytes.fromhex(key_as_hex)).hexdigest()
	 */
	data, err := hex.DecodeString(val)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", []byte(sum[:])), nil
}

/*
	python func from the docs:

def decrypt_payload(payload_str: str, key: str) -> str:

	"""Decrypt the payload with the given key."""
	payload = base64.b64decode(payload_str)
	cleartext = AESGCM(bytes.fromhex(key))
		.decrypt(nonce=payload[:12], data=payload[12:], associated_data=None)
	return cleartext.decode()
*/
func DecryptPayload(p []byte, k string) (string, error) {
	key, _ := hex.DecodeString(k)
	// ciphertext, err := base64.StdEncoding.DecodeString(p[12:])
	ciphertext := p[12:]

	// nonce, _ := hex.DecodeString(p[:12])
	nonce := p[:12]
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
