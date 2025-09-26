package test

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
)

type HttpResponse struct {
	StatusCode int
	Body       []byte
}

func (r HttpResponse) Respond(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.StatusCode)
	if nil != r.Body {
		w.Write(r.Body)
	}
}

func MakeJson(body any) []byte {
	j, err := json.Marshal(body)
	if nil != err {
		panic(err)
	}

	return j
}

func NewHttpTestSetup(handlerFunc http.HandlerFunc) (*httptest.Server, *http.Client, string) {
	ts := httptest.NewTLSServer(http.HandlerFunc(handlerFunc))
	client := ts.Client()

	jar, err := cookiejar.New(nil)
	if nil != err {
		panic(err)
	}
	client.Jar = jar
	tsUrl, _ := url.Parse(ts.URL)

	return ts, client, tsUrl.Host
}
