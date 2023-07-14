package messenger

import (
	"fmt"
	"syscall/js"
	"time"
)

// A messenger which will display messages to the user
type Messenger interface {
	Error(duration time.Duration, args ...any)
	Warning(duration time.Duration, args ...any)
	Info(duration time.Duration, args ...any)
	Success(duration time.Duration, args ...any)
}

type SimpleMessenger struct {
	// The element which will be used to display messages
	Element js.Value `jsc:"element"`
	// Foreground and background colors for the message
	Colors Styling `jsc:"contrast"`
}

type Styling struct {
	ForeGround Colors `jsc:"foreGround"`
	BackGround Colors `jsc:"backGround"`
	FontSize   string `jsc:"fontSize"`
	Padding    string `jsc:"padding"`
}

func (m *Styling) defaults() {
	if m.ForeGround.Error == "" {
		m.ForeGround.Error = "white"
	}
	if m.ForeGround.Warning == "" {
		m.ForeGround.Warning = "white"
	}
	if m.ForeGround.Info == "" {
		m.ForeGround.Info = "white"
	}
	if m.ForeGround.Success == "" {
		m.ForeGround.Success = "white"
	}
	if m.BackGround.Error == "" {
		m.BackGround.Error = "#e02500"
	}
	if m.BackGround.Warning == "" {
		m.BackGround.Warning = "#e09900"
	}
	if m.BackGround.Info == "" {
		m.BackGround.Info = "#008ee0"
	}
	if m.BackGround.Success == "" {
		m.BackGround.Success = "#25e000"
	}
	if m.FontSize == "" {
		m.FontSize = "24px"
	}
	if m.Padding == "" {
		m.Padding = "10px"
	}
}

type Colors struct {
	Error   string `jsc:"error"`
	Warning string `jsc:"warning"`
	Info    string `jsc:"info"`
	Success string `jsc:"success"`
}

var document = js.Global().Get("document")

// Create a new messenger
func New(style Styling) Messenger {
	var e = document.Call("createElement", "div")
	e.Set("id", "crater-messenger-container")
	e.Set("className", "message-container")
	document.Get("body").Call("appendChild", e)
	var m = &SimpleMessenger{
		Element: e,
		Colors:  style,
	}

	m.Colors.defaults()

	var jsStyle = document.Call("getElementById", "crater-messenger-style")
	if jsStyle.IsUndefined() || jsStyle.IsNull() {
		jsStyle = document.Call("createElement", "style")
		jsStyle.Set("id", "crater-messenger-style")
		jsStyle.Set("textContent", `
			.message-container {
				display: flex;
				flex-direction: column;
				justify-content: center;
				align-items: center;
				position: fixed;
				top: 0;
				left: 0;
				right: 0;
				z-index: 1000;
				cursor: pointer;
			}
			.message {
				font-size: `+m.Colors.FontSize+`;
				padding: `+m.Colors.Padding+`;
				width: 100%;
				margin: 0;
				display: flex;
				justify-content: center;
				align-items: center;
				opacity: 0;
				transform: translateY(-100%);
				transition: all 0.5s ease-in-out;
				transform-origin: top;
			}
			.message.active {
				opacity: 1;
				transform: translateY(0);
			}`)
		document.Get("head").Call("appendChild", jsStyle)
	}

	return m
}

// Display an error message
func (m *SimpleMessenger) Error(duration time.Duration, args ...any) {
	m.displayMessage(duration, "error", m.Colors.ForeGround.Error, m.Colors.BackGround.Error, args...)
}

// Display a warning message
func (m *SimpleMessenger) Warning(duration time.Duration, args ...any) {
	m.displayMessage(duration, "warning", m.Colors.ForeGround.Warning, m.Colors.BackGround.Warning, args...)
}

// Display an info message
func (m *SimpleMessenger) Info(duration time.Duration, args ...any) {
	m.displayMessage(duration, "info", m.Colors.ForeGround.Info, m.Colors.BackGround.Info, args...)
}

// Display a success message
func (m *SimpleMessenger) Success(duration time.Duration, args ...any) {
	m.displayMessage(duration, "success", m.Colors.ForeGround.Success, m.Colors.BackGround.Success, args...)
}

type OnClickFunc func()

func (f OnClickFunc) OnClick() {
	f()
}

type OnClicker interface {
	OnClick()
}

func (m *SimpleMessenger) displayMessage(duration time.Duration, msgType, foreGround, backGround string, args ...any) {
	if duration == 0 {
		duration = time.Second * 3
	}

	var printArgs = make([]interface{}, 0, len(args))
	var onClickFunc OnClickFunc
	for _, arg := range args {
		switch arg := arg.(type) {
		case OnClickFunc:
			onClickFunc = arg
		case OnClicker:
			onClickFunc = arg.OnClick
		default:
			printArgs = append(printArgs, arg)
		}
	}

	var (
		stringMessage = fmt.Sprint(printArgs...)
		message       = document.Call("createElement", "div")
	)
	message.Set("className", "message")
	message.Get("style").Set("color", foreGround)
	message.Get("style").Set("background-color", backGround)

	var span = document.Call("createElement", "span")
	span.Set("className", "message-text")
	span.Set("innerHTML", stringMessage)

	message.Call("appendChild", span)
	m.Element.Call("appendChild", message)

	message.Get("classList").Call("add", "active")
	var (
		remainingTime = duration
		startTime     time.Time
		timer         *time.Timer
		onEnter       js.Func
		onLeave       js.Func
		onClick       js.Func
		stopped       = false
		releaseFuncs  = func() {
			if stopped {
				return
			}
			message.Get("classList").Call("remove", "active")
			stopped = true
			timer.Stop()
			go func() {
				<-time.After(time.Millisecond * 500)
				span.Call("removeEventListener", "click", onClick)
				message.Call("removeEventListener", "mouseenter", onEnter)
				message.Call("removeEventListener", "mouseleave", onLeave)
				message.Call("removeEventListener", "click", onClick)
				message.Call("remove")
				onEnter.Release()
				onLeave.Release()
				onClick.Release()
			}()
		}
	)
	onEnter = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		timer.Stop()
		remainingTime = duration - time.Since(startTime)
		return nil
	})
	onLeave = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if remainingTime > 0 {
			timer.Reset(remainingTime)
		} else {
			releaseFuncs()
		}
		return nil
	})
	onClick = js.FuncOf(func(this js.Value, args []js.Value) any {
		if onClickFunc != nil {
			onClickFunc()
		}
		releaseFuncs()
		return nil
	})

	startTime = time.Now()
	timer = time.AfterFunc(remainingTime, releaseFuncs)

	// Start the timer
	go func() {
		if stopped {
			return
		}
		<-timer.C
		releaseFuncs()
	}()

	message.Call("addEventListener", "mouseenter", onEnter)
	message.Call("addEventListener", "mouseleave", onLeave)
	message.Call("addEventListener", "click", onClick)
	span.Call("addEventListener", "click", onClick)

}
