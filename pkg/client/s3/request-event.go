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

	signVer    int
	authArgsV2 *s3lib.SignV2
	authArgsV4 *s3lib.SignV4
}

// NewS3RequestEvent creates a new s3 request event.
func NewS3RequestEvent(w http.ResponseWriter, r *http.Request) (client.RequestEvent, error) {
	e := &S3RequestEvent{
		protocol:    client.S3,
		httpWriter:  w,
		httpRequest: r,
	}

	authStr := r.Header.Get("Authorization")
	v := s3lib.SignatureVersion(authStr)
	switch v {
	case s3lib.V2:
		authArgs, s3err := s3lib.ParseSignV2(authStr)
		if s3err != s3lib.ErrNone {
			return nil, client.ErrInvalidProtocol
		}
		e.signVer = s3lib.V2
		e.authArgsV2 = authArgs

	case s3lib.V4:
		authArgs, s3err := s3lib.ParseSignV4(authStr)
		if s3err != s3lib.ErrNone {
			return nil, client.ErrInvalidProtocol
		}
		e.signVer = s3lib.V4
		e.authArgsV4 = authArgs

	default:
		return nil, client.ErrInvalidProtocol
	}

	return e, nil
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
	switch r.signVer {
	case s3lib.V2:
		return r.authArgsV2.Credential.AccessKey
	case s3lib.V4:
		return r.authArgsV4.Credential.AccessKey
	default:
		return ""
	}
}

// Region is a getter of region.
func (r *S3RequestEvent) Region() string {
	switch r.signVer {
	case s3lib.V2:
		return "KR"
	case s3lib.V4:
		return r.authArgsV4.Credential.Region
	default:
		return ""
	}
}

// Bucket is a getter of bucket.
func (r *S3RequestEvent) Bucket() string {
	return strings.Trim(r.httpRequest.RequestURI, "/")
}

// Auth checks the given secret key is same with the encoded secret key in the http request.
func (r *S3RequestEvent) Auth(secretKey string) bool {
	switch r.signVer {
	case s3lib.V2:
		return true
	case s3lib.V4:
		return r.authV4(secretKey)
	default:
		return false
	}
}

func (r *S3RequestEvent) authV4(secretKey string) bool {
	// Task 1: Create a Canonical Request for Signature Version 4.
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-canonical-request.html
	canonicalRequest := s3lib.GenCanonicalRequest(
		r.httpRequest.Method,
		r.httpRequest.RequestURI,
		r.httpRequest.URL.Query().Encode(),
		s3lib.GenCanonicalHeaders(r.httpRequest, r.authArgsV4.SignedHeaders),
		s3lib.GenSignedHeadersString(r.authArgsV4.SignedHeaders),
		r.httpRequest.Header.Get("X-Amz-Content-Sha256"),
	)

	// Task 2: Create a String to Sign for Signature Version 4.
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-string-to-sign.html
	sha256CanonicalRequest := sha256.Sum256([]byte(canonicalRequest))
	stringToSign := s3lib.GenStringToSign(
		"AWS4-HMAC-SHA256",
		r.httpRequest.Header.Get("X-Amz-Date"),
		r.authArgsV4.Credential.Scope(),
		hex.EncodeToString(sha256CanonicalRequest[:]),
	)

	// Task 3: Calculate the Signature for AWS Signature Version 4
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-calculate-signature.html
	signatureKey := s3lib.GenSignatureKey(secretKey, r.authArgsV4.Credential.Date, r.authArgsV4.Credential.Region, r.authArgsV4.Credential.Service)

	derivedSignature := s3lib.GenSignature(signatureKey, stringToSign)
	if r.authArgsV4.Signature != derivedSignature {
		return false
	}

	return true
}

// SendSuccess sends success message to the client.
func (r *S3RequestEvent) SendSuccess() {
	if r.requireMD5() {
		r.httpWriter.Header().Set("ETag", r.MD5())
	}
	s3lib.SendSuccess(r.httpWriter)
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

// CopyAuthHeader copy headers which is used to authenticate.
func (r *S3RequestEvent) CopyAuthHeader() map[string]string {
	header := make(map[string]string)
	header["Authorization"] = r.httpRequest.Header.Get("Authorization")

	switch r.signVer {
	case s3lib.V2:
		header["Amz-Sdk-Invocation-Id"] = r.httpRequest.Header.Get("Amz-Sdk-Invocation-Id")
	case s3lib.V4:
		for _, key := range r.authArgsV4.SignedHeaders {
			header[key] = r.httpRequest.Header.Get(key)
		}
	default:
	}

	return header
}

// Type get the type of the request.
func (r *S3RequestEvent) Type() client.RequestType {
	t := r.httpRequest.Header.Get("Request-Type")
	switch t {
	case client.WriteToPrimary.String():
		return client.WriteToPrimary
	case client.WriteToFollower.String():
		return client.WriteToFollower
	default:
		return client.UnknownType
	}
}

func (r *S3RequestEvent) requireMD5() bool {
	requestType := r.Type()
	if requestType != client.WriteToPrimary && requestType != client.WriteToFollower {
		return false
	}

	if isS3cmd(r.httpRequest) == false {
		return false
	}

	return true
}

// MD5 returns md5 string.
func (r *S3RequestEvent) MD5() string {
	if isS3cmd(r.httpRequest) {
		attrs := r.httpRequest.Header.Get("X-Amz-Meta-S3cmd-Attrs")
		for _, attr := range strings.Split(attrs, "/") {
			if strings.HasPrefix(attr, "md5:") {
				md5str := strings.Split(attr, ":")[1]
				return md5str
			}
		}
	}
	if md5 := r.httpRequest.Header.Get("Md5"); md5 != "" {
		return md5
	}
	return "01234567890123456789012345678901"
	// return r.httpRequest.Header.Get("Md5")
}

func isS3cmd(r *http.Request) bool {
	return r.Header.Get("X-Amz-Meta-S3cmd-Attrs") != ""
}
