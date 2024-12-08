package go_rpc

import (
	"context"
	"errors"
	"go-rpc/message"
	"go-rpc/serialize"
	"net"
	"reflect"
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

		resp, err := s.Invoke(context.Background(), req)
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
	if err != nil {
		return nil, err
	}

	return &message.Response{
		RequestId:  req.RequestId,
		Version:    req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
		Data:       respData,
	}, nil
}

type reflectionStub struct {
	s          Service
	value      reflect.Value
	serializes map[uint8]serialize.Serializer
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	method := s.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
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
	if results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}
	return serializer.Encode(results[0].Interface())
}
