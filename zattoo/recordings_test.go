package zattoo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/rokeller/zt-dl/test"
)

func Test_session_getPlaylist(t *testing.T) {
	type fields struct {
		sessionToken   string
		powerGuideHash string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []recording
		wantErr bool
		resp    test.HttpResponse
	}{
		{
			name:    "Server Error",
			wantErr: true,
			resp:    test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Invalid JSON",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`malformed JSON`),
			},
		},
		{
			name:    "Unsuccessful",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":false}`),
			},
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
			a := Account{
				domain: host,
			}

			s := &session{
				client:         client,
				sessionToken:   tt.fields.sessionToken,
				powerGuideHash: tt.fields.powerGuideHash,
			}
			got, err := s.getPlaylist(a)
			if (err != nil) != tt.wantErr {
				t.Errorf("session.getPlaylist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("session.getPlaylist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_session_getRecording(t *testing.T) {
	type fields struct {
		sessionToken   string
		powerGuideHash string
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    stream
		wantErr bool
		resp    test.HttpResponse
	}{
		{
			name:    "Server Error",
			args:    args{id: 1},
			wantErr: true,
			resp:    test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Invalid JSON",
			args:    args{id: 2},
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`malformed JSON`),
			},
		},
		{
			name:    "Unsuccessful",
			args:    args{id: 3},
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":false}`),
			},
		},
		{
			name: "Successful",
			args: args{id: 4},
			want: stream{
				Url: "https://foo.bar/blah/blotz",
			},
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"stream":{"url":"https://foo.bar/blah/blotz"}}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == fmt.Sprintf("/zapi/watch/recording/%d", tt.args.id) &&
					r.Method == http.MethodPost {
					tt.resp.Respond(w)
					return
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := Account{
				domain: host,
			}

			s := &session{
				client:         client,
				sessionToken:   tt.fields.sessionToken,
				powerGuideHash: tt.fields.powerGuideHash,
			}
			got, err := s.getRecording(a, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("session.getRecording() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("session.getRecording() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_session_getProgramDetails(t *testing.T) {
	type fields struct {
		sessionToken   string
		powerGuideHash string
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    programDetails
		wantErr bool
		resp    test.HttpResponse
	}{
		{
			name:    "Server Error",
			args:    args{id: 1},
			wantErr: true,
			resp:    test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Invalid JSON",
			args:    args{id: 2},
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`malformed JSON`),
			},
		},
		{
			name:    "Unsuccessful",
			args:    args{id: 3},
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":false}`),
			},
		},
		{
			name: "Successful",
			args: args{id: 4},
			want: programDetails{
				ChannelName: "Test Channel",
				ChannelId:   "test_channel",
				Title:       "Test Show",
				Description: "This is a test show.",
				Year:        2025,
				Start:       123,
				End:         234,
			},
			resp: test.HttpResponse{
				StatusCode: 200,
				Body: test.MakeJson(map[string]any{
					"success": true,
					"programs": []any{
						map[string]any{
							"channel_name": "Test Channel",
							"cid":          "test_channel",
							"t":            "Test Show",
							"d":            "This is a test show.",
							"year":         2025,
							"s":            123,
							"e":            234,
						},
					},
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == fmt.Sprintf("/zapi/v2/cached/program/power_details/%s?program_ids=%d", tt.fields.powerGuideHash, tt.args.id) &&
					r.Method == http.MethodGet {
					tt.resp.Respond(w)
					return
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := Account{
				domain: host,
			}

			s := &session{
				client:         client,
				sessionToken:   tt.fields.sessionToken,
				powerGuideHash: tt.fields.powerGuideHash,
			}
			got, err := s.getProgramDetails(a, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("session.getProgramDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("session.getProgramDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
