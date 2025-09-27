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

func Test_session_fetchSessionToken(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		wantToken string
		resp      test.HttpResponse
	}{
		{
			name:    "Failure/Status500",
			wantErr: true,
			resp:    test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Failure/MalformedJSON",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`malformed JSON`),
			},
		},
		{
			name:    "Failure/NotSuccessful",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":false}`),
			},
		},
		{
			name:      "Success",
			wantToken: "test-session-token",
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"session_token":"test-session-token"}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/token.json":
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

			s := &session{client: client}
			if err := s.fetchSessionToken(a); (err != nil) != tt.wantErr {
				t.Errorf("session.fetchSessionToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if s.sessionToken != tt.wantToken {
				t.Errorf("session.sessionToken mismatch: want %q, got %q", tt.wantToken, s.sessionToken)
			}
		})
	}
}

func Test_session_fetchSession(t *testing.T) {
	type fields struct {
		sessionToken string
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		wantHash string
		resp     test.HttpResponse
	}{
		{
			name:    "Failure/Status500",
			wantErr: true,
			resp:    test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Failure/MalformedJSON",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`malformed JSON`),
			},
		},
		{
			name:    "Failure/NotSuccessful",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":false}`),
			},
		},
		{
			name: "Success",
			fields: fields{
				sessionToken: "token-success",
			},
			wantHash: "success-hash",
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"power_guide_hash":"success-hash"}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/zapi/v3/session/hello":
					if r.Method == http.MethodPost {
						if err := r.ParseForm(); nil != err {
							t.Errorf("failed to parse body: %v", err)
						}
						if r.FormValue("lang") != "test" {
							t.Errorf(`lang expected "test", got %q`, r.FormValue("lang"))
						}
						if r.FormValue("format") != "json" {
							t.Errorf(`format expected "json", got %q`, r.FormValue("format"))
						}
						if r.FormValue("app_version") != "3.2533.0" {
							t.Errorf(`app_version expected "3.2533.0", got %q`, r.FormValue("app_version"))
						}
						if r.FormValue("client_app_token") != tt.fields.sessionToken {
							t.Errorf(`client_app_token expected %q, got %q`, tt.fields.sessionToken, r.FormValue("client_app_token"))
						}

						tt.resp.Respond(w)
						return
					}
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := Account{
				domain:   host,
				language: "test",
			}

			s := &session{
				client:       wrapHttpClientTransport(client, host),
				sessionToken: tt.fields.sessionToken,
			}
			if err := s.fetchSession(a); (err != nil) != tt.wantErr {
				t.Errorf("session.fetchSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if s.powerGuideHash != tt.wantHash {
				t.Errorf("session.powerGuideHash mismatch: want %q, got %q", tt.wantHash, s.powerGuideHash)
			}
		})
	}
}

func Test_session_login(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		resp    test.HttpResponse
	}{
		{
			name:    "Failure/Status500",
			wantErr: true,
			resp:    test.HttpResponse{StatusCode: 500},
		},
		{
			name:    "Failure/MalformedJSON",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`malformed JSON`),
			},
		},
		{
			name:    "Failure/NotActive",
			wantErr: true,
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"active":false}`),
			},
		},
		{
			name: "Success",
			resp: test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"active":true,"power_guide_hash":"success-hash"}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/zapi/v3/account/login":
					if r.Method == http.MethodPost {
						if err := r.ParseForm(); nil != err {
							t.Errorf("failed to parse body: %v", err)
						}
						if r.FormValue("login") != "test@user.com" {
							t.Errorf(`login expected "test@user.com", got %q`, r.FormValue("login"))
						}
						if r.FormValue("password") != tt.name {
							t.Errorf(`password expected %q, got %q`, tt.name, r.FormValue("password"))
						}
						if r.FormValue("remember") != "true" {
							t.Errorf(`remember expected "true", got %q`, r.FormValue("remember"))
						}
						if r.FormValue("format") != "json" {
							t.Errorf(`format expected "json", got %q`, r.FormValue("format"))
						}

						tt.resp.Respond(w)
						return
					}
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := Account{
				domain:   host,
				language: "test",
				email:    "test@user.com",
				password: tt.name,
			}

			s := &session{
				client: wrapHttpClientTransport(client, host),
			}
			if err := s.login(a); (err != nil) != tt.wantErr {
				t.Errorf("session.login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
