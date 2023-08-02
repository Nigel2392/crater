package crater

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Nigel2392/crater/craterhttp"
	"github.com/Nigel2392/crater/logger"
	"github.com/Nigel2392/crater/messenger"
	"github.com/Nigel2392/jsext/v2"
	"github.com/Nigel2392/jsext/v2/jse"
	"github.com/Nigel2392/jsext/v2/state"
	"github.com/Nigel2392/jsext/v2/websocket"
	"github.com/Nigel2392/mux"
)

type OnClickFunc messenger.OnClickFunc

func (f OnClickFunc) OnClick() {
	f()
}

var application *app

type lastTemplate struct {
	name string
	fun  func(args ...interface{}) Marshaller
}

type app struct {
	*jse.Element     `jsc:"rootElement"`
	elementEmbedFunc func(ctx context.Context, page *jse.Element) *jse.Element `jsc:"-"`
	templates        map[string]func(args ...interface{}) Marshaller           `jsc:"-"`
	lastUsedTemplate *lastTemplate                                             `jsc:"-"`
	config           *Config                                                   `jsc:"-"`
	exit             chan error                                                `jsc:"-"`
	Mux              *mux.Mux                                                  `jsc:"-"`
	Loader           Loader                                                    `jsc:"-"`
	Logger           Logger                                                    `jsc:"-"`
	OnResponseError  func(error)                                               `jsc:"-"`
	Messenger        Messenger                                                 `jsc:"-"`
	Websocket        *websocket.WebSocket                                      `jsc:"-"`
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
		Mux:              mux.New(),
		Element:          (*jse.Element)(&c.RootElement),
		exit:             make(chan error),
		config:           c,
		Loader:           c.Loader,
		Messenger:        c.Messenger,
		Logger:           c.Logger,
		OnResponseError:  c.OnResponseError,
		elementEmbedFunc: c.EmbedFunc,
		templates:        c.Templates,
	}

	if c.InitialPageURL == "" {
		c.InitialPageURL = "/"
	}

	application.Mux.InvokeHandler(c.Flags.Has(F_CHANGE_PAGE_EACH_CLICK))
	application.Mux.FirstPage(c.InitialPageURL)
	if c.NotFoundHandler != nil {
		application.Mux.NotFoundHandler = makeHandleFunc(c.NotFoundHandler)
	}
}

type SockOpts struct {
	Protocols []string
	OnOpen    func(*websocket.WebSocket, websocket.MessageEvent)
	OnMessage func(*websocket.WebSocket, websocket.MessageEvent)
	OnClose   func(*websocket.WebSocket, jsext.Event)
	OnError   func(*websocket.WebSocket, jsext.Event)
}

func (o *SockOpts) OpenSock(url string) *websocket.WebSocket {
	var sock *websocket.WebSocket
	if o != nil || len(o.Protocols) > 0 {
		sock = websocket.New(url, o.Protocols...)
	} else {
		sock = websocket.New(url)
	}
	return sock
}

func (o *SockOpts) Apply(sock *websocket.WebSocket) {
	if o.OnOpen != nil {
		sock.OnOpen(o.OnOpen)
	}
	if o.OnMessage != nil {
		sock.OnMessage(o.OnMessage)
	}
	if o.OnClose != nil {
		sock.OnClose(o.OnClose)
	}
	if o.OnError != nil {
		sock.OnError(o.OnError)
	}
}

// Open a websocket for the application.
func OpenSock(url string, options *SockOpts) {
	checkApp()

	if application.Websocket == nil || !application.Websocket.IsOpen() {
		application.Websocket = options.OpenSock(url)
	}

	if options == nil {
		return
	}

	options.Apply(application.Websocket)
}

// Retrieve the application's path multiplexer.
func Mux() *mux.Mux {
	checkApp()
	return application.Mux
}

// Retrieve the application's root element.
func Canvas() *jse.Element {
	checkApp()
	return application.Element
}

