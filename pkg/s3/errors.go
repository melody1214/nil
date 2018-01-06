package s3

import (
	"net/http"
)

// See https://docs.aws.amazon.com/ko_kr/AmazonS3/latest/API/ErrorResponses.html

// Error is the (xml body of) s3 error response message.
type Error struct {
	Code      string
	Message   string
	Resource  string
	RequestId string
}

// ErrorInfo contains the information for error codes.
type ErrorInfo struct {
	Code        string
	Description string
	HTTPCode    int
}

// ErrorCode lists from Amazon S3
type ErrorCode int

const (
	ErrNone = iota
	ErrAccessDenied
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
	ErrInvalidRequestSignVersion
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

var errorInfos = map[ErrorCode]ErrorInfo{
	ErrAccessDenied: {
		Code:        "AccessDenied",
		Description: "Access denied.",
		HTTPCode:    http.StatusForbidden,
	},
	ErrBucketAlreadyExists: {
		Code:        "BucketAlreadyExists",
		Description: "The requested bucket name is not available. The bucket namespace is shared by all users of the system. Please select a different name and try again.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrBucketNotEmpty: {
		Code:        "BucketNotEmpty",
		Description: "The bucket you tried to delete is not empty.",
		HTTPCode:    http.StatusConflict,
	},
	ErrInternalError: {
		Code:        "InternalError",
		Description: "We encountered an internal error. Please try again.",
		HTTPCode:    http.StatusInternalServerError,
	},
	ErrInvalidAccessKeyId: {
		Code:        "InvalidAccessKeyId",
		Description: "The AWS access key Id you provided does not exist in our records.",
		HTTPCode:    http.StatusForbidden,
	},
	ErrInvalidBucketName: {
		Code:        "InvalidBucketName",
		Description: "The specified bucket is not valid.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrInvalidBucketState: {
		Code:        "InvalidBucketState",
		Description: "The request is not valid with the current state of the bucket.",
		HTTPCode:    http.StatusConflict,
	},
	ErrInvalidLocationConstraint: {
		Code:        "InvalidLocationConstraint",
		Description: "The specified location constraint is not valid. For more information about regions, see How to Select a Region for Your Buckets.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrInvalidRequestSignVersion: {
		Code:        "InvalidRequest",
		Description: "The authorization mechanism you have provided is not supported. Please use AWS4-HMAC-SHA256.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrInvalidURI: {
		Code:        "InvalidURI",
		Description: "Couldn't parse the specified URI.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrKeyTooLongError: {
		Code:        "KeyTooLongError",
		Description: "Your key is too long.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrMissingSecurityHeader: {
		Code:        "MissingSecurityHeader",
		Description: "Your request is missing a required header.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrNoSuchBucket: {
		Code:        "NoSuchBucket",
		Description: "The specified key does not exist.",
		HTTPCode:    http.StatusNotFound,
	},
	ErrNotSignedUp: {
		Code:        "NotSignedUp",
		Description: "Your account is not signed up for the Amazon S3 service. You must sign up before you can use Amazon S3. You can sign up at the following URL: https://aws.amazon.com/s3",
		HTTPCode:    http.StatusForbidden,
	},
	ErrRequestTimeout: {
		Code:        "RequestTimeout",
		Description: "Your socket connection to the server was not read from or written to within the timeout period.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrTooManyBuckets: {
		Code:        "TooManyBuckets",
		Description: "You have attempted to create more buckets than allowed.",
		HTTPCode:    http.StatusBadRequest,
	},
	ErrUserKeyMustBeSpecified: {
		Code:        "UserKeyMustBeSpecified",
		Description: "The bucket POST must contain the specified field name. If it is specified, check the order of the fields.",
		HTTPCode:    http.StatusBadRequest,
	},
}

// SendError writes error response to the given http.responseWriter.
func SendError(w http.ResponseWriter, code ErrorCode, resource, requestId string) {
	e := GetErrorInfo(code)

	response := Error{
		Code:      e.Code,
		Message:   e.Description,
		Resource:  resource,
		RequestId: requestId,
	}

	writeResponse(w, response, e.HTTPCode)
}

// GetErrorInfo returns a information structure for the S3 error code.
// It returns the empty structure if the code is unknown.
func GetErrorInfo(code ErrorCode) ErrorInfo {
	return errorInfos[code]
}
