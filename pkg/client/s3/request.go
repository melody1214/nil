package s3

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/chanyoung/nil/pkg/client"
	s3lib "github.com/chanyoung/nil/pkg/s3"
)

type S3RequestEvent struct {
	protocol client.Protocol

	httpWriter  http.ResponseWriter
	httpRequest *http.Request

	authArgs *s3lib.SignV4
}

func NewS3RequestEvent(w http.ResponseWriter, r *http.Request) (client.RequestEvent, error) {
	authArgs, s3err := getAuthArgsV4(r.Header.Get("Authorization"))
	if s3err != s3lib.ErrNone {
		return nil, client.ErrInvalidProtocol
	}

	return &S3RequestEvent{
		protocol:    client.S3,
		httpWriter:  w,
		httpRequest: r,
		authArgs:    authArgs,
	}, nil
}

func (r *S3RequestEvent) Protocol() client.Protocol {
	return r.protocol
}

func (r *S3RequestEvent) ResponseWriter() http.ResponseWriter {
	return r.httpWriter
}

func (r *S3RequestEvent) Request() *http.Request {
	return r.httpRequest
}

func (r *S3RequestEvent) AccessKey() string {
	return r.authArgs.Credential.AccessKey
}

func (r *S3RequestEvent) Auth(secretKey string) bool {
	// Task 1: Create a Canonical Request for Signature Version 4.
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-canonical-request.html
	canonicalRequest := s3lib.GenCanonicalRequest(
		r.httpRequest.Method,
		r.httpRequest.RequestURI,
		r.httpRequest.URL.Query().Encode(),
		s3lib.GenCanonicalHeaders(r.httpRequest, r.authArgs.SignedHeaders),
		s3lib.GenSignedHeadersString(r.authArgs.SignedHeaders),
		r.httpRequest.Header.Get("X-Amz-Content-Sha256"),
	)

	// Task 2: Create a String to Sign for Signature Version 4.
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-string-to-sign.html
	sha256CanonicalRequest := sha256.Sum256([]byte(canonicalRequest))
	stringToSign := s3lib.GenStringToSign(
		"AWS4-HMAC-SHA256",
		r.httpRequest.Header.Get("X-Amz-Date"),
		r.authArgs.Credential.Scope(),
		hex.EncodeToString(sha256CanonicalRequest[:]),
	)

	// Task 3: Calculate the Signature for AWS Signature Version 4
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-calculate-signature.html
	signatureKey := s3lib.GenSignatureKey(secretKey, r.authArgs.Credential.Date, r.authArgs.Credential.Region, r.authArgs.Credential.Service)

	derivedSignature := s3lib.GenSignature(signatureKey, stringToSign)
	if r.authArgs.Signature != derivedSignature {
		return false
	}

	return true
}

func getAuthArgsV4(authString string) (*s3lib.SignV4, s3lib.ErrorCode) {
	// Check the sign version is supported.
	if err := s3lib.ValidateSignVersion(authString); err != s3lib.ErrNone {
		return nil, err
	}

	// Parse auth string.
	return s3lib.ParseSignV4(authString)
}
