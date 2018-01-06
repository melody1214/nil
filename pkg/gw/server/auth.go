package server

import (
	"net/http"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
)

func (s *Server) authRequest(r *http.Request) s3.ErrorCode {
	// Get authentication string from header.
	authString := r.Header.Get("Authorization")

	// Check the sign version is supported.
	if err := s3.ValidateSignVersion(authString); err != s3.ErrNone {
		return err
	}

	// Parse auth string.
	authArgs, err := s3.ParseSignV4(authString)
	if err != s3.ErrNone {
		return err
	}

	// Make key.
	// key = accessKey/secretKey
	var key string
	ak := authArgs.GetAccessKey()
	// Lookup cache first.
	if sk := s.authCache.Get(ak); sk != nil {
		key = ak + "/" + sk.(string)
	} else {
		conn, err := nilrpc.Dial(s.cfg.FirstMds, nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			return s3.ErrInternalError
		}
		defer conn.Close()

		req := &nilrpc.GetCredentialRequest{AccessKey: ak}
		res := &nilrpc.GetCredentialResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call("Server.GetCredential", req, res); err != nil {
			return s3.ErrInternalError
		}

		if res.Exist == false {
			return s3.ErrInvalidAccessKeyId
		}

		key = ak + "/" + res.SecretKey
	}

	// Make V4 checker.
	// User provided SignV4
	_ = key

	return s3.ErrNone
}
