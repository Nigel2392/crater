package tasks

import (
	"context"

	"github.com/Nigel2392/crater/craterhttp"
	"github.com/Nigel2392/crater/tasker"
	"github.com/Nigel2392/jsext/v2/errs"
)

var (
	ErrRequestNotOK = errs.Error("request not ok")
)

type HttpRequestOptions struct {
	RequestFunc func(context.Context) (*craterhttp.Request, error)
	OnSuccess   func(context.Context, *craterhttp.Response) error
	Client      *craterhttp.Client

	tasker.Task
}

func HttpRequest(option HttpRequestOptions) tasker.Task {
	if option.Client == nil {
		option.Client = craterhttp.DefaultClient
	}
	var t = tasker.Task{
		Name:      option.Name,
		OnError:   option.OnError,
		Duration:  option.Duration,
		OnDequeue: option.OnDequeue,
		Func: func(ctx context.Context) error {
			if option.Func != nil {
				if err := option.Func(ctx); err != nil {
					return err
				}
			}
			var req, err = option.RequestFunc(ctx)
			if err != nil {
				return err
			}
			resp, err := option.Client.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode >= 400 {
				return ErrRequestNotOK
			}
			if option.OnSuccess == nil {
				return nil
			}
			return option.OnSuccess(ctx, resp)
		},
	}
	return t
}
