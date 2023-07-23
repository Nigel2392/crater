package craterhttp

import (
	"context"
	"syscall/js"

	"github.com/Nigel2392/jsext/v2/fetch"
	"github.com/Nigel2392/mux"
)

type Request fetch.Request

type RequestFunc func(mux.Variables) (*Request, error)

func NewRequestFunc(method string, url string, body any) RequestFunc {
	return func(v mux.Variables) (*Request, error) {
		return NewRequest(method, url, body)
	}
}

func NewRequest(method string, url string, body any) (*Request, error) {
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

func (r *Request) Context() context.Context {
	return (*fetch.Request)(r).Context()
}

func (r *Request) SetContext(ctx context.Context) {
	(*fetch.Request)(r).SetContext(ctx)
}

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
