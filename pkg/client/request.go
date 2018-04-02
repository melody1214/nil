package client

import "net/http"

type RequestEvent interface {
	// Getter
	Protocol() Protocol
	ResponseWriter() http.ResponseWriter
	Request() *http.Request
	AccessKey() string

	// Method includes business logic.
	Auth(secretKey string) bool
}
