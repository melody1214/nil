package delivery

import (
	"net/http"

	"github.com/gorilla/mux"
)

func makeHandler(bs BucketService, os ObjectService) http.Handler {
	r := mux.NewRouter()

	// API routers.
	ar := r.PathPrefix("/").Subrouter()
	br := ar.PathPrefix("/{bucket}").Subrouter()
	or := br.PathPrefix("/{object:.+}").Subrouter()

	// Bucket request handlers
	br.Methods("PUT").HandlerFunc(bs.MakeBucketHandler)
	br.Methods("DELETE").HandlerFunc(bs.RemoveBucketHandler)

	// Object request handlers
	or.Methods("PUT").HandlerFunc(os.PutObjectHandler)
	or.Methods("GET").HandlerFunc(os.GetObjectHandler)
	or.Methods("DELETE").HandlerFunc(os.DeleteObjectHandler)

	return r
}

// httpTypeBytes returns rpc type bytes which is used to multiplexing.
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
		0x01, // rpcRaft
		0x02, // rpcNil
	}
}
