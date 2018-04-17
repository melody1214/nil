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
	compare func(string, string) bool
	notiC   chan interface{}
}

// RegisterCustomHeader registers a comparable custom field with the given key and function.
// If the compare result is true, then send notification via the given channel.
func (s *Server) RegisterCustomHeader(key, value string, compare func(have string, rcv string) bool, notiC chan interface{}) error {
	if key == "" || compare == nil || notiC == nil {
		return fmt.Errorf("invalid arguments")
	}

	s.headerFunc[key] = &customHeader{
		compare: compare,
		notiC:   notiC,
	}
	s.header.Set(key, value)

	return nil
}

// SetCustomHeader sets the given value into the custom header.
func (s *Server) SetCustomHeader(key, value string) {
	s.header.Set(key, value)
}
