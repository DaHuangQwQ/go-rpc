package message

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResDecodeEncode(t *testing.T) {
	testCases := []struct {
		name string
		req  *Response
	}{
		{
			name: "normal",
			req: &Response{
				Version:    1,
				Compresser: 12,
				Serializer: 123,
				Error:      []byte("my error"),
				Data:       []byte("hello world"),
			},
		},
		{
			name: "no error",
			req: &Response{
				Version:    1,
				Compresser: 1,
				Serializer: 1,
				Data:       []byte("hello world"),
			},
		},
		{
			name: "no data",
			req: &Response{
				Version:    1,
				Compresser: 12,
				Serializer: 123,
				Error:      []byte("my error ni hao"),
			},
		},
		{
			name: "data with \n",
			req: &Response{
				Version:    1,
				Compresser: 1,
				Serializer: 1,
				Error:      []byte("my error"),
				Data:       []byte("hello world\n"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.CalculateHeaderLength()
			tc.req.CalculateBodyLength()
			data := tc.req.Encode()
			req := DecodeRes(data)
			require.Equal(t, tc.req, req)
		})
	}
}
