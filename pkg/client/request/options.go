package request

type Option func(*options)

type options struct {
	useS3 bool
}

var defaultOptions = options{
	useS3: true,
}

func WithS3(enabled bool) Option {
	return func(o *options) {
		o.useS3 = enabled
	}
}
