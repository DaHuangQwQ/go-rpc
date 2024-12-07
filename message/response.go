package message

import (
	"bytes"
	"encoding/binary"
)

type Response struct {
	HeadLength uint32
	BodyLength uint32
	RequestId  uint32
	Version    uint8
	Compresser uint8
	Serializer uint8

	Error []byte
	Data  []byte
}

func (res *Response) CalculateHeaderLength() {
	res.HeadLength = 15 + uint32(len(res.Error)) + 1
}

func (res *Response) CalculateBodyLength() {
	res.BodyLength = uint32(len(res.Data))
}

func (res *Response) Encode() []byte {
	bs := make([]byte, res.HeadLength+res.BodyLength)

	binary.BigEndian.PutUint32(bs[0:4], res.HeadLength)
	binary.BigEndian.PutUint32(bs[4:8], res.BodyLength)
	binary.BigEndian.PutUint32(bs[8:12], res.RequestId)
	bs[12] = res.Version
	bs[13] = res.Compresser
	bs[14] = res.Serializer

	cur := bs[15:]

	copy(cur, res.Error)
	cur = cur[len(res.Error):]

	cur[0] = '\n'
	cur = cur[1:]

	copy(cur, res.Data)

	return bs
}

func DecodeRes(data []byte) *Response {
	res := &Response{}

	res.HeadLength = binary.BigEndian.Uint32(data[0:4])
	res.BodyLength = binary.BigEndian.Uint32(data[4:8])
	res.RequestId = binary.BigEndian.Uint32(data[8:12])
	res.Version = data[12]
	res.Compresser = data[13]
	res.Serializer = data[14]

	cur := data[15:res.HeadLength]
	index := bytes.IndexByte(cur, '\n')

	if len(cur[:index]) != 0 {
		res.Error = cur[:index]
	}

	if res.BodyLength != 0 {
		res.Data = data[res.HeadLength:]
	}

	return res
}
