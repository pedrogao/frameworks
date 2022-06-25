package codec

import (
	"io"
)

type Header struct {
	ServiceMethod string // eg. Service.Method
	Seq           uint64 // sequence number by client
	Error         string // error represent string
}

type Codec interface {
	io.Closer
	ReadeHeader(*Header) error
	ReadBody(any) error
	Write(*Header, any) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob" // not implemented
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[JsonType] = NewJsonCodec
}
