package decoder

import (
	"encoding/json"
)

type Decoder interface {
	DecodeResponse(resp []byte, dst any) error
}

type SimpleDecoder struct {
	Decode func([]byte, any) error
}

func New(decoder func([]byte, any) error) Decoder {
	return &SimpleDecoder{decoder}
}

func (d *SimpleDecoder) DecodeResponse(b []byte, dst any) error {
	return d.Decode(b, dst)
}

var (
	JSONDecoder = New(json.Unmarshal)
)
