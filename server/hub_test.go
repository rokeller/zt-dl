package server

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func consumeEvents(t *testing.T, events <-chan event, wantEvents []event) {
	t.Helper()
	if nil != wantEvents {
		for _, want := range wantEvents {
			consumeEvent(t, events, want)
		}
	}
}

func consumeEvent(t *testing.T, events <-chan event, want event) {
	t.Helper()
	got := <-events
	if !reflect.DeepEqual(got, want) {
		t.Errorf("consumed event; got %+v, want %+v", got, want)
	}
}

func ensureNoMoreEvents(t *testing.T, events <-chan event) {
	t.Helper()
	if len(events) != 0 {
		t.Errorf("events queue has %d events, want 0", len(events))
	}
}

func consumeTextMessageFromWebSocket(t *testing.T, conn *websocket.Conn, prefix string, timeout time.Duration) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	mt, r, err := conn.NextReader()
	if nil != err {
		t.Errorf("expected to read from websocket, got error %v", err)
	} else if mt != websocket.TextMessage {
		t.Errorf("expected text message from websocket, got type %d", mt)
	} else {
		body, _ := io.ReadAll(r)
		if !strings.HasPrefix(string(body), prefix) {
			t.Errorf("expected text message with prefix %q, got %q", prefix, body)
		}
	}
}

func ensureNoMoreMessagesFromWebSocket(t *testing.T, conn *websocket.Conn, d time.Duration) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(time.Millisecond))
	_, _, err := conn.NextReader()
	if nil == err {
		t.Error("expected read error, got nothing")
	}
}

func blockingSleep(t *testing.T, d time.Duration) {
	t.Helper()

	tt := time.NewTicker(d)
	<-tt.C
	tt.Stop()
}

func Test_newHub(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want *wsHub
	}{
		{
			name: "Success",
			want: &wsHub{
				upgrader: websocket.Upgrader{
					ReadBufferSize:  1024,
					WriteBufferSize: 1024,
				},
				outbox:     make(chan event, 32),
				register:   make(chan *wsClient, 8),
				unregister: make(chan *wsClient, 8),
				clients:    make(map[*wsClient]struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newHub()
			if !reflect.DeepEqual(got.upgrader, tt.want.upgrader) {
				t.Errorf("newHub() outbox = %+v, want %+v", got.upgrader, tt.want.upgrader)
			}
		})
	}
}

