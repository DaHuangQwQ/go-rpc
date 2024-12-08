package go_rpc

import (
	"context"
	"errors"
	"go-rpc/internal/errs"
	"go-rpc/message"
	"go-rpc/serialize"
	"net"
	"reflect"
	"strconv"
	"time"
)

type Server struct {
	services   map[string]reflectionStub
	serializes map[uint8]serialize.Serializer
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]reflectionStub, 16),
		serializes: map[uint8]serialize.Serializer{
			1: &serialize.JsonSerializer{},
			2: &serialize.ProtoSerializer{},
		},
	}
}

func (s *Server) RegisterSerializer(serializer serialize.Serializer) {
	s.serializes[serializer.Code()] = serializer
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s:          service,
		value:      reflect.ValueOf(service),
		serializes: s.serializes,
	}
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		req := message.DecodeReq(reqBs)
		if err != nil {
			return err
		}

		ctx := context.Background()

		var cancel context.CancelFunc
		if deadlineStr, ok := req.Meta["deadline"]; ok {
			if deadline, er := strconv.ParseInt(deadlineStr, 10, 64); er == nil {
				ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(deadline))
			}

		}

		oneway, ok := req.Meta["one-way"]
		if ok && oneway == "true" {
			go func() {
				_, _ = s.Invoke(CtxWithOneway(ctx), req)
			}()
			cancel()
			return errs.ErrIsOneway
		}

		resp, err := s.Invoke(ctx, req)
		cancel()
		if err != nil {
			resp.Error = []byte(err.Error())
		}
		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		_, err = conn.Write(resp.Encode())
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("你要调用的服务不存在")
	}

	respData, err := service.invoke(ctx, req)

	return &message.Response{
		RequestId:  req.RequestId,
		Version:    req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
		Data:       respData,
	}, err
}

type reflectionStub struct {
	s          Service
	value      reflect.Value
	serializes map[uint8]serialize.Serializer
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	method := s.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(ctx)
	inReq := reflect.New(method.Type().In(1).Elem())

	serializer, ok := s.serializes[req.Serializer]
	if !ok {
		return nil, errors.New("go-rpc: no such serializer")
	}

	err := serializer.Decode(req.Data, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq
	results := method.Call(in)

	resp, err := serializer.Encode(results[0].Interface())
	if err != nil {
		return nil, err
	}

	if results[1].Interface() != nil {
		return resp, results[1].Interface().(error)
	}

	return resp, nil
}
