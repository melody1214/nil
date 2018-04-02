package client

import "net/http"

type RequestEvent interface {
	Protocol() Protocol
	ResponseWriter() http.ResponseWriter
	Request() *http.Request
}
