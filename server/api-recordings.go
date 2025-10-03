package server

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rokeller/zt-dl/zattoo"
)

type recordings struct {
	a      *zattoo.Account
	dlq    *downloadQueue
	outdir string
}

func AddRecordingsApi(a *zattoo.Account, dlq *downloadQueue, outdir string, api *mux.Router) {
	r := api.PathPrefix("/recordings").Subrouter()

	rec := recordings{
		a:      a,
		dlq:    dlq,
		outdir: outdir,
	}
	r.HandleFunc("/", rec.listAll).Methods(http.MethodGet)
	r.HandleFunc("/{recordingId}/enqueue", rec.enqueue).Methods(http.MethodPost)
}

func (rec recordings) listAll(w http.ResponseWriter, r *http.Request) {
	recordings, err := rec.a.GetAllRecordings()

	w.Header().Add("content-type", "application/json")
	j := json.NewEncoder(w)
	if nil != err {
		w.WriteHeader(500)
		j.Encode(map[string]any{
			"err": err,
		})
		return
	}

	w.WriteHeader(200)
	j.Encode(recordings)
}

func (rec recordings) enqueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordingIdStr := vars["recordingId"]

	w.Header().Add("content-type", "application/json")
	j := json.NewEncoder(w)

	recordingId, err := strconv.ParseInt(recordingIdStr, 10, 64)
	if nil != err {
		w.WriteHeader(400)
		j.Encode(map[string]any{
			"code": "error_parsing_recordingId",
			"err":  err,
		})
		return
	}

	if err := r.ParseForm(); nil != err {
		w.WriteHeader(400)
		j.Encode(map[string]any{
			"code": "error_parsing_body",
			"err":  err,
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

	outputPath := path.Join(rec.outdir, filename)
	rec.dlq.Enqueue(recordingId, outputPath)

	w.WriteHeader(200)
	j.Encode(map[string]any{
		"result": true,
	})
}
