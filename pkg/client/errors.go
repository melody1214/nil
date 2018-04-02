package client

import "errors"

// ErrInvalidProtocol means that there is no matched protocol.
var ErrInvalidProtocol = errors.New("invalid protocol")
