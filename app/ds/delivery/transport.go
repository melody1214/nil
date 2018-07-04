package delivery

import (
	"net/http"

	"github.com/chanyoung/nil/app/ds/application/object"
	"github.com/gorilla/mux"
)

func makeHandler(oh object.Handlers) http.Handler {
	r := mux.NewRouter()

	// API routers.
	ar := r.PathPrefix("/").Subrouter()
	cr := ar.PathPrefix("/chunk").Subrouter()
	br := ar.PathPrefix("/{bucket}").Subrouter()
	or := br.PathPrefix("/{object:.+}").Subrouter()

	// Chunk request handlers
	cr.Methods("PUT").HandlerFunc(oh.PutChunkHandler)
	cr.Methods("GET").HandlerFunc(oh.GetChunkHandler)

	// Bucket request handlers
	br.Methods("PUT").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	br.Methods("PUT").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	// Object request handlers
	or.Methods("PUT").HandlerFunc(oh.PutObjectHandler)
	or.Methods("GET").HandlerFunc(oh.GetObjectHandler)
	or.Methods("DELETE").HandlerFunc(oh.DeleteObjectHandler)

	return r
}

// httpTypeBytes returns type bytes which is used to multiplexing.
func httpTypeBytes() []byte {
	return []byte{
		0x44, // 'D' of DELETE
		0x47, // 'G' of GET
		0x50, // 'P' of POST, PUT
	}
}

// rpcTypeBytes returns rpc type bytes which is used to multiplexing.
func rpcTypeBytes() []byte {
	return []byte{
		0x02, // rpcNil
	}
}

// membershipTypeBytes returns rpc type bytes which is used to multiplexing.
func membershipTypeBytes() []byte {
	return []byte{
		0x03, // rpcSwim
	}
}
