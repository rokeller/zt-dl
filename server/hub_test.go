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

type countingClientEventHandler struct {
	counter   *int
	increment int
}

var _ clientEventHandler = countingClientEventHandler{}

// Handle implements [clientEventHandler].
func (c countingClientEventHandler) Handle(e sourcedClientEvent) {
	if nil != c.counter {
		*c.counter += c.increment
	}
}

func consumeServerEvents(t *testing.T, events <-chan serverEvent, wantEvents []serverEvent) {
	t.Helper()
	if nil != wantEvents {
		for _, want := range wantEvents {
			consumeServerEvent(t, events, want)
		}
	}
}

func consumeServerEvent(t *testing.T, events <-chan serverEvent, want serverEvent) {
	t.Helper()
	got := <-events
	if !reflect.DeepEqual(got, want) {
		t.Errorf("consumed event; got %+v, want %+v", got, want)
	}
}

func ensureNoMoreServerEvents(t *testing.T, events <-chan serverEvent) {
	t.Helper()
	if len(events) != 0 {
		t.Errorf("events queue has %d events, want 0", len(events))
	}
}

func ensureNoMoreClientEvents(t *testing.T, events <-chan sourcedClientEvent) {
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

func sendJsonMessageToWebSocket(t *testing.T, conn *websocket.Conn, json any, timeout time.Duration) {
	t.Helper()
	err := conn.SetWriteDeadline(time.Now().Add(timeout))
	if nil != err {
		t.Errorf("expected no error, got %v", err)
	}
	err = conn.WriteJSON(json)
	if nil != err {
		t.Errorf("expected no error, got %v", err)
	}
}

func ensureNoMoreMessagesFromWebSocket(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(time.Millisecond))
	_, _, err := conn.NextReader()
	if nil == err {
		t.Error("expected read error, got nothing")
	}
}

func ensureSourceClientEventMatches(
	t *testing.T,
	actual sourcedClientEvent,
	wantC *wsClient,
	wantE clientEvent,
) {
	t.Helper()
	if actual.client != wantC {
		t.Error("expected client is different from actual")
	}
	if !reflect.DeepEqual(actual.event, wantE) {
		t.Errorf("got clientEvent %#v, want %#v", actual.event, wantE)
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
				outbox:     make(chan serverEvent, 32),
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
				c := &wsClient{outbox: make(chan serverEvent, 2)}
				hub.register <- c
				blockingSleep(t, time.Millisecond)
				consumeServerEvents(t, c.outbox, []serverEvent{
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
				c := &wsClient{outbox: make(chan serverEvent)}
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
				hub.outbox <- serverEvent{}
				cancel()
			},
		},
		{
			name: "OutboxEvent/BuffersEvents",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				queueUpdated := &eventQueueUpdated{[]toDownload{{2, "b"}}}
				downloadStarted := &eventDownloadStarted{"def"}
				hub.outbox <- serverEvent{QueueUpdated: queueUpdated}
				hub.outbox <- serverEvent{DownloadStarted: downloadStarted}
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
				c := &wsClient{outbox: make(chan serverEvent)}
				hub.register <- c
				blockingSleep(t, time.Millisecond)
				hub.outbox <- serverEvent{}
				blockingSleep(t, time.Millisecond)
				cancel()
			},
			wantNumClients: 0,
		},
		{
			name: "OutboxEvent/ForwardsEventToClient",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				c := &wsClient{outbox: make(chan serverEvent, 1)}
				wantE := serverEvent{StateUpdated: &eventStateUpdated{State: "test", Reason: "unit test"}}
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
		{
			name: "InboxEvent/NoHandlers",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				hub.inbox <- sourcedClientEvent{
					event: clientEvent{},
				}
			},
		},
		{
			name: "InboxEvent/WithHandlers",
			produce: func(t *testing.T, hub *wsHub, cancel context.CancelFunc) {
				weightedCalls := 0
				h1 := countingClientEventHandler{counter: &weightedCalls, increment: 1}
				h2 := countingClientEventHandler{counter: &weightedCalls, increment: 2}
				hub.addHandler <- h1
				hub.addHandler <- h2
				defer func() { hub.removeHandler <- h1 }()
				defer func() { hub.removeHandler <- h2 }()

				hub.inbox <- sourcedClientEvent{
					event: clientEvent{},
				}
				blockingSleep(t, time.Millisecond)

				if weightedCalls != 3 {
					t.Errorf("weightedCalls got %d, want 3", weightedCalls)
				}
			},
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
				c.outbox = make(chan serverEvent)
				close(c.outbox)
				blockingSleep(t, time.Millisecond*10)

				ensureNoMoreMessagesFromWebSocket(t, conn)
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
				c.outbox = make(chan serverEvent, 2)
				blockingSleep(t, time.Millisecond*10)
				c.outbox <- serverEvent{}
				c.outbox <- serverEvent{
					QueueUpdated: &eventQueueUpdated{Queue: []toDownload{}},
				}
			},
			verify: func(t *testing.T, c *wsClient) {
				consumeTextMessageFromWebSocket(t, c.conn, "{}", time.Second)
				consumeTextMessageFromWebSocket(t, c.conn, `{"queueUpdated":{"queue":[]}}`, time.Second)
				ensureNoMoreMessagesFromWebSocket(t, c.conn)
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

func Test_wsClient_readPump(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		setup  func(t *testing.T, c *wsClient)
		verify func(t *testing.T, c *wsClient)
	}{
		{
			name: "NoopForClosedConnection",
			setup: func(t *testing.T, c *wsClient) {
				conn, _, err := websocket.DefaultDialer.Dial("wss://echo.websocket.org", http.Header{})
				if nil != err {
					t.Fatalf("failed to create websocket connection: %v", err)
				}
				c.conn = conn
				consumeTextMessageFromWebSocket(t, conn, "Request served by", time.Second)
				conn.Close()
			},
		},
		{
			name: "EventReceived/ForwardedToHub",
			setup: func(t *testing.T, c *wsClient) {
				conn, _, err := websocket.DefaultDialer.Dial("wss://echo.websocket.org", http.Header{})
				if nil != err {
					t.Fatalf("failed to create websocket connection: %v", err)
				}
				c.conn = conn
				consumeTextMessageFromWebSocket(t, conn, "Request served by", time.Second)
				c.hub = &wsHub{
					inbox: make(chan sourcedClientEvent, 10),
				}
				blockingSleep(t, time.Millisecond*10)
			},
			verify: func(t *testing.T, c *wsClient) {
				sendJsonMessageToWebSocket(t, c.conn, clientEvent{}, time.Millisecond*10)
				sendJsonMessageToWebSocket(t, c.conn, clientEvent{
					Correlation: "my-test",
				}, time.Millisecond*10)
				sendJsonMessageToWebSocket(t, c.conn, clientEvent{
					Correlation: "test2",
					StreamsSelected: &eventStreamsSelected{
						SelectedStreams: []sourceStream{
							{
								Index: 123,
							},
						},
					},
				}, time.Millisecond*10)

				ensureSourceClientEventMatches(t, <-c.hub.inbox, c, clientEvent{})
				ensureSourceClientEventMatches(t, <-c.hub.inbox, c, clientEvent{
					Correlation: "my-test",
				})
				ensureSourceClientEventMatches(t, <-c.hub.inbox, c, clientEvent{
					Correlation: "test2",
					StreamsSelected: &eventStreamsSelected{
						SelectedStreams: []sourceStream{
							{
								Index: 123,
							},
						},
					},
				})
				ensureNoMoreClientEvents(t, c.hub.inbox)
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
			c.readPump()
		})
	}
}
