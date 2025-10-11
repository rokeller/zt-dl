package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

func Test_logMiddlewareWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{
			name:   "StatusPropaged/200",
			status: 200,
		},
		{
			name:   "StatusPropaged/400",
			status: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			w := &logMiddlewareWriter{
				ResponseWriter: rec,
			}
			w.WriteHeader(tt.status)
			if w.status != tt.status {
				t.Errorf("logMiddlewareWriter stored status = %d, want %d", w.status, tt.status)
			}
		})
	}
}

func Test_logMiddlewareWriter_Unwrap(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsBaseResponseWriter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			w := &logMiddlewareWriter{
				ResponseWriter: rec,
			}
			if got := w.Unwrap(); got != rec {
				t.Errorf("logMiddlewareWriter.Unwrap() = %v, want %v", got, rec)
			}
		})
	}
}

func TestNewLogMiddleware(t *testing.T) {
	tests := []struct {
		name string
		want *logMiddleware
	}{
		{
			name: "Default",
			want: &logMiddleware{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLogMiddleware(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_logMiddleware_Func(t *testing.T) {
	tests := []struct {
		name string
		m    *logMiddleware
		want mux.MiddlewareFunc
	}{
		{
			name: "ForwardRequest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "TestOutput")
				handlerCalled = true
			}
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "https://localhost:8443", nil)
			if nil != err {
				t.Fatalf("failed to create new request: %v", err)
			}

			m := &logMiddleware{}
			h := m.Func()(http.HandlerFunc(handler))
			h.ServeHTTP(rec, req)
			if !handlerCalled {
				t.Error("expected handler to be called, but wasn't")
			}
		})
	}
}
