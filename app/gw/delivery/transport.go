package delivery

import (
	"net/http"

	"github.com/chanyoung/nil/app/gw/usecase/client"
	"github.com/gorilla/mux"
)

func makeHandler(ch client.Handlers) http.Handler {
	r := mux.NewRouter()

	// API routers.
	ar := r.PathPrefix("/").Subrouter()
	cr := ar.PathPrefix("/chunk").Subrouter()
	br := ar.PathPrefix("/{bucket}").Subrouter()
	or := br.PathPrefix("/{object:.+}").Subrouter()

	// Chunk request handlers
	cr.Methods("GET").HandlerFunc(ch.GetChunkHandler)
	cr.Methods("PUT").HandlerFunc(ch.RenameChunkHandler)

	// Bucket request handlers
	br.Methods("PUT").HandlerFunc(ch.MakeBucketHandler)
	br.Methods("DELETE").HandlerFunc(ch.RemoveBucketHandler)

	// Object request handlers
	or.Methods("PUT").HandlerFunc(ch.PutObjectHandler)
	or.Methods("GET").HandlerFunc(ch.GetObjectHandler)
	or.Methods("DELETE").HandlerFunc(ch.DeleteObjectHandler)

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
