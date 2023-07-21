# Crater

A simple and easy to use webassembly library which allows you to more easily write frontend code inside of Golang.

Best to be compiled with TinyGo, this can save you a lot of size on the resulting binary.

**Example file:**

```go
import (
	"fmt"
	"time"

	"github.com/Nigel2392/crater"
	"github.com/Nigel2392/crater/craterhttp"
	"github.com/Nigel2392/crater/decoder"
	"github.com/Nigel2392/crater/loader"
	"github.com/Nigel2392/crater/logger"
	"github.com/Nigel2392/crater/messenger"
	"github.com/Nigel2392/jsext/v2"
	"github.com/Nigel2392/jsext/v2/console"
	"github.com/Nigel2392/jsext/v2/websocket"
)

var app, _ = jsext.GetElementById("app")

// Main webassembly entry point
func main() {
	var loader, err = loader.NewLoader(
		loader.ID("loader"),                  // set the loader element's id to "loader"
		loader.ClassList([]string{"loader"}), // add the loader class to the loader element
		loader.QuerySelector("body"),         // the element where the attributes will be added to (Can be added also with loader.Element())
		loader.Element(jsext.Body),           // the element where the attributes will be added to (Can be added also with loader.QuerySelector())
	)
	if err != nil {
		console.Error(err)
		return
	}

	// Initialize a logger to use.
	var log = logger.StdLogger
	log.Loglevel(logger.Debug)

	// Initialize a new application
	crater.New(&crater.Config{
		RootElement:     app,
		Loader:          loader,
		Logger:          log,
		Flags:           crater.F_CHANGE_PAGE_EACH_CLICK | crater.F_LOG_EACH_MESSAGE,
		Messenger:       messenger.New(messenger.Styling{}),
		NotFoundHandler: NotFound,
	})

	// Open a websocket and handle the connection.
	crater.OpenSock("ws://127.0.0.1:8080/ws", &crater.SockOpts{
		OnOpen: func(w *websocket.WebSocket, event websocket.MessageEvent) {
			crater.LogInfo("Websocket opened")
			fmt.Println(event.Data().String())
			w.SendBytes([]byte("Hello World!"))
		},
		OnMessage: func(w *websocket.WebSocket, event websocket.MessageEvent) {
			crater.LogInfo("Received message")
			fmt.Println(event.Data().String())
		},
	})

	// Handle the root path.
	crater.Handle("/", crater.ToPageFunc(func(p *crater.Page) {
		p.Heading(1, "Hello World!")
	}))

	// Handle the hello path.
	crater.Handle("/hello/<<name>>", crater.ToPageFunc(func(p *crater.Page) {
		p.Heading(1, "Hello "+p.Variables.Get("name"))

		go func() {
			time.Sleep(time.Second * 5)
			crater.ErrorMessage(2*time.Second, "This", "is", "an", "error", "message", "number", i)
			crater.WarningMessage(2*time.Second, "This", "is", "a", "warning", "message", "number", i)
			crater.InfoMessage(2*time.Second, "This", "is", "an", "info", "message", "number", i)
			crater.SuccessMessage(2*time.Second, "This", "is", "a", "success", "message", "number", i, crater.OnClickFunc(func() {
				crater.HandlePath("/")
			}))
		}()
	}))

	// Easily handle a request to a server.
	crater.HandleEndpoint("/test", craterhttp.NewRequestFunc("GET", "https://jsonplaceholder.typicode.com/todos/1", nil), crater.ToPageFunc(func(p *crater.Page) {
		var respM = make(map[string]interface{})
		var err = p.DecodeResponse(decoder.JSONDecoder, &respM)
		if err != nil {
			console.Error(err.Error())
			return
		}
		crater.LogInfof("Received response: %s", respM)
		p.InnerHTML(fmt.Sprintf("%s", respM))
	}))

	// Run the application.
	err = crater.Run()
	if err != nil {
		console.Error(err.Error())
		return
	}
}


```
