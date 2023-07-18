package crater

import (
	"context"

	"github.com/Nigel2392/crater/craterhttp"
	"github.com/Nigel2392/jsext/v2/jse"
	"github.com/Nigel2392/mux"
)

type PageFunc func(p *Page)

// Page represents a page in the application.
type Page struct {
	// The root element of the page
	//
	// This is the element which will act as a canvas for the page
	*jse.Element `jsc:"root"`

	// The response received from the server
	*craterhttp.Response `jsc:"-"`

	// The variables received from the server
	Variables mux.Variables `jsc:"variables"`

	// The context of the page
	//
	// This will be reset for each page render.
	Context context.Context `jsc:"-"`

	// A function which can be arbitrarily set, and will be called after the page is rendered.
	AfterRender func(p *Page) `jsc:"-"`
}

func (p *Page) Clear() {
	p.InnerHTML("")
}

func (p *Page) InnerElements(e ...*jse.Element) {
	p.AppendChild(e...)
}
