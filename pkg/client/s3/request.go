package s3

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/chanyoung/nil/pkg/client"
	s3lib "github.com/chanyoung/nil/pkg/s3"
)

// S3RequestEvent is used to handling s3 type of requests.
type S3RequestEvent struct {
	protocol client.Protocol

	httpWriter  http.ResponseWriter
	httpRequest *http.Request

	authArgs *s3lib.SignV4
}

// NewS3RequestEvent creates a new s3 request event.
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

// Protocol is a getter of protocol type.
func (r *S3RequestEvent) Protocol() client.Protocol {
	return r.protocol
}

// ResponseWriter is a getter of http response writer.
func (r *S3RequestEvent) ResponseWriter() http.ResponseWriter {
	return r.httpWriter
}

// Request is a getter of http request.
func (r *S3RequestEvent) Request() *http.Request {
	return r.httpRequest
}

// AccessKey is a getter of access key.
func (r *S3RequestEvent) AccessKey() string {
	return r.authArgs.Credential.AccessKey
}

// Region is a getter of region.
func (r *S3RequestEvent) Region() string {
	return r.authArgs.Credential.Region
}

// Bucket is a getter of bucket.
func (r *S3RequestEvent) Bucket() string {
	return strings.Trim(r.httpRequest.RequestURI, "/")
}

// Auth checks the given secret key is same with the encoded secret key in the http request.
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

// SendInternalError sends s3 internal error to the client.
func (r *S3RequestEvent) SendInternalError() {
	s3lib.SendError(r.httpWriter, s3lib.ErrInternalError, r.httpRequest.RequestURI, "")
}

// SendIncorrectKey sends s3 incorrect key error to the client.
func (r *S3RequestEvent) SendIncorrectKey() {
	// TODO: implement
	s3lib.SendError(r.httpWriter, s3lib.ErrInvalidAccessKeyId, r.httpRequest.RequestURI, "")
}

// SendNoSuchKey sends s3 no such key error to the client.
func (r *S3RequestEvent) SendNoSuchKey() {
	s3lib.SendError(r.httpWriter, s3lib.ErrInvalidAccessKeyId, r.httpRequest.RequestURI, "")
}

// SendInvalidURI sends s3 invalud uri error to the client.
func (r *S3RequestEvent) SendInvalidURI() {
	s3lib.SendError(r.httpWriter, s3lib.ErrInvalidURI, r.httpRequest.RequestURI, "")
}

func getAuthArgsV4(authString string) (*s3lib.SignV4, s3lib.ErrorCode) {
	// Check the sign version is supported.
	if err := s3lib.ValidateSignVersion(authString); err != s3lib.ErrNone {
		return nil, err
	}

	// Parse auth string.
	return s3lib.ParseSignV4(authString)
}
