package craterhttp

import (
	"bytes"
	"io"
	"strconv"

	"github.com/Nigel2392/crater/decoder"
	"github.com/Nigel2392/jsext/v2/fetch"
)

type Decoder interface {
	DecodeResponse(resp io.ReadCloser, dst any) error
}

type Response fetch.Response

func (r *Response) String() string {
	if r == nil {
		return "<Response: nil>"
	}
	var b bytes.Buffer
	b.WriteString("Response: {\n")
	b.WriteString("\tStatus: ")
	b.WriteString(strconv.Itoa(r.StatusCode))
	b.WriteString("\n")
	if len(r.Headers) > 0 {
		b.WriteString("\tHeaders: ")
		for key, values := range r.Headers {
			b.WriteString("\t\t")
			b.WriteString(key)
			b.WriteString(": ")
			b.Grow(2 * len(values))
			for i, value := range values {
				b.WriteString(value)
				if i < len(values)-1 {
					b.WriteString(", ")
				}
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("}")
	return b.String()
}

func (r *Response) JSON(dst any) error {
	return decoder.JSONDecoder.DecodeResponse(r.Body, dst)
}
