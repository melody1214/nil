package s3

import (
	"strings"
)

const (
	// Identifies signature version 2.
	v2Identifier = "AWS"
	v2           = 2

	// Identifies signature version 4.
	v4Identifier = "AWS4-HMAC-SHA256"
	v4           = 4
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

func signatureVersion(authString string) int {
	if strings.HasPrefix(authString, v4Identifier) {
		return v4 // Signature V4
	} else if strings.HasPrefix(authString, v2Identifier) {
		return v2 // Signature V2
	}

	return -1 // Invalid signature identifiers.
}

func signatureV4(authString string) bool {
	if signatureVersion(authString) == v4 {
		return true
	}
	return false
}

// SignV4 contains the fields which are required to make v4 sign.
// https://docs.aws.amazon.com/ko_kr/general/latest/gr/signature-v4-examples.html
type SignV4 struct {
	Credential    string
	SignedHeaders string
	Signature     string
}

func (s *SignV4) GetAccessKey() string {
	args := strings.Split(s.Credential, "/")
	if len(args) < 4 {
		return ""
	}
	return args[0]
}

func (s *SignV4) GetDateStamp() string {
	args := strings.Split(s.Credential, "/")
	if len(args) < 4 {
		return ""
	}
	return args[1]
}

func (s *SignV4) GetRegion() string {
	args := strings.Split(s.Credential, "/")
	if len(args) < 4 {
		return ""
	}
	return args[2]
}

func (s *SignV4) GetService() string {
	args := strings.Split(s.Credential, "/")
	if len(args) < 4 {
		return ""
	}
	return args[3]
}

func (s *SignV4) Sign(key, dateStamp, region, service string) {
	keys := strings.Split(key, "/")

	s.Credential = keys[0] + "/" + dateStamp + "/" + region + "/" + service
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

	return &SignV4{
		Credential:    fields[0],
		SignedHeaders: fields[1],
		Signature:     fields[2],
	}, ErrNone
}