func Test_wsHub_run(t *testing.T) {
	tests := []struct {
		name           string // description of this test case
		ctxFactory     func(parent context.Context) (context.Context, context.CancelFunc)
		produce        func(t *testing.T, hub *wsHub, cancel context.CancelFunc)
		wantNumClients int
	}{
		{
			name: "CancelStopsRunning",
			ctxFactory: func(parent context.Context) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(parent)
				cancel()
				return ctx, func() {}
			},
		},
		{
			name: "RegisterAddsOneClient",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				hub.register <- &wsClient{}
				blockingSleep(t, time.Millisecond)
				cancel()
			},
			wantNumClients: 1,
		},
		{
			name: "RegisterSendsBufferedEvents",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				hub.lastQueueUpdated = &eventQueueUpdated{[]toDownload{{1, "a"}}}
				hub.lastDownloadStarted = &eventDownloadStarted{"abc"}
				c := &wsClient{outbox: make(chan event, 2)}
				hub.register <- c
				blockingSleep(t, time.Millisecond)
				consumeEvents(t, c.outbox, []event{
					{QueueUpdated: &eventQueueUpdated{Queue: []toDownload{{1, "a"}}}},
					{DownloadStarted: &eventDownloadStarted{"abc"}},
				})
				cancel()
			},
			wantNumClients: 1,
		},
		{
			name: "UnregisterIsNoopWhenNoClients",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				hub.unregister <- &wsClient{}
				blockingSleep(t, time.Millisecond)
				cancel()
			},
			wantNumClients: 0,
		},
		{
			name: "RegisterFollowedByUnregisterCleansUpProperly",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				c := &wsClient{outbox: make(chan event)}
				hub.register <- c
				blockingSleep(t, time.Millisecond)
				hub.unregister <- c
				blockingSleep(t, time.Millisecond)
				cancel()
			},
			wantNumClients: 0,
		},
		{
			name: "OutboxEvent/NoopForEmptyClients",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				hub.outbox <- event{}
				cancel()
			},
		},
		{
			name: "OutboxEvent/BuffersEvents",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				queueUpdated := &eventQueueUpdated{[]toDownload{{2, "b"}}}
				downloadStarted := &eventDownloadStarted{"def"}
				hub.outbox <- event{QueueUpdated: queueUpdated}
				hub.outbox <- event{DownloadStarted: downloadStarted}
				blockingSleep(t, time.Millisecond)
				cancel()

				if !reflect.DeepEqual(hub.lastQueueUpdated, queueUpdated) {
					t.Errorf("lastQueueUpdated is %v, want %v", hub.lastQueueUpdated, queueUpdated)
				}
				if !reflect.DeepEqual(hub.lastDownloadStarted, downloadStarted) {
					t.Errorf("lastDownloadStarted is %v, want %v", hub.lastQueueUpdated, downloadStarted)
				}
			},
		},
		{
			name: "OutboxEvent/ClosesAndDeletesClientOnFullOutbox",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				c := &wsClient{outbox: make(chan event)}
				hub.register <- c
				blockingSleep(t, time.Millisecond)
				hub.outbox <- event{}
				blockingSleep(t, time.Millisecond)
				cancel()
			},
			wantNumClients: 0,
		},
		{
			name: "OutboxEvent/ForwardsEventToClient",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				c := &wsClient{outbox: make(chan event, 1)}
				wantE := event{StateUpdated: &eventStateUpdated{State: "test", Reason: "unit test"}}
				hub.register <- c
				blockingSleep(t, time.Millisecond)
				hub.outbox <- wantE
				blockingSleep(t, time.Millisecond)
				gotE := <-c.outbox
				if !reflect.DeepEqual(gotE, wantE) {
					t.Errorf("broadcast event got %+v, want %+v", gotE, wantE)
				}
				cancel()
			},
			wantNumClients: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctxFactory == nil {
				tt.ctxFactory = context.WithCancel
			}
			ctx, cancel := tt.ctxFactory(t.Context())
			defer cancel()

			h := newHub()
			go h.run(ctx)
			if tt.produce != nil {
				tt.produce(t, h, cancel)
			}

			if len(h.clients) != tt.wantNumClients {
				t.Errorf("# clients got %d, want %d", len(h.clients), tt.wantNumClients)
			}
		})
	}
}

func Test_wsClient_writePump(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		setup  func(t *testing.T, c *wsClient)
		verify func(t *testing.T, c *wsClient)
	}{
		{
			name: "NoopForClosedOutbox",
			setup: func(t *testing.T, c *wsClient) {
				conn, _, err := websocket.DefaultDialer.Dial("wss://echo.websocket.org", http.Header{})
				if nil != err {
					t.Fatalf("failed to create websocket connection: %v", err)
				}
				c.conn = conn
				consumeTextMessageFromWebSocket(t, conn, "Request served by", time.Second)
				c.outbox = make(chan event)
				close(c.outbox)
				blockingSleep(t, time.Millisecond*10)

				ensureNoMoreMessagesFromWebSocket(t, conn, time.Millisecond*10)
			},
		},
		{
			name: "EventWrittenToWebSocket",
			setup: func(t *testing.T, c *wsClient) {
				conn, _, err := websocket.DefaultDialer.Dial("wss://echo.websocket.org", http.Header{})
				if nil != err {
					t.Fatalf("failed to create websocket connection: %v", err)
				}
				c.conn = conn
				consumeTextMessageFromWebSocket(t, conn, "Request served by", time.Second)
				c.outbox = make(chan event, 2)
				blockingSleep(t, time.Millisecond*10)
				c.outbox <- event{}
				c.outbox <- event{
					QueueUpdated: &eventQueueUpdated{Queue: []toDownload{}},
				}
				// close(c.outbox)
			},
			verify: func(t *testing.T, c *wsClient) {
				consumeTextMessageFromWebSocket(t, c.conn, "{}", time.Second)
				consumeTextMessageFromWebSocket(t, c.conn, `{"queueUpdated":{"queue":[]}}`, time.Second)
				ensureNoMoreMessagesFromWebSocket(t, c.conn, time.Millisecond*10)
				close(c.outbox)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &wsClient{}

			if tt.setup != nil {
				tt.setup(t, c)
			}
			if tt.verify != nil {
				go tt.verify(t, c)
			}
			c.writePump()
		})
	}
}
