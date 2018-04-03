package request

// Option holds the request factory options.
type Option func(*options)

type options struct {
	useS3 bool
}

var defaultOptions = options{
	useS3: true,
}

// WithS3 means allow the factory to create s3 type of requests.
func WithS3(enabled bool) Option {
	return func(o *options) {
		o.useS3 = enabled
	}
}
