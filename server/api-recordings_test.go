package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rokeller/zt-dl/test"
	"github.com/rokeller/zt-dl/zattoo"
)

func Test_recordingsApiController_listAll(t *testing.T) {
	tests := []struct {
		name       string
		resp       test.HttpResponse
		wantStatus int
		wantBody   []byte
	}{
		{
			name:       "Status/500",
			resp:       test.HttpResponse{StatusCode: 456},
			wantStatus: 500,
			wantBody: []byte(`{"err":"failed to playlist with status 456"}
`),
		},
		{
			name: "Status/200",
			resp: test.HttpResponse{
				StatusCode: 200,
				Body: test.MakeJson(map[string]any{
					"success": true,
					"recordings": []any{
						map[string]any{
							"id":    1234,
							"title": "Test",
						},
					},
				}),
			},
			wantStatus: 200,
			wantBody: []byte(`[{"id":1234,"program_id":0,"cid":"","image_url":"","partial":false,"level":"","title":"Test","episode_title":"","start":"0001-01-01T00:00:00Z","end":"0001-01-01T00:00:00Z"}]
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/zapi/v2/playlist":
					if r.Method == http.MethodGet {
						tt.resp.Respond(w)
						return
					}
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := zattoo.NewAccountWithSession(t, host, client)
			s := &server{
				a:   a,
				dlq: &downloadQueue{},
			}
			c := recordingsApiController{s}

			r, _ := http.NewRequest(http.MethodGet, "blah", nil)
			w := httptest.NewRecorder()
			c.listAll(w, r)
			if w.Result().StatusCode != tt.wantStatus {
				t.Errorf("response status got %d, want %d", w.Result().StatusCode, tt.wantStatus)
			}
			if w.Body.String() != string(tt.wantBody) {
				t.Errorf("response body got %q, want %q", w.Body.String(), string(tt.wantBody))
			}
		})
	}
}

func Test_recordingsApiController_enqueueDownloadDownload(t *testing.T) {
	tests := []struct {
		name               string
		recordingId        string
		requestContentType string
		requestBody        []byte
		wantStatus         int
		wantBody           []byte
		wantQueue          []toDownload
		wantEvents         []event
	}{
		{
			name:        "Status400/MalformedRecordingId",
			recordingId: "not-an-int",
			wantStatus:  400,
			wantBody: []byte(`{"code":"error_parsing_recordingId","err":"strconv.ParseInt: parsing \"not-an-int\": invalid syntax"}
`),
		},
		{
			name:               "Status400/MalformedBody",
			recordingId:        "1234",
			requestContentType: "application/json",
			requestBody:        nil,
			wantStatus:         400,
			wantBody: []byte(`{"code":"error_parsing_body","err":"missing form body"}
`),
		},
		{
			name:               "Status400/MissingFilename",
			recordingId:        "2345",
			requestContentType: "application/x-www-form-urlencoded",
			requestBody:        []byte("foo"),
			wantStatus:         400,
			wantBody: []byte(`{"code":"missing_filename"}
`),
		},
		{
			name:               "Status200",
			recordingId:        "3456",
			requestContentType: "application/x-www-form-urlencoded",
			requestBody:        []byte("filename=my-file.mp4"),
			wantStatus:         200,
			wantBody: []byte(`{"result":true}
`),
			wantQueue: []toDownload{
				{RecordingId: 3456, OutputPath: "/tmp/test/my-file.mp4"},
			},
			wantEvents: []event{
				{QueueUpdated: &eventQueueUpdated{Queue: []toDownload{
					{RecordingId: 3456, OutputPath: "/tmp/test/my-file.mp4"},
				}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				hub:    newHub(),
				outdir: "/tmp/test",
			}
			s.dlq = newDownloadQueue(s)
			c := recordingsApiController{s}
			var reqBody io.Reader
			if nil != tt.requestBody {
				reqBody = bytes.NewBuffer(tt.requestBody)
			}
			r, _ := http.NewRequest(http.MethodPost, "blah", reqBody)
			r = mux.SetURLVars(r, map[string]string{
				"recordingId": tt.recordingId,
			})
			r.Header.Add("content-type", tt.requestContentType)
			w := httptest.NewRecorder()
			c.enqueueDownload(w, r)

			if w.Result().StatusCode != tt.wantStatus {
				t.Errorf("response status got %d, want %d", w.Result().StatusCode, tt.wantStatus)
			}
			if w.Body.String() != string(tt.wantBody) {
				t.Errorf("response body got %q, want %q", w.Body.String(), string(tt.wantBody))
			}
			if len(c.dlq.q) > 0 {
				if !reflect.DeepEqual(c.dlq.q, tt.wantQueue) {
					t.Errorf("got queue %v, want %v", c.dlq.q, tt.wantQueue)
				}
			} else if len(tt.wantQueue) > 0 {
				t.Errorf("got queue %v, want %v", c.dlq.q, tt.wantQueue)
			}

			consumeEvents(t, s.hub.outbox, tt.wantEvents)
			ensureNoMoreEvents(t, s.hub.outbox)
		})
	}
}
