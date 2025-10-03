package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pongTimeout  = 60 * time.Second
	pingPeriod   = (pongTimeout * 9) / 10
)

type queueApi struct {
	upgrader websocket.Upgrader
}

func AdQueueApi(api *mux.Router) {
	r := api.PathPrefix("/queue").Subrouter()

	q := &queueApi{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	r.HandleFunc("/events", q.handleEventsWebsocket).Methods(http.MethodGet)
}

func (q *queueApi) handleEventsWebsocket(w http.ResponseWriter, r *http.Request) {
	wupgrade := w
	if u, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
		wupgrade = u.Unwrap()
		fmt.Printf("unwrapped ResponseWriter: used to be %T, now is %T\n", w, wupgrade)
	}
	fmt.Printf("ResponseWriter is %T\n", wupgrade)

	ws, err := q.upgrader.Upgrade(wupgrade, r, nil)
	if nil != err {
		if _, ok := err.(websocket.HandshakeError); !ok {
			fmt.Fprintf(os.Stderr, "failed to upgrade to websocket: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "handshake error for websocket: %v\n", err)
		}
		return
	}

	go q.writeEventsWebsocket(ws)
	go q.readEventsWebsocket(ws)
}

func (q *queueApi) readEventsWebsocket(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(pongTimeout))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})
	for {
		_, _, err := ws.ReadMessage()
		if nil != err {
			break
		}
	}
}

func (q *queueApi) writeEventsWebsocket(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()

	for {
		select {
		case e := <-events:
			ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := ws.WriteJSON(&e); nil != err {
				fmt.Fprintf(os.Stderr, "failed to write event JSON: %v\n", err)
			}

		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); nil != err {
				return
			}
		}
	}
}
