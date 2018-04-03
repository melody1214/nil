package auth

// Repository provides an authentication cache.
type Repository interface {
	Find(accessKey string) (secretKey string, ok bool)
	Add(accessKey string, secretKey string)
}
