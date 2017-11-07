package rest

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRESTEndPoints(t *testing.T) {
	server := httptest.NewServer(NewServer())
	addr := server.Listener.Addr().String()
	log.Println("Test server is listening on ", addr)

	testCases := []struct {
		method, key, content, response string
		status                         int
	}{
		{"GET", "Hello", "", "404 page not found\n", 404},
		{"PUT", "Hello", "World!!", "", 200},
		{"GET", "Hello", "", "World!!", 200},
		{"DELETE", "Hello", "", "", 200},
	}

	for _, c := range testCases {
		req, err := http.NewRequest(c.method, "http://"+addr+objectBasePath+c.key, strings.NewReader(c.content))
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		b, _ := ioutil.ReadAll(resp.Body)
		if string(b) != c.response {
			t.Errorf("%s %s: expected response %q, got %q", req.Method, req.URL.Path, c.response, string(b))
		}
	}
}
