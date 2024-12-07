package go_rpc

import (
	"context"
	"go-rpc/message"
)

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}
