package swim

import (
	"fmt"
)

// A Header represents the key-value pairs in a SWIM ping message header.
// Upper layer can make custom fields and compare it by using header.
type Header map[string]string

// Set sets the header entries associated with key.
func (h Header) Set(key, value string) {
	h[key] = value
}

// Get gets the value associated with the given key.
func (h Header) Get(key string) string {
	return h[key]
}

type customHeader struct {
	key     string
	compare func(string, string) bool
	notiC   chan interface{}
}

// SetCustomHeader sets a comparable custom field with the given key and function.
// If the compare result is true, then send notification via the given channel.
func (s *Server) SetCustomHeader(key string, compare func(string, string) bool, notiC chan interface{}) error {
	if key == "" || compare == nil || notiC == nil {
		return fmt.Errorf("invalid arguments")
	}

	s.headers[key] = customHeader{
		key:     key,
		compare: compare,
		notiC:   notiC,
	}

	return nil
}
