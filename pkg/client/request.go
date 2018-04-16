package client

import "net/http"

// Request is the client request for rest API calling.
type Request interface {
	Send() (*http.Response, error)
}
