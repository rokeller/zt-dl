package server

// clientEventHandler defines the interface for a handler for client events.
type clientEventHandler interface {
	Handle(e sourcedClientEvent)
}
