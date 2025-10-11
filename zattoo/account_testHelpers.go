package zattoo

import (
	"net/http"
	"testing"
)

func NewAccountWithSession(tb testing.TB, host string, client *http.Client) *Account {
	tb.Helper()
	a := NewAccount("test@user.com", host)
	s := &session{
		client: client,
	}
	a.s = s
	return a
}
