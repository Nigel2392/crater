package crater

import (
	"syscall/js"

	"github.com/Nigel2392/crater/decoder"
	"github.com/Nigel2392/crater/logger"
	"github.com/Nigel2392/crater/messenger"
)

// A loader which will display when a page is loading
type Loader interface {
	Show()
	Hide()
}

type (
	// A logger which will be used to log messages
	Logger logger.Logger

	// The decoder interface for decoding responses with Response.DecodeResponse
	Decoder decoder.Decoder

	// A messenger which will display messages to the user
	Messenger messenger.Messenger
)

type Marshaller interface {
	MarshalJS() js.Value
}
