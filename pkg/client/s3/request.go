package s3

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/chanyoung/nil/pkg/client"
	s3lib "github.com/chanyoung/nil/pkg/s3"
	"github.com/chanyoung/nil/pkg/security"
)

type s3request struct {
	request   *http.Request
	transport *http.Transport
	client    *http.Client
}

func (r *s3request) Send() (*http.Response, error) {
	return r.client.Do(r.request)
}

// NewS3Request creates a new s3 request.
func NewS3Request(request *http.Request, genSign bool, cred, copyHeader map[string]string) (client.Request, error) {
	if genSign {
		request.Header.Set("X-Amz-Date", time.Now().UTC().Format(time.RFC3339))
		request.Header.Set("X-Amz-Content-Sha256", cred["content-hash"])
		signedHeaders := []string{"host", "X-Amz-Content-Sha256"}

		// Gen canonical request and set to request header.
		canonicalRequest := s3lib.GenCanonicalRequest(
			request.Method,
			request.RequestURI,
			request.URL.Query().Encode(),
			s3lib.GenCanonicalHeaders(request, signedHeaders),
			s3lib.GenSignedHeadersString(signedHeaders),
			cred["content-hash"],
		)

		// Gen string to sign.
		sha256CanonicalRequest := sha256.Sum256([]byte(canonicalRequest))
		credV4 := s3lib.CredV4{
			AccessKey:   cred["access-key"],
			Date:        time.Now().UTC().Format("20060102"),
			Region:      cred["region"],
			Service:     "s3",
			Termination: "aws4_request",
		}
		stringToSign := s3lib.GenStringToSign(
			"AWS4-HMAC-SHA256",
			request.Header.Get("X-Amz-Date"),
			credV4.Scope(),
			hex.EncodeToString(sha256CanonicalRequest[:]),
		)

		// Derive signature.
		signatureKey := s3lib.GenSignatureKey(
			cred["secret-key"],
			credV4.Date,
			credV4.Region,
			credV4.Service,
		)
		signature := s3lib.GenSignature(signatureKey, stringToSign)
		authString := "AWS4-HMAC-SHA256 " +
			"Credential=" + credV4.AccessKey + "/" + credV4.Scope() + "," +
			"SignedHeaders=" + strings.Join(signedHeaders, ";") + "," +
			"Signature=" + signature
		request.Header.Set("Authorization", authString)
	} else {
		for key, value := range copyHeader {
			request.Header.Set(key, value)
		}
	}

	// Create http transport.
	transport := &http.Transport{
		Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
		TLSClientConfig:     security.DefaultTLSConfig(),
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &s3request{
		request:   request,
		transport: transport,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}, nil
}
