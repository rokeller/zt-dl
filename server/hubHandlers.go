package server

// clientEventHandler defines the interface for a handler for client events.
type clientEventHandler interface {
	Handle(e sourcedClientEvent)
}

// deregisterFunc is the function to remove a registered handler from the hub.
type deregisterFunc func()

// registerHandler registers a [clientEventHandler] for events sent to the hub
// by any client.
func (h *wsHub) registerHandler(handler clientEventHandler) deregisterFunc {
	dereg := func() {
		for i, hh := range h.clientEventHandlers {
			if hh == handler {
				h.clientEventHandlers = append(h.clientEventHandlers[:i], h.clientEventHandlers[i+1:]...)
			}
		}
	}
	h.clientEventHandlers = append(h.clientEventHandlers, handler)
	return dereg
}
