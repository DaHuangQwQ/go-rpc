package message

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	HeadLength uint32
	BodyLength uint32
	RequestId  uint32
	Version    uint8
	Compresser uint8
	Serializer uint8

	ServiceName string
	MethodName  string

	Meta map[string]string

	Data []byte
}

func (req *Request) CalculateHeaderLength() {
	req.HeadLength = 15 + uint32(len(req.ServiceName)) + 1 + uint32(len(req.MethodName)) + 1
	for key, val := range req.Meta {
		req.HeadLength += uint32(len(key)+1) + uint32(len(val)+1)
	}
}

func (req *Request) CalculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}

func (req *Request) Encode() []byte {
	bs := make([]byte, req.HeadLength+req.BodyLength)

	binary.BigEndian.PutUint32(bs[:4], req.HeadLength)
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	binary.BigEndian.PutUint32(bs[8:12], req.RequestId)
	bs[12] = req.Version
	bs[13] = req.Compresser
	bs[14] = req.Serializer

	cur := bs[15:]

	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]

	cur[0] = '\n'
	cur = cur[1:]

	copy(cur, req.MethodName)
	cur = cur[len(req.MethodName):]

	cur[0] = '\n'
	cur = cur[1:]

	for key, val := range req.Meta {
		copy(cur, key)
		cur = cur[len(key):]

		cur[0] = '\r'
		cur = cur[1:]

		copy(cur, val)
		cur = cur[len(val):]

		cur[0] = '\n'
		cur = cur[1:]
	}

	copy(cur, req.Data)

	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}

	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	req.RequestId = binary.BigEndian.Uint32(data[8:12])
	req.Version = data[12]
	req.Compresser = data[13]
	req.Serializer = data[14]

	header := data[15:req.HeadLength]
	index := bytes.IndexByte(header, '\n')
	req.ServiceName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, '\n')
	req.MethodName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, '\n')

	if index != -1 {
		meta := make(map[string]string, 4)
		for index != -1 {
			pair := header[:index]

			pairIndex := bytes.IndexByte(pair, '\r')

			key := string(pair[:pairIndex])
			val := string(pair[pairIndex+1:])
			meta[key] = val

			header = header[index+1:]
			index = bytes.IndexByte(header, '\n')
		}
		req.Meta = meta
	}

	if req.BodyLength != 0 {
		req.Data = data[req.HeadLength:]
	}

	return req
}
