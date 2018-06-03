package s3

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	// Identifies signature version 2.
	v2Identifier = "AWS"
	V2           = 2

	// Identifies signature version 4.
	v4Identifier = "AWS4-HMAC-SHA256"
	V4           = 4
)

// ValidateSignVersion checks the request provides valid sign version.
func ValidateSignVersion(authString string) ErrorCode {
	// Check signature version.
	// Only v4 is allowed.
	if signatureV4(authString) == false {
		return ErrCredentialsNotSupported
	}
	return ErrNone
}

func SignatureVersion(authString string) int {
	if strings.HasPrefix(authString, v4Identifier) {
		return V4 // Signature V4
	} else if strings.HasPrefix(authString, v2Identifier) {
		return V2 // Signature V2
	}

	return -1 // Invalid signature identifiers.
}

func signatureV4(authString string) bool {
	if SignatureVersion(authString) == V4 {
		return true
	}
	return false
}

// SignV4 contains the fields which are required to make v4 sign.
// https://docs.aws.amazon.com/ko_kr/general/latest/gr/signature-v4-examples.html
type SignV4 struct {
	Credential    CredV4
	SignedHeaders []string
	Signature     string
}

// CredV4 is the parsed information of SignV4.Credential.
type CredV4 struct {
	AccessKey   string
	Date        string
	Region      string
	Service     string
	Termination string // fixed: aws4_request
}

type SignV2 struct {
	Credential CredV2
	Signature  string
}

type CredV2 struct {
	AccessKey string
}

// Scope returns credential scope of signature v4.
func (c *CredV4) Scope() string {
	return strings.Join([]string{
		c.Date,
		c.Region,
		c.Service,
		c.Termination,
	}, "/")
}

func parseCredV4(credString string) (c CredV4, err ErrorCode) {
	// credString = accessKey + "/" + date + "/" + region + "/" service + "/" + "aws4_request"
	args := strings.Split(credString, "/")
	if len(args) != 5 {
		return c, ErrInvalidSignatureFormat
	}

	// Check the date format is right.
	if _, err := time.Parse("20060102", args[1]); err != nil {
		return c, ErrInvalidSignatureFormat
	}

	// Check is service name valid.
	if args[3] != "s3" {
		return c, ErrInvalidSignatureFormat
	}

	// Check termination string.
	if args[4] != "aws4_request" {
		return c, ErrInvalidSignatureFormat
	}

	return CredV4{
		AccessKey:   args[0],
		Date:        args[1],
		Region:      args[2],
		Service:     args[3],
		Termination: args[4],
	}, ErrNone
}

func parseSignedHeaderV4(headerString string) []string {
	// headerString = Lowercase(HeaderName0) + ";" Lowercase(HeaderName1) + ";" + ... + Lowercase(HeaderNameN)
	return strings.Split(headerString, ";")
}

// ParseSignV4 parses the v4 authorization string.
func ParseSignV4(authString string) (*SignV4, ErrorCode) {
	// Remove sign algorithm identifier.
	authString = strings.TrimPrefix(authString, v4Identifier)
	fields := strings.Split(strings.TrimSpace(authString), ",")
	if len(fields) < 3 {
		return nil, ErrMissingSecurityHeader
	}

	for i := range fields {
		f := strings.Split(fields[i], "=")
		if len(f) < 2 {
			return nil, ErrMissingSecurityHeader
		}

		fields[i] = f[1]
	}

	cred, err := parseCredV4(fields[0])
	if err != ErrNone {
		return nil, err
	}

	return &SignV4{
		Credential:    cred,
		SignedHeaders: parseSignedHeaderV4(fields[1]),
		Signature:     fields[2],
	}, ErrNone
}

func ParseSignV2(authString string) (*SignV2, ErrorCode) {
	authString = strings.TrimPrefix(authString, v2Identifier)
	authString = strings.TrimSuffix(authString, "=")
	fields := strings.Split(strings.TrimSpace(authString), ":")
	if len(fields) < 2 {
		return nil, ErrMissingSecurityHeader
	}

	return &SignV2{
		Credential: CredV2{
			AccessKey: fields[0],
		},
		Signature: fields[1],
	}, ErrNone
}

// GenCanonicalHeaders generates canonical headers from the http request.
func GenCanonicalHeaders(r *http.Request, signedHeaders []string) string {
	sort.Strings(signedHeaders)

	var canonicalHeader string
	for _, signedHeader := range signedHeaders {
		v := r.Header.Get(http.CanonicalHeaderKey(signedHeader))
		// Handling empty value headers.
		if v == "" {
			v = _GetCanonicalHeaders(r, signedHeader)
		}

		canonicalHeaderEntry := strings.ToLower(signedHeader) + ":" + strings.TrimSpace(v) + "\n"
		canonicalHeader += canonicalHeaderEntry
	}

	return canonicalHeader
}

func _GetCanonicalHeaders(r *http.Request, signedHeader string) string {
	switch signedHeader {
	case "host":
		return r.Host
	default:
		return ""
	}
}

// GenSignedHeadersString generates a string with signed headers.
func GenSignedHeadersString(signedHeaders []string) string {
	sort.Strings(signedHeaders)
	for i := range signedHeaders {
		signedHeaders[i] = strings.ToLower(signedHeaders[i])
	}
	return strings.Join(signedHeaders, ";")
}

// GenCanonicalRequest generates canonical request.
// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-canonical-request.html
func GenCanonicalRequest(httpRequestMethod, uri, queryString, headers, signedHeaders, payload string) string {
	return httpRequestMethod + "\n" +
		uri + "\n" +
		queryString + "\n" +
		headers + "\n" +
		signedHeaders + "\n" +
		payload
}

// GenStringToSign generates string to sign.
// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-string-to-sign.html
func GenStringToSign(algorithm, requestDateTime, credentialScope, hashedCanonicalRequest string) string {
	return algorithm + "\n" +
		requestDateTime + "\n" +
		credentialScope + "\n" +
		hashedCanonicalRequest
}

// GenSignatureKey generates signature key.
// https://docs.aws.amazon.com/ko_kr/general/latest/gr/signature-v4-examples.html
func GenSignatureKey(secretKey, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(regionName))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))

	return kSigning
}

// GenSignature generates signature.
func GenSignature(signatureKey []byte, stringToSign string) string {
	return hex.EncodeToString(hmacSHA256(signatureKey, []byte(stringToSign)))
}

func hmacSHA256(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}
