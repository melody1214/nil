package s3

// See https://docs.aws.amazon.com/ko_kr/AmazonS3/latest/API/ErrorResponses.html

// Error is the container for all error elements.
type Error struct {
	Code      string
	Message   string
	Resource  string
	RequestId string
}

// ErrorCode lists from Amazon S3
type ErrorCode int

const (
	ErrAccessDenied = iota
	ErrAccountProblem
	ErrBadDigest
	ErrBucketAlreadyExists
	ErrBucketAlreadyOwnedByYou
	ErrBucketNotEmpty
	ErrCredentialsNotSupported
	ErrEntityTooSmall
	ErrEntityTooLarge
	ErrExpiredToken
	ErrIllegalVersioningConfigurationException
	ErrIncompleteBody
	ErrIncorrectNumberOfFilesInPostRequest
	ErrInlineDataTooLarge
	ErrInternalError
	ErrInvalidAccessKeyId
	ErrInvalidAddressingHeader
	ErrInvalidArgument
	ErrInvalidBucketName
	ErrInvalidBucketState
	ErrInvalidDigest
	ErrInvalidEncryptionAlgorithmError
	ErrInvalidLocationConstraint
	ErrInvalidObjectState
	ErrInvalidPayer
	ErrInvalidRange
	ErrInvalidRequestHMAC
	ErrInvalidSecurity
	ErrInvalidStorageClass
	ErrInvalidToken
	ErrInvalidURI
	ErrKeyTooLongError
	ErrMalformedACLError
	ErrMalformedPOSTRequest
	ErrMalformedXML
	ErrMaxMessageLengthExceeded
	ErrMaxPostPreDataLengthExceededError
	ErrMetadataTooLarge
	ErrMethodNotAllowed
	ErrMissingContentLength
	ErrMissingRequestBodyError
	ErrMissingSecurityHeader
	ErrNoSuchBucket
	ErrNoSuchKey
	ErrNoSuchUpload
	ErrNoSuchVersion
	ErrNotImplemented
	ErrNotSignedUp
	ErrNoSuchBucketPolicy
	ErrOperationAborted
	ErrPermanentRedirect
	ErrPreconditionFailed
	ErrRedirect
	ErrRestoreAlreadyInProgress
	ErrRequestTimeout
	ErrRequestTimeTooSkewed
	ErrSignatureDoesNotMatch
	ErrServiceUnavailable
	ErrSlowDown
	ErrTemporaryRedirect
	ErrTokenRefreshRequired
	ErrTooManyBuckets
	ErrUnexpectedContent
	ErrUnresolvableGrantByEmailAddress
	ErrUserKeyMustBeSpecified
)
