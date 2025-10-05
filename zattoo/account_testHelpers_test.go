package zattoo_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/rokeller/zt-dl/zattoo"
)

func TestNewAccountWithSession(t *testing.T) {
	c := &http.Client{
		Timeout: time.Second * 23,
	}
	tests := []struct {
		name   string // description of this test case
		host   string
		client *http.Client
	}{
		{
			name:   "PassThrough",
			host:   "abc",
			client: c,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zattoo.NewAccountWithSession(t, tt.host, tt.client)
		})
	}
}
