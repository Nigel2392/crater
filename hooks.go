package crater

// If the value returned is unspecified, it is safe to assume the value to be nil.
const (
	// SignalRun is sent when the application starts.
	SignalRun = "crater.Run"

	// SignalExit is sent when the application exits.
	//
	// The value sent is the error.
	SignalExit = "crater.Exit"

	// SignalPageChange is sent when a page is changed.
	SignalPageChange = "crater.PageChange"

	// SignalPageRendered is sent when a page is rendered.
	//
	// The value sent is the page.
	SignalPageRendered = "crater.PageRendered"

	// SignalSockConnected is sent when a websocket is connected.
	//
	// The value sent is the websocket.
	SignalSockConnected = "crater.SockConnected"

	// SignalClientResponse is sent when the client receives a response.
	//
	// The value sent is the client.
	SignalClientResponse = "crater.ClientResponse"

	// SignalHandlerAdded is sent when a handler is added.
	//
	// The value sent is the handler.
	SignalHandlerAdded = "crater.HandlerAdded"
)
