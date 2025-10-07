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

// wsHub maintains the set of active clients and broadcasts messages to the clients.
type wsHub struct {
	upgrader   websocket.Upgrader
	clients    map[*wsClient]struct{}
	outbox     chan event
	register   chan *wsClient
	unregister chan *wsClient

	lastQueueUpdated    *eventQueueUpdated
	lastDownloadStarted *eventDownloadStarted
}

type wsClient struct {
	hub    *wsHub
	conn   *websocket.Conn
	outbox chan event
}

func newHub() *wsHub {
	return &wsHub{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		outbox:     make(chan event, 32),
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

		// new client subscribing to receive events.
		case client := <-h.register:
			h.clients[client] = struct{}{}
			if nil != h.lastQueueUpdated {
				client.outbox <- event{QueueUpdated: h.lastQueueUpdated}
			}
			if nil != h.lastDownloadStarted {
				client.outbox <- event{DownloadStarted: h.lastDownloadStarted}
			}

		// client unsubscribing from events.
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.outbox)
			}

		// event received in hub's outbox to broadcast to clients
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
	client := &wsClient{hub: h, conn: conn, outbox: make(chan event, 64)}
	client.hub.register <- client

	// Start the write pump that writes events out to the client.
	go client.writePump()
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
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
