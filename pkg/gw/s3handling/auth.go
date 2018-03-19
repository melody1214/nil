package s3handling

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/s3"
)

func (h *Handler) authRequest(r *http.Request) (accessKey string, err s3.ErrorCode) {
	// Get authentication string from header.
	authString := r.Header.Get("Authorization")

	// Check the sign version is supported.
	if err := s3.ValidateSignVersion(authString); err != s3.ErrNone {
		return "", err
	}

	// Parse auth string.
	authArgs, err := s3.ParseSignV4(authString)
	if err != s3.ErrNone {
		return "", err
	}

	// Make key.
	accessKey = authArgs.Credential.AccessKey
	secretKey, e := h.getSecretKey(accessKey)
	if e != nil {
		return accessKey, s3.ErrInternalError
	} else if secretKey == "" {
		return accessKey, s3.ErrInvalidAccessKeyId
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
		return accessKey, s3.ErrSignatureDoesNotMatch
	}

	return accessKey, s3.ErrNone
}

func (h *Handler) getSecretKey(accessKey string) (string, error) {
	// 1. Lookup cache first.
	if sk := h.authCache.Get(accessKey); sk != nil {
		return sk.(string), nil
	}

	// 2. Lookup mds from cluster map.
	mds, err := h.clusterMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		return "", nil
	}

	// 3. Try dial to mds.
	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	req := &nilrpc.GetCredentialRequest{AccessKey: accessKey}
	res := &nilrpc.GetCredentialResponse{}

	// 4. Request the secret key.
	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.GetCredential.String(), req, res); err != nil {
		return "", err
	}

	// 5. No matched key.
	if res.Exist == false {
		return "", nil
	}

	return res.SecretKey, nil
}
