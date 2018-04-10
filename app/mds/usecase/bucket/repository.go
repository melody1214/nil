package bucket

// Repository provides access to bucket database.
type Repository interface {
	MakeBucket(bucketName, accessKey, region string) (err error)
}
