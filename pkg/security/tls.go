package security

import (
	"crypto/tls"
)

const (
	// Public certificate file.
	publicCertFile = "public.crt"

	// Private key file for tls protocol.
	privateKeyFile = "private.key"
)

// DefaultTLSConfig loads default tls config with cert and key files.
func DefaultTLSConfig() *tls.Config {
	return &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},

		MinVersion: tls.VersionTLS12,

		SessionTicketsDisabled: true,
	}
}

// PublicCertFile returns public certificate file name.
func PublicCertFile() string {
	return publicCertFile
}

// PrivateKeyFile returns private key file name.
func PrivateKeyFile() string {
	return privateKeyFile
}
