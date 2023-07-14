package craterhttp

import (
	"syscall/js"

	"github.com/Nigel2392/jsext/v2/fetch"
)

type Request fetch.Request

func (r *Request) SetHeader(key string, value string) {
	(*fetch.Request)(r).SetHeader(key, value)
}

func (r *Request) AddHeader(key string, value string) {
	(*fetch.Request)(r).AddHeader(key, value)
}

func (r *Request) DeleteHeader(key string) {
	(*fetch.Request)(r).DeleteHeader(key)
}

func (r *Request) SetBody(body any) error {
	return (*fetch.Request)(r).SetBody(body)
}

func (r *Request) MarshalJS() (js.Value, error) {
	return (*fetch.Request)(r).MarshalJS()

}

type RequestFunc func(vars map[string][]string) (*Request, error)

func NewRequestFunc(method string, url string, body any) RequestFunc {
	return func(v map[string][]string) (*Request, error) {
		var r = &Request{
			Method: method,
			URL:    url,
		}
		if body != nil {
			if err := r.SetBody(body); err != nil {
				return nil, err
			}
		}
		return r, nil
	}
}
