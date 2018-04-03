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

	// Method includes business logic.
	Auth(secretKey string) bool

	// Methods for handling errors.
	SendInternalError()
	SendIncorrectKey()
	SendNoSuchKey()
	SendInvalidURI()
}
