package crater

import (
	"context"
	"time"

	"github.com/Nigel2392/jsext/v2"
	"github.com/Nigel2392/jsext/v2/jse"
)

type CraterFlags uint32

const (
	// Change the page on each click of a link
	//
	// If not set, the page will only change when the URL changes.
	F_CHANGE_PAGE_EACH_CLICK CraterFlags = 1 << iota

	// Log each message sent with crater.InfoMessage(), crater.ErrorMessage() etc.
	F_LOG_EACH_MESSAGE

	// Close all websocket connections after switching pages
	//
	// If not set, websocket connections will be kept open
	// and must be closed manually, or they will be reused for the handler it was set on.
	F_CLOSE_SOCKS_EACH_PAGE

	// Append the canvas to the application's element instead of replacing it
	F_APPEND_CANVAS
)

// Check if the flag is set
func (f CraterFlags) Has(flag CraterFlags) bool {
	return f&flag != 0
}

type Config struct {
	// The function which will be called when a page is not found
	NotFoundHandler PageFunc `jsc:"-"`

	// The function which will be called when an error occurs in HandleEndpoint()
	OnResponseError func(error) `jsc:"-"`

	// The application's loader.
	Loader Loader `jsc:"-"`

	// The application's logger.
	Logger Logger `jsc:"-"`

	// The application's messenger.
	//
	// This will display messages to the user.
	Messenger Messenger `jsc:"-"`

	// The root element of the application.
	//
	// This will be passed to the Page struct, and will be used to render the page.
	RootElement jsext.Element `jsc:"-"`

	// Optional flags to change the behavior of the application.
	Flags CraterFlags `jsc:"-"`

	// The initial page URL.
	InitialPageURL string `jsc:"-"`

	// HttpClientTimeout is the timeout for the http client.
	HttpClientTimeout time.Duration `jsc:"-"`

	// Allows you to embed the canvas written to inside of another canvas.
	//
	// Useful for navbars, footers etc.
	//
	// The page should be embedded into the element returned by this function.
	EmbedFunc func(ctx context.Context, page *jse.Element) *jse.Element `jsc:"-"`

	// Templates which can be set, these can be used globally in the application.
	//
	// The arguments passed to the function are the arguments passed to the template.
	Templates map[string]func(args ...interface{}) Marshaller `jsc:"-"`
}
