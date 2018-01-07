package server

import (
	"crypto/sha256"
	"encoding/hex"
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
	accessKey := authArgs.Credential.AccessKey
	secretKey, e := s.getSecretKey(accessKey)
	if e != nil {
		return s3.ErrInternalError
	} else if secretKey == "" {
		return s3.ErrInvalidAccessKeyId
	}

	// Task 1: Create a Canonical Request for Signature Version 4.
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-canonical-request.html
	canonicalRequest := s3.GenCanonicalRequest(
		r.Method,
		r.RequestURI,
		r.URL.Query().Encode(),
		s3.GenCanonicalHeaders(r, authArgs.SignedHeaders),
		s3.GenSignedHeadersString(authArgs.SignedHeaders),
		r.Header.Get("X-Amz-Content-Sha256"),
	)

	// Task 2: Create a String to Sign for Signature Version 4.
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-create-string-to-sign.html
	sha256CanonicalRequest := sha256.Sum256([]byte(canonicalRequest))
	stringToSign := s3.GenStringToSign(
		"AWS4-HMAC-SHA256",
		r.Header.Get("X-Amz-Date"),
		authArgs.Credential.Scope(),
		hex.EncodeToString(sha256CanonicalRequest[:]),
	)

	// Task 3: Calculate the Signature for AWS Signature Version 4
	// https://docs.aws.amazon.com/ko_kr/general/latest/gr/sigv4-calculate-signature.html
	signatureKey := s3.GenSignatureKey(secretKey, authArgs.Credential.Date, authArgs.Credential.Region, authArgs.Credential.Service)

	derivedSignature := s3.GenSignature(signatureKey, stringToSign)
	if authArgs.Signature != derivedSignature {
		return s3.ErrSignatureDoesNotMatch
	}

	return s3.ErrNone
}

func (s *Server) getSecretKey(accessKey string) (string, error) {
	// Lookup cache first.
	if sk := s.authCache.Get(accessKey); sk != nil {
		return sk.(string), nil
	}

	conn, err := nilrpc.Dial(s.cfg.FirstMds, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	req := &nilrpc.GetCredentialRequest{AccessKey: accessKey}
	res := &nilrpc.GetCredentialResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call("Server.GetCredential", req, res); err != nil {
		return "", err
	}

	if res.Exist == false {
		return "", nil
	}

	return res.SecretKey, nil
}
