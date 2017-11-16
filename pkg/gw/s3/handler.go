package s3

import (
	"fmt"
	"net/http"
)

// Make bucket.
func (s *s3APIHandlers) makeBucket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

// Remove bucket.
func (s *s3APIHandlers) removeBucket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

// Put object into bucket.
func (s *s3APIHandlers) putObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

// Get object from bucket.
func (s *s3APIHandlers) getObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}

// Delete object from bucket.
func (s *s3APIHandlers) deleteObject(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", r)
}
