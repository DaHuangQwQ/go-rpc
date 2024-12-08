package serialize

import "encoding/json"

type JsonSerializer struct {
}

func (s *JsonSerializer) Code() uint8 {
	return 1
}

func (s *JsonSerializer) Encode(val any) ([]byte, error) {
	return json.Marshal(val)
}

func (s *JsonSerializer) Decode(data []byte, val any) error {
	return json.Unmarshal(data, val)
}
