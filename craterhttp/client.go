package craterhttp

import (
	"context"
	"time"

	"github.com/Nigel2392/jsext/v2/errs"
)

type Client struct {
	DefaultHeaders func(r *Request) map[string][]string
	OnResponse     func(*Response) error
	Timeout        time.Duration
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		Timeout: timeout,
	}
}

var DefaultClient = NewClient(5 * time.Second)

func (c *Client) Do(r *Request) (*Response, error) {

	if r == nil {
		return nil, errs.Error("request is nil")
	}

	if r.URL == "" {
		return nil, errs.Error("url is empty")
	}

	if r.Method == "" {
		r.Method = "GET"
	}

	if c.Timeout > 0 {
		var ctx = r.Context()
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
		r.SetContext(ctx)
	}

	if c.DefaultHeaders != nil {
		var headers = c.DefaultHeaders(r)
		for key, value := range headers {
			for _, v := range value {
				r.AddHeader(key, v)
			}
		}
	}
	var resp, err = Fetch(r)
	if err != nil {
		return nil, err
	}

	if c.OnResponse != nil {
		if err := c.OnResponse(resp); err != nil {
			return nil, err
		}
	}

	return resp, nil
}
