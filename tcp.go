package go_rpc

import (
	"encoding/binary"
	"net"
)

const numOfLengthBytes = 8

func ReadMsg(conn net.Conn) ([]byte, error) {
	lenBs := make([]byte, numOfLengthBytes)

	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}

	headerLength := binary.BigEndian.Uint32(lenBs[:4])
	bodyLength := binary.BigEndian.Uint32(lenBs[4:8])
	length := bodyLength + headerLength
	data := make([]byte, length)
	_, err = conn.Read(data[8:])
	copy(data[:8], lenBs)
	return data, err
}
