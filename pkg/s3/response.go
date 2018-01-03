package s3

import (
	"bytes"
	"encoding/xml"
	"net/http"
)

// writeResponse writes the given message to the http writer.
func writeResponse(w http.ResponseWriter, response interface{}, httpCode int) {
	// Write headers.
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(httpCode)

	// Write encoded response.
	w.Write(encodeResponse(response))

	// Flush if flusher is supported.
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func encodeResponse(response interface{}) []byte {
	var bytesBuffer bytes.Buffer
	bytesBuffer.WriteString(xml.Header)
	e := xml.NewEncoder(&bytesBuffer)
	e.Encode(response)
	return bytesBuffer.Bytes()
}
