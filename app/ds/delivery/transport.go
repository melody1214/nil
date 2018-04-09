package delivery

import (
	"net/http"

	"github.com/chanyoung/nil/app/ds/usecase/object"
	"github.com/gorilla/mux"
)

func makeHandler(oh object.ObjectHandlers) http.Handler {
	r := mux.NewRouter()

	// API routers.
	ar := r.PathPrefix("/").Subrouter()
	br := ar.PathPrefix("/{bucket}").Subrouter()
	or := br.PathPrefix("/{object:.+}").Subrouter()

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

// objectTypeBytes returns type bytes which is used to multiplexing.
func objectTypeBytes() []byte {
	return []byte{
		0x44, // 'D' of DELETE
		0x47, // 'G' of GET
		0x50, // 'P' of POST, PUT
	}
}

// adminTypeBytes returns rpc type bytes which is used to multiplexing.
func adminTypeBytes() []byte {
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
