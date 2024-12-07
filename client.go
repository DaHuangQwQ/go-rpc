package go_rpc

import (
	"context"
	"github.com/silenceper/pool"
	"go-rpc/message"
	"net"
	"time"
)

type Client struct {
	pool pool.Pool
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
		pool: p,
	}, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	req.CalculateHeaderLength()
	req.CalculateBodyLength()
	data := req.Encode()
	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return message.DecodeRes(resp), nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
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
	return ReadMsg(conn)
}
