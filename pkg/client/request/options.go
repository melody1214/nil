package request

// EventFactoryOption holds the request factory options.
type EventFactoryOption func(*eventFactoryOptions)

type eventFactoryOptions struct {
	useS3 bool
}

var defaultEventFactoryOptions = eventFactoryOptions{
	useS3: true,
}

// WithS3EventFactory means allow the factory to create s3 type of requests.
func WithS3EventFactory(enabled bool) EventFactoryOption {
	return func(o *eventFactoryOptions) {
		o.useS3 = enabled
	}
}

// Option allows to set the client request options.
type Option func(*options)

type options struct {
	useS3   bool
	genSign bool
	cred    map[string]string
}

var defaultOptions = options{
	useS3:   true,
	genSign: false,
	cred: map[string]string{
		// Hash value of empty string.
		"content-hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	},
}

// WithS3 means to generate client request protocol to s3.
func WithS3(enabled bool) Option {
	return func(o *options) {
		o.useS3 = enabled
	}
}

// WithSign set the hash of payload which is required for generating sign.
func WithSign(accessKey, secretKey, region, contentHash string) Option {
	return func(o *options) {
		o.genSign = true
		o.cred["access-key"] = accessKey
		o.cred["secret-key"] = secretKey
		o.cred["region"] = region
		o.cred["content-hash"] = contentHash
	}
}
