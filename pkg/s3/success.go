package s3

import "net/http"

// SendSuccess writes ok response to the given http.responseWriter.
func SendSuccess(w http.ResponseWriter) {
	writeResponse(w, nil, http.StatusOK)
}