// Socket returns the application's websocket.
func Socket() *websocket.WebSocket {
	checkApp()
	return application.Websocket
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

var (
	// global websocket connections
	socks    = make([]*websocket.WebSocket, 0)
	socksMut = new(sync.Mutex)
)

func makeHandleFunc(h PageFunc) mux.Handler {

	if h == nil {
		panic("HandleFunc cannot be nil")
	}

	// Initialization of the handler, if it is supported.
	//
	// This is useful for initializing data on the handler.
	if initter, ok := h.(Initter); ok {
		initter.Init()
	}

	// Templates for the handler.
	if templater, ok := h.(Templater); ok {
		for k, v := range templater.Templates() {
			SetTemplate(k, v)
		}
	}

	// Websocket for this specific handler.
	var ws *websocket.WebSocket

	return mux.NewHandler(func(v mux.Variables) {
		application.Element.InnerHTML("")

		// Close all open sockets if the flag is set.
		//
		// However, do not close the global application's websocket.
		if application.config.Flags.Has(F_CLOSE_SOCKS_EACH_PAGE) {
			socksMut.Lock()
			for _, sock := range socks {
				if sock != nil && sock.IsOpen() {
					sock.Close(1000)
				}
			}
			socks = make([]*websocket.WebSocket, 0)
			socksMut.Unlock()
			ws = nil
		}

		// If SockConfigurator is implemented, open a socket with the given options.
		//
		// This will run each time the page is visited && ws is nil.
		if wsOpts, ok := h.(SockConfigurator); ok && ws == nil {
			var url, sockOpts = wsOpts.SockOptions()
			ws = sockOpts.OpenSock(url)
			sockOpts.Apply(ws)
			if application.config.Flags.Has(F_CLOSE_SOCKS_EACH_PAGE) {
				socks = append(socks, ws)
			}
		}

		// Set up the page.
		var canvas *jse.Element = jse.Div("crater-canvas")
		var page = &Page{
			Canvas:    canvas,
			Variables: v,
			Context:   context.Background(),
			State:     state.New(canvas.MarshalJS()),
			Sock:      ws,
		}

		// Initialization functions which will run each time the page is visited.
		if preloader, ok := h.(Preloader); ok {
			preloader.Preload(page)
		}

		// Serve the page, this will render elements onto the canvas.
		h.Serve(page)

		// Embed if needed.
		//
		// Pass context of the page to support logic based embedding.
		if application.elementEmbedFunc != nil {
			canvas = application.elementEmbedFunc(page.Context, canvas)
		}

		// If the node is a body element we cannot replace it, so we will just append the canvas.
		if application.Element.Get("nodeName").String() == "BODY" {
			application.Element.InnerHTML("")
			application.Element.AppendChild(canvas)
		} else {
			// Replace the application's root element with the canvas.
			application.Element.Replace(canvas)
		}

		// After render functions which will run
		// each time the page is visited and the serve function returns.
		if page.AfterRender != nil {
			page.AfterRender(page)
		}
	})
}

// Handle a path with a page function.
//
// The page passed to this function will have acess to page.DecodeResponse and page.Response fields.
//
// The page function will be called when the path is visited.
func HandleEndpoint(path string, r craterhttp.RequestFunc, h PageFunc) {
	checkApp()
	LogDebugf("Adding handler for path: %s", path)
	Handle(path, ToPageFunc(func(p *Page) {
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
		h.Serve(p)
	}))
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
//
// This function will spin up a goroutine to log the message.
func LogErrorf(format string, v ...interface{}) {
	checkApp()
	LogError(fmt.Sprintf(format, v...))
}

// Log an info message in Sprintf format.
//
// This function will spin up a goroutine to log the message.
func LogInfof(format string, v ...interface{}) {
	checkApp()
	LogInfo(fmt.Sprintf(format, v...))
}

// Log a debug message in Sprintf format.
//
// This function will spin up a goroutine to log the message.
func LogDebugf(format string, v ...interface{}) {
	checkApp()
	LogDebug(fmt.Sprintf(format, v...))
}

// An error message to be shown to the user.
//
// This function will spin up a goroutine to log the message.
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
//
// This function will spin up a goroutine to log the message.
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
//
// This function will spin up a goroutine to log the message.
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
//
// This function will spin up a goroutine to log the message.
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

// WithEmbed sets the application's embed function.
//
// This can be used to embed the page element, useful for navbars, footers etc.
func WithEmbed(f func(pageCtx context.Context, page *jse.Element) *jse.Element) {
	checkApp()
	application.elementEmbedFunc = f
}

// SetTemplate sets the application's template.
func SetTemplate(name string, f func(args ...interface{}) Marshaller) {
	checkApp()
	if application.templates == nil {
		application.templates = make(map[string]func(args ...interface{}) Marshaller)
	}
	application.templates[name] = f
}

// WithTemplate adds a template to the application.
//
// This function will panic if the template does not exist.
//
// The arguments passed to this function will be passed to the template function.
func WithTemplate(name string, args ...interface{}) Marshaller {
	checkApp()
	if application.templates == nil {
		application.templates = make(map[string]func(args ...interface{}) Marshaller)
	}
	// Some templates may be used more than once sequentially, we will cache the last used template.
	if application.lastUsedTemplate != nil && application.lastUsedTemplate.name == name {
		return application.lastUsedTemplate.fun(args...)
	}
	var v, ok = application.templates[name]
	if !ok || v == nil {
		panic(fmt.Sprintf("Template %s not found", name))
	}
	application.lastUsedTemplate = &lastTemplate{
		name: name,
		fun:  v,
	}
	return v(args...)
}

// WithoutTemplate removes a template from the application.
func WithoutTemplate(name string) {
	checkApp()
	if application.templates == nil {
		return
	}
	if application.lastUsedTemplate != nil && application.lastUsedTemplate.name == name {
		application.lastUsedTemplate = nil
	}
	delete(application.templates, name)
}

// WithNotFoundHandler sets the application's not found handler.
func WithNotFoundHandler(h PageFunc) {
	checkApp()
	application.Mux.NotFoundHandler = makeHandleFunc(h)
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
