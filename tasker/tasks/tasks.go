package tasks

import (
	"context"
	"time"

	"github.com/Nigel2392/crater/craterhttp"
	"github.com/Nigel2392/crater/tasker"
	"github.com/Nigel2392/jsext/v2/errs"
)

var (
	ErrRequestNotOK = errs.Error("request not ok")
)

type HttpRequest struct {
	RequestFunc func() *craterhttp.Request
	OnSuccess   func(*craterhttp.Response) error
	Client      *craterhttp.Client

	Name     string
	Duration time.Duration
	OnError  func(error)
}

func HttpRequestTask(option HttpRequest) tasker.Task {
	var t = tasker.Task{
		Name:     option.Name,
		OnError:  option.OnError,
		Duration: option.Duration,
	}
	if option.Client == nil {
		option.Client = craterhttp.DefaultClient
	}
	t.Func = func(ctx context.Context) error {
		var req = option.RequestFunc()
		var resp, err = option.Client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return ErrRequestNotOK
		}
		if option.OnSuccess == nil {
			return nil
		}
		return option.OnSuccess(resp)
	}
	return t
}
