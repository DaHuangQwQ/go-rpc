package go_rpc

import (
	"context"
	"github.com/silenceper/pool"
	"go-rpc/internal/errs"
	"go-rpc/message"
	"go-rpc/serialize"
	"net"
	"time"
)

type Client struct {
	pool       pool.Pool
	serializer serialize.Serializer
}

func NewClient(addr string) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap:  1,
		MaxCap:      30,
		MaxIdle:     10,
		IdleTimeout: time.Minute,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, time.Second*3)
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		pool:       p,
		serializer: &serialize.JsonSerializer{},
	}, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ch := make(chan struct{})

	var (
		resp *message.Response
		err  error
	)
	go func() {
		resp, err = c.doInvoke(ctx, req)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	req.CalculateHeaderLength()
	req.CalculateBodyLength()
	data := req.Encode()
	resp, err := c.Send(ctx, data)
	if err != nil {
		return nil, err
	}
	return message.DecodeRes(resp), nil
}

func (c *Client) Send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = c.pool.Put(val)
	}()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	if isOneway(ctx) {
		return nil, errs.ErrIsOneway
	}
	return ReadMsg(conn)
}
