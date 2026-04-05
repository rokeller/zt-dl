package server

import (
	"testing"
)

type testClientEventHandler struct{}

var _ clientEventHandler = testClientEventHandler{}

// Handle implements [clientEventHandler].
func (t testClientEventHandler) Handle(e sourcedClientEvent) {
	panic("unimplemented")
}

func Test_wsHub_registerHandler(t *testing.T) {
	type args struct {
		handler clientEventHandler
	}
	tests := []struct {
		name    string
		handler clientEventHandler
	}{
		{
			name:    "DeregistrationWorks",
			handler: testClientEventHandler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &wsHub{}
			dereg := h.registerHandler(tt.handler)
			if len(h.clientEventHandlers) != 1 {
				t.Errorf("wsHub.clientEventHandlers = %v, want only 1", h.clientEventHandlers)
			}
			dereg()
			if len(h.clientEventHandlers) != 0 {
				t.Errorf("wsHub.clientEventHandlers = %v, want 0", h.clientEventHandlers)
			}
		})
	}
}
