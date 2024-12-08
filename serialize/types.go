package serialize

type Serializer interface {
	Code() uint8
	Encode(val any) ([]byte, error)
	// Decode val 是一个结构体指针
	Decode(data []byte, val any) error
}
