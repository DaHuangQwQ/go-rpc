package message

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReqDecodeEncode(t *testing.T) {
	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req: &Request{
				Version:     1,
				Compresser:  12,
				Serializer:  123,
				ServiceName: "UserService",
				MethodName:  "GetById",
				Meta: map[string]string{
					"err": "my err",
				},
				Data: []byte("hello world"),
			},
		},
		{
			name: "no meta",
			req: &Request{
				Version:     1,
				Compresser:  1,
				Serializer:  1,
				ServiceName: "UserService",
				MethodName:  "GetById",
				Data:        []byte("hello world"),
			},
		},
		{
			name: "no data",
			req: &Request{
				Version:     1,
				Compresser:  1,
				Serializer:  1,
				ServiceName: "UserService",
				MethodName:  "GetById",
				Meta: map[string]string{
					"err":      "my err",
					"trace id": "123",
				},
			},
		},
		{
			name: "data with \n",
			req: &Request{
				Version:     1,
				Compresser:  1,
				Serializer:  1,
				ServiceName: "UserService",
				MethodName:  "GetById",
				Meta: map[string]string{
					"err":      "my err",
					"trace id": "123",
				},
				Data: []byte("hello world\n"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.CalculateHeaderLength()
			tc.req.CalculateBodyLength()
			data := tc.req.Encode()
			req := DecodeReq(data)
			require.Equal(t, tc.req, req)
		})
	}
}
