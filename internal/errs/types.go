package errs

import "errors"

var (
	ErrIsOneway = errors.New("go-rpc: warn! this is oneway")
)
