package serialize

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

type ProtoSerializer struct {
}

func (s *ProtoSerializer) Code() uint8 {
	return 2
}

func (s *ProtoSerializer) Encode(val any) ([]byte, error) {
	msg, ok := val.(proto.Message)
	if !ok {
		return nil, errors.New("micro: 必须是 proto.Message")
	}
	return proto.Marshal(msg)
}

func (s *ProtoSerializer) Decode(data []byte, val any) error {
	msg, ok := val.(proto.Message)
	if !ok {
		return errors.New("micro: 必须是 proto.Message")
	}
	return proto.Unmarshal(data, msg)
}
