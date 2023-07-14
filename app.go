package crater

import (
	"context"
	"fmt"
	"time"

	"github.com/Nigel2392/crater/craterhttp"
	"github.com/Nigel2392/crater/logger"
	"github.com/Nigel2392/crater/messenger"
	"github.com/Nigel2392/jsext/v2"
	"github.com/Nigel2392/jsext/v2/jse"
	"github.com/Nigel2392/jsext/v2/websocket"
	"github.com/Nigel2392/mux"
)

type OnClickFunc messenger.OnClickFunc

func (f OnClickFunc) OnClick() {
	f()
}

var application *app

type app struct {
	*jse.Element    `jsc:"rootElement"`
	config          *Config              `jsc:"-"`
	exit            chan error           `jsc:"-"`
	Mux             *mux.Mux             `jsc:"-"`
	Loader          Loader               `jsc:"-"`
	Logger          Logger               `jsc:"-"`
	OnResponseError func(error)          `jsc:"-"`
	Messenger       Messenger            `jsc:"-"`
	Websocket       *websocket.WebSocket `jsc:"-"`
}

// Helper function to check if the application has been initialized
func checkApp() {
	if application == nil {
		panic("Application not initialized, call crater.New() first")
	}
}

// Helper function to check if an error is nil, if not, log it and call the OnResponseError function
func checkErr(err error) bool {
	if err == nil {
		return false
	}
	LogError(fmt.Sprintf("Error: %s", err.Error()))
	if application.OnResponseError != nil {
		application.OnResponseError(err)
	} else {
		panic(err)
	}
	return true
}

// Initialize a new application
//
// The config parameter is optional, if nil, the default config will be used
func New(c *Config) {
	if application != nil {
		panic("Application already initialized")
	}
	if c == nil {
		c = &Config{
			RootElement: jsext.Body,
		}
	}
	application = &app{
		Mux:             mux.New(),
		Element:         (*jse.Element)(&c.RootElement),
		exit:            make(chan error),
		config:          c,
		Loader:          c.Loader,
		Messenger:       c.Messenger,
		Logger:          c.Logger,
		OnResponseError: c.OnResponseError,
	}
	application.Mux.InvokeHandler(c.Flags.Has(F_CHANGE_PAGE_EACH_CLICK))
	application.Mux.FirstPage(c.InitialPageURL)
	if c.NotFoundHandler != nil {
		application.Mux.NotFoundHandler = func(vars mux.Variables) {
			c.NotFoundHandler(&Page{
				Element:   application.Element,
				Variables: vars,
				Context:   context.Background(),
			})
		}
	}
}

type SockOpts struct {
	Protocols []string
	OnOpen    func(*websocket.WebSocket, websocket.MessageEvent)
	OnMessage func(*websocket.WebSocket, websocket.MessageEvent)
	OnClose   func(*websocket.WebSocket, jsext.Event)
	OnError   func(*websocket.WebSocket, jsext.Event)
}

// Open a websocket for the application.
func OpenSock(url string, options *SockOpts) {
	checkApp()

	if application.Websocket == nil {
		var sock *websocket.WebSocket
		if options != nil {
			sock = websocket.New(url, options.Protocols...)
		} else {
			sock = websocket.New(url)
		}
		application.Websocket = sock
	}

	if options == nil {
		return
	}

	if options.OnOpen != nil {
		application.Websocket.OnOpen(options.OnOpen)
	}
	if options.OnMessage != nil {
		application.Websocket.OnMessage(options.OnMessage)
	}
	if options.OnClose != nil {
		application.Websocket.OnClose(options.OnClose)
	}
	if options.OnError != nil {
		application.Websocket.OnError(options.OnError)
	}
}

// Retrieve the application's path multiplexer.
func Mux() *mux.Mux {
	checkApp()
	return application.Mux
}

// Retrieve the application's root element.
func Element() *jse.Element {
	checkApp()
	return application.Element
}

// Exit the application with an error.
func Exit(err error) {
	checkApp()
	application.exit <- err
}

// Run the application.
//
// This function will block until the application exits.
func Run() error {
	checkApp()
	application.Mux.ListenForChanges()
	return <-application.exit
}

// Change page to the given path.
func HandlePath(path string) {
	checkApp()
	application.Mux.HandlePath(path)
}

type Route interface {
	Handle(path string, h PageFunc) Route
}

// The route used to handle child routes, and handle pages.
type route struct {
	r *mux.Route
}

// Handle a path with a page function.
//
// The page function will be called when the path is visited.
//
// This function returns a route that can be used to add children.
func (r *route) Handle(path string, h PageFunc) Route {
	checkApp()
	var rt = r.r.Handle(path, makeHandleFunc(h))
	return &route{
		r: rt,
	}
}

// Handle a path with a page function.
//
// The page function will be called when the path is visited.
//
// This function returns a route that can be used to add children.
func Handle(path string, h PageFunc) Route {
	checkApp()
	var rt = application.Mux.Handle(path, makeHandleFunc(h))
	return &route{
		r: rt,
	}
}

func makeHandleFunc(h PageFunc) mux.HandleFunc {
	return func(v mux.Variables) {
		application.Element.InnerHTML("")
		var canvas = jse.Div()
		h(&Page{
			Element:   canvas,
			Variables: v,
			Context:   context.Background(),
		})
		application.Element.AppendChild(canvas)
	}
}

