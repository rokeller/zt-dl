package server

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
)

type recordingsApiController struct {
	*server
}

func AddRecordingsApi(s *server, api *mux.Router) {
	r := api.PathPrefix("/recordings").Subrouter()

	c := recordingsApiController{s}
	r.HandleFunc("/", c.listAll).Methods(http.MethodGet)
	r.HandleFunc("/{recordingId}/enqueue", c.enqueueDownload).Methods(http.MethodPost)
	r.HandleFunc("/{recordingId}/dequeue", c.dequeueDownload).Methods(http.MethodPost)
}

func (c recordingsApiController) listAll(w http.ResponseWriter, r *http.Request) {
	recordings, err := c.a.GetAllRecordings()

	w.Header().Add("content-type", "application/json")
	j := json.NewEncoder(w)
	if nil != err {
		w.WriteHeader(500)
		j.Encode(map[string]any{
			"err": err.Error(),
		})
		return
	}

	w.WriteHeader(200)
	j.Encode(recordings)
}

func (c recordingsApiController) enqueueDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordingIdStr := vars["recordingId"]

	w.Header().Add("content-type", "application/json")
	j := json.NewEncoder(w)

	recordingId, err := strconv.ParseInt(recordingIdStr, 10, 64)
	if nil != err {
		w.WriteHeader(400)
		j.Encode(map[string]any{
			"code": "error_parsing_recordingId",
			"err":  err.Error(),
		})
		return
	}

	if c.dlq.InQueue(recordingId) {
		w.WriteHeader(409)
		j.Encode(map[string]any{
			"code": "recording_already_queued",
		})
		return
	}

	if err := r.ParseForm(); nil != err {
		w.WriteHeader(400)
		j.Encode(map[string]any{
			"code": "error_parsing_body",
			"err":  err.Error(),
		})
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		w.WriteHeader(400)
		j.Encode(map[string]any{
			"code": "missing_filename",
		})
		return
	}

	outputPath := path.Join(c.outdir, filename)
	c.dlq.Enqueue(recordingId, outputPath)

	w.WriteHeader(200)
	j.Encode(map[string]any{
		"result": true,
	})
}

func (c recordingsApiController) dequeueDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordingIdStr := vars["recordingId"]

	w.Header().Add("content-type", "application/json")
	j := json.NewEncoder(w)

	recordingId, err := strconv.ParseInt(recordingIdStr, 10, 64)
	if nil != err {
		w.WriteHeader(400)
		j.Encode(map[string]any{
			"code": "error_parsing_recordingId",
			"err":  err.Error(),
		})
		return
	}

	c.dlq.Dequeue(recordingId)

	w.WriteHeader(200)
	j.Encode(map[string]any{
		"result": true,
	})
}
