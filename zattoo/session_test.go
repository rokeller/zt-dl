package zattoo

import (
	"net/http"
	"testing"

	"github.com/rokeller/zt-dl/test"
)

func Test_session_load(t *testing.T) {
	type fields struct {
		sessionToken   string
		powerGuideHash string
	}
	tests := []struct {
		name                  string
		fields                fields
		wantErr               bool
		fetchSessionTokenResp test.HttpResponse
		fetchSessionResp      test.HttpResponse
		loginResp             test.HttpResponse
	}{
		{
			name:                  "Failure/fetchSessionToken",
			wantErr:               true,
			fetchSessionTokenResp: test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Failure/fetchSession",
			wantErr: true,
			fetchSessionTokenResp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true}`),
			},
			fetchSessionResp: test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Failure/login",
			wantErr: true,
			fetchSessionTokenResp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true}`),
			},
			fetchSessionResp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"power_guide_hash":"1234"}`),
			},
			loginResp: test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Success",
			wantErr: false,
			fetchSessionTokenResp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true}`),
			},
			fetchSessionResp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"power_guide_hash":"1234"}`),
			},
			loginResp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"active":true}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/token.json":
					if r.Method == http.MethodGet {
						tt.fetchSessionTokenResp.Respond(w)
						return
					}
				case "/zapi/v3/session/hello":
					if r.Method == http.MethodPost {
						tt.fetchSessionResp.Respond(w)
						return
					}
				case "/zapi/v3/account/login":
					if r.Method == http.MethodPost {
						tt.loginResp.Respond(w)
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
			if err := s.load(a); (err != nil) != tt.wantErr {
				t.Errorf("session.load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