// Handle a path with a page function.
//
// The page passed to this function will have acess to page.DecodeResponse and page.Response fields.
//
// The page function will be called when the path is visited.
func HandleEndpoint(path string, r craterhttp.RequestFunc, h PageFunc) {
	checkApp()
	LogDebugf("Adding handler for path: %s", path)
	Handle(path, func(p *Page) {
		var (
			request *craterhttp.Request
			err     error
		)
		LogInfof("Handling endpoint: %s", path)
		ShowLoader()
		request, err = r(p.Variables)
		if checkErr(err) {
			HideLoader()
			return
		}
		LogDebugf("Making fetch request to %s", request.URL)
		p.Response, err = craterhttp.Fetch(request)
		if checkErr(err) {
			HideLoader()
			return
		}
		HideLoader()
		LogDebug("Received fetch response...")
		go h(p)
	})
}

// Show the application's loader.
func ShowLoader() {
	checkApp()
	if application.Loader != nil {
		LogDebug("Showing loader...")
		application.Loader.Show()
	}
}

// Hide the application's loader.
func HideLoader() {
	checkApp()
	if application.Loader != nil {
		LogDebug("Hiding loader...")
		application.Loader.Hide()
	}
}

// Set the application's log level.
func SetLogLevel(level logger.LogLevel) {
	checkApp()
	if application.Logger != nil {
		LogDebugf("Setting log level to: %s", level)
		application.Logger.Loglevel(level)
	}
}

// Log an error.
func LogError(s ...any) {
	checkApp()
	if application.Logger != nil {
		application.Logger.Error(s...)
	}
}

// Log an info message.
func LogInfo(s ...any) {
	checkApp()
	if application.Logger != nil {
		application.Logger.Info(s...)
	}
}

// Log a debug message.
func LogDebug(s ...any) {
	checkApp()
	if application.Logger != nil {
		application.Logger.Debug(s...)
	}
}

// Log an error in Sprintf format.
func LogErrorf(format string, v ...interface{}) {
	checkApp()
	LogError(fmt.Sprintf(format, v...))
}

// Log an info message in Sprintf format.
func LogInfof(format string, v ...interface{}) {
	checkApp()
	LogInfo(fmt.Sprintf(format, v...))
}

// Log a debug message in Sprintf format.
func LogDebugf(format string, v ...interface{}) {
	checkApp()
	LogDebug(fmt.Sprintf(format, v...))
}

// An error message to be shown to the user.
func ErrorMessage(d time.Duration, s ...any) {
	checkApp()
	if application.Messenger != nil {
		if application.config.Flags.Has(F_LOG_EACH_MESSAGE) {
			LogInfof("Logging error message: %s", s)
		}
		go application.Messenger.Error(d, s...)
	} else {
		LogErrorf("Application does not have a messenger. Message: %s", s)
	}
}

// A warning message to be shown to the user.
func WarningMessage(d time.Duration, s ...any) {
	checkApp()
	if application.Messenger != nil {
		if application.config.Flags.Has(F_LOG_EACH_MESSAGE) {
			LogInfof("Logging warning message: %s", s)
		}
		go application.Messenger.Warning(d, s...)
	} else {
		LogErrorf("Application does not have a messenger. Message: %s", s)
	}
}

// An info message to be shown to the user.
func InfoMessage(d time.Duration, s ...any) {
	checkApp()
	if application.Messenger != nil {
		if application.config.Flags.Has(F_LOG_EACH_MESSAGE) {
			LogInfof("Logging info message: %s", s)
		}
		go application.Messenger.Info(d, s...)
	} else {
		LogErrorf("Application does not have a messenger. Message: %s", s)
	}
}

// A success message to be shown to the user.
func SuccessMessage(d time.Duration, s ...any) {
	checkApp()
	if application.Messenger != nil {
		if application.config.Flags.Has(F_LOG_EACH_MESSAGE) {
			LogInfof("Logging success message: %s", s)
		}
		go application.Messenger.Success(d, s...)
	} else {
		LogErrorf("Application does not have a messenger. Message: %s", s)
	}
}

// WithLoader sets the application's loader.
func WithLoader(l Loader) {
	checkApp()
	application.Loader = l
}

// WithLogger sets the application's logger.
func WithLogger(l Logger) {
	checkApp()
	application.Logger = l
}

// WithMessenger sets the application's messenger.
func WithMessenger(m Messenger) {
	checkApp()
	application.Messenger = m
}

// WithNotFoundHandler sets the application's not found handler.
func WithNotFoundHandler(h PageFunc) {
	checkApp()
	application.Mux.NotFoundHandler = func(vars mux.Variables) {
		h(&Page{
			Element:   application.Element,
			Variables: vars,
			Context:   context.Background(),
		})
	}
}

// WithOnResponseError sets the application's OnResponseError function.
func WithOnResponseError(f func(error)) {
	checkApp()
	application.OnResponseError = f
}

// WithFlags sets the application's flags.
func WithFlags(flags CraterFlags) {
	checkApp()
	application.config.Flags = flags
}
