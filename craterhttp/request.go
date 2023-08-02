package craterhttp

import (
	"github.com/Nigel2392/jsext/v2/fetch"
	"github.com/Nigel2392/mux"
)

type Request = fetch.Request

type RequestFunc func(mux.Variables) (*Request, error)

func NewRequestFunc(method string, url string, body any) RequestFunc {
	return func(v mux.Variables) (*Request, error) {
		return NewRequest(method, url, body)
	}
}

func NewRequest(method string, url string, body any) (*Request, error) {
	var r = fetch.NewRequest(method, url)
	if body != nil {
		if err := r.SetBody(body); err != nil {
			return nil, err
		}
	}
	return r, nil
}
