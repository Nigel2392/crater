package craterhttp

import "github.com/Nigel2392/jsext/v2/fetch"

func Fetch(r *Request) (*Response, error) {
	var resp, err = fetch.Fetch((*fetch.Request)(r))
	if err != nil {
		return nil, err
	}
	return (*Response)(resp), nil
}
