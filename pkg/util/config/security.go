package config

// Security related information container.
type Security struct {
	// CertsDir is the directory path which contains files for https configuration.
	CertsDir string

	// RootCAPem is the RootCA.pem file name.
	RootCAPem string
	// ServerKey is the server private key file name.
	ServerKey string
	// ServerCrt is the server certification file name.
	ServerCrt string
}
