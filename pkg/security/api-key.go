package security

import (
	"encoding/base64"
	"math/rand"
	"time"
)

const (
	accessKeyLength = 20
	secretKeyLength = 20

	letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// APIKey is the API key for user authenticating.
type APIKey struct {
	accessKey string
	secretKey string
}

// NewAPIKey creates new API key.
func NewAPIKey() APIKey {
	ak := APIKey{}

	// Generates new access key.
	b1 := make([]byte, accessKeyLength)
	for i := range b1 {
		b1[i] = letters[rand.Intn(len(letters))]
	}
	ak.accessKey = string(b1)

	// Generates secret key.
	b2 := make([]byte, secretKeyLength)
	for i := range b2 {
		b2[i] = letters[rand.Intn(len(letters))]
	}
	ak.secretKey = string([]byte(base64.StdEncoding.EncodeToString(b2))[:secretKeyLength])

	return ak
}

// AccessKey returns access key.
func (ak *APIKey) AccessKey() string {
	return ak.accessKey
}

// SecretKey returns secret key.
func (ak *APIKey) SecretKey() string {
	return ak.secretKey
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
