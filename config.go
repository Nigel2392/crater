package crater

import (
	"github.com/Nigel2392/jsext/v2"
	"github.com/Nigel2392/jsext/v2/jse"
)

type CraterFlags uint32

const (
	F_CHANGE_PAGE_EACH_CLICK CraterFlags = 1 << iota
	F_LOG_EACH_MESSAGE
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

	// Allows you to embed the canvas written to inside of another canvas.
	//
	// Useful for navbars, footers etc.
	EmbedFunc func(root *jse.Element) `jsc:"-"`

	// Templates which can be set, these can be used globally in the application.
	Templates map[string]func() *jse.Element `jsc:"-"`
}
