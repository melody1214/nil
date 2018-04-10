package nilrpc

// MAUGetCredentialRequest requests a credential for the given access key.
type MAUGetCredentialRequest struct {
	AccessKey string
}

// MAUGetCredentialResponse response the credential.
type MAUGetCredentialResponse struct {
	Exist     bool
	AccessKey string
	SecretKey string
}
