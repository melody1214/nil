package auth

// Repository provides access to auth database.
type Repository interface {
	FindSecretKey(accessKey string) (secretKey string, err error)
}
