package grpc

import (
	"fmt"
	"net/http"
)

// Test.
func (g *gRPCHandlers) test(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("gRPC routing success, \nr: %#v r.Header: %v\n", r, r.Header)
}
