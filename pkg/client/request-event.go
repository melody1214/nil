package client

import "net/http"

type RequestEvent interface {
	// Getter
	Protocol() Protocol
	ResponseWriter() http.ResponseWriter
	Request() *http.Request
	AccessKey() string
	Region() string
	Bucket() string
	Type() RequestType
	MD5() string

	// Method includes business logic.
	Auth(secretKey string) bool
	// CopyAuthHeader copy headers which is used to authenticate.
	CopyAuthHeader() map[string]string

	// Methods for handling errors.
	SendSuccess()
	SendInternalError()
	SendIncorrectKey()
	SendNoSuchKey()
	SendInvalidURI()
}
