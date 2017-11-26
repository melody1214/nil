package rest

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/chanyoung/nil/pkg/kv"
)

const objectBasePath = "/obj/"

// Server serves object with restful API.
// Each osd has an own rest server.
type Server struct {
	mux *http.ServeMux
	kv  kv.DB
}

// NewServer creates a new server object.
func NewServer() *Server {
	s := &Server{
		mux: http.NewServeMux(),
		kv:  kv.New(),
	}
	s.init()

	return s
}

// Start starts to listen and serve http requests.
func (s *Server) Start() error {
	return http.ListenAndServe("localhost:3389", s)
}

// ServeHTTP is required to implement the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) init() {
	s.mux.HandleFunc(objectBasePath, s.handleObjectAction)
}

func (s *Server) handleObjectAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleObjectGetAction(w, r)
	case "PUT", "POST":
		s.handleObjectPutAction(w, r)
	case "DELETE":
		s.handleObjectDeleteAction(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (s *Server) handleObjectGetAction(w http.ResponseWriter, r *http.Request) {
	id, err := objID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Object ID: ", id)
	http.ServeFile(w, r, id)
}

func (s *Server) handleObjectPutAction(w http.ResponseWriter, r *http.Request) {
	id, err := objID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, err := os.Create(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleObjectDeleteAction(w http.ResponseWriter, r *http.Request) {
	id, err := objID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	os.Remove(id)
	w.WriteHeader(http.StatusOK)
}

func objID(path string) (string, error) {
	return url.QueryUnescape(strings.TrimPrefix(path, objectBasePath))
}
