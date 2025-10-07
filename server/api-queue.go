package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type queueApiController struct {
	*server
}

func AddQueuesApis(s *server, api *mux.Router) {
	r := api.PathPrefix("/queues").Subrouter()

	c := queueApiController{s}
	r.Handle("/events", c.hub).Methods(http.MethodGet)
}
