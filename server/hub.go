package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout = 10 * time.Second
)

// TODO: buffer some events (eventQueueUpdated, eventDownloadStarted, or even last of every event) so new websocket connections can receive those too

// wsHub maintains the set of active clients, broadcasts messages to the clients,
// and receives messages from the clients.
type wsHub struct {
	upgrader   websocket.Upgrader
	clients    map[*wsClient]struct{}
	outbox     chan serverEvent
	inbox      chan sourcedClientEvent
	register   chan *wsClient
	unregister chan *wsClient

	clientEventHandlers []clientEventHandler

	lastQueueUpdated    *eventQueueUpdated
	lastDownloadStarted *eventDownloadStarted
}

type wsClient struct {
	hub    *wsHub
	conn   *websocket.Conn
	outbox chan serverEvent
}

type sourcedClientEvent struct {
	client *wsClient
	event  clientEvent
}

func newHub() *wsHub {
	return &wsHub{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		outbox:     make(chan serverEvent, 32),
		inbox:      make(chan sourcedClientEvent, 1),
		register:   make(chan *wsClient, 8),
		unregister: make(chan *wsClient, 8),
		clients:    make(map[*wsClient]struct{}),
	}
}

func (h *wsHub) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		// New client subscribing to receive events.
		case client := <-h.register:
			h.clients[client] = struct{}{}
			if nil != h.lastQueueUpdated {
				client.outbox <- serverEvent{QueueUpdated: h.lastQueueUpdated}
			}
			if nil != h.lastDownloadStarted {
				client.outbox <- serverEvent{DownloadStarted: h.lastDownloadStarted}
			}

		// Client unsubscribing from events.
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.outbox)
			}

		// Event received in hub's outbox to broadcast to clients.
		case event := <-h.outbox:
			// Buffer some important events for new clients.
			if nil != event.QueueUpdated {
				h.lastQueueUpdated = event.QueueUpdated
			} else if nil != event.DownloadStarted {
				h.lastDownloadStarted = event.DownloadStarted
			}

			for client := range h.clients {
				select {
				case client.outbox <- event:

				default:
					close(client.outbox)
					delete(h.clients, client)
				}
			}

		// Event received from a client in hub's inbox.
		case event := <-h.inbox:
			for _, h := range h.clientEventHandlers {
				h.Handle(event)
			}
		}
	}
}

func (h *wsHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wupgrade := w
	if u, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
		// Unwrap writer from middleware because otherwise connection hijacking
		// needed to upgrade to WebSockets won't work.
		wupgrade = u.Unwrap()
	}

	conn, err := h.upgrader.Upgrade(wupgrade, r, nil)
	if nil != err {
		if _, ok := err.(websocket.HandshakeError); !ok {
			fmt.Fprintf(os.Stderr, "Failed to upgrade to websocket: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Handshake error for websocket: %v\n", err)
		}
		return
	}
	client := &wsClient{hub: h, conn: conn, outbox: make(chan serverEvent, 64)}
	client.hub.register <- client

	// Start the write pump that writes events out to the client.
	go client.writePump()
	// Start the read pump that reads events from the client.
	go client.readPump()
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The application
// ensures that there is at most one writer to a connection by executing all
// writes from this goroutine.
func (c *wsClient) writePump() {
	defer c.conn.Close()

	for e := range c.outbox {
		c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))

		if err := c.conn.WriteJSON(e); nil != err {
			fmt.Fprintf(os.Stderr, "Failed to write event to the websocket: %v\n", err)
			return
		}
	}

	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// readPump pumps messages from the client to the hub.
//
// A goroutine running readPump is started for each connection.
func (c *wsClient) readPump() {
	defer c.conn.Close()

	for {
		var e clientEvent
		err := c.conn.ReadJSON(&e)
		if nil != err {
			fmt.Fprintf(os.Stderr, "Failed to read event from the websocket: %v\n", err)
			break
		}
		c.hub.inbox <- sourcedClientEvent{
			client: c,
			event:  e,
		}
	}
}
