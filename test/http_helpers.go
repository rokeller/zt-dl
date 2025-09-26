package test

import (
	"encoding/json"
	"net/http"
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
