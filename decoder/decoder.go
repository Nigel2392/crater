package decoder

import (
	"io"

	"github.com/Nigel2392/jsext/v2/encoding"
)

type Decoder interface {
	DecodeResponse(resp io.ReadCloser, dst any) error
}

type SimpleDecoder struct {
	Decode func([]byte, any) error
}

func New(decoder func([]byte, any) error) Decoder {
	return &SimpleDecoder{decoder}
}

func (d *SimpleDecoder) DecodeResponse(b io.ReadCloser, dst any) error {
	var data, err = io.ReadAll(b)
	if err != nil {
		return err
	}
	return d.Decode(data, dst)
}

var (
	JSONDecoder = New(encoding.DecodeJSON[[]byte])
)
