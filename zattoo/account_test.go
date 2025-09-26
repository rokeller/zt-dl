package zattoo

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/rokeller/zt-dl/test"
)

func TestDefaultHttpClientFactory(t *testing.T) {
	c := httpClientFactory()
	if c.Transport != http.DefaultTransport {
		t.Errorf("httpClientFactory(): Transport mismatch")
	}
}

func TestDefaultReadPassword(t *testing.T) {
	_, err := readPassword()
	if nil == err {
		t.Errorf("readPassword(): expected error, got nil")
	}
}

func TestNewAccount(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		email  string
		domain string
		want   *Account
	}{
		{
			name:   "Parameters passed",
			email:  "john.doe@foo.com",
			domain: "white-labeled.tv",
			want: &Account{
				email:    "john.doe@foo.com",
				domain:   "white-labeled.tv",
				language: "en",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAccount(tt.email, tt.domain)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_Login(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		email    string
		password string
		wantErr  bool

		tokenResp   test.HttpResponse
		sessionResp test.HttpResponse
		loginResp   test.HttpResponse
	}{
		{
			name:      "Error",
			email:     "login-error",
			password:  "abc123.",
			wantErr:   true,
			tokenResp: test.HttpResponse{StatusCode: 404},
		},
		{
			name:     "Success",
			email:    "successful-login",
			password: "zxcasd,",
			wantErr:  false,

			tokenResp: test.HttpResponse{
				StatusCode: 200,
				Body: test.MakeJson(map[string]any{
					"session_token": "the-session-token-from-token.json",
					"success":       true,
				}),
			},
			sessionResp: test.HttpResponse{
				StatusCode: 200,
				Body: test.MakeJson(map[string]any{
					"power_guide_hash": "the-power-guide-hash-from-session-hello",
					"success":          true,
				}),
			},
			loginResp: test.HttpResponse{
				StatusCode: 200,
				Body: test.MakeJson(map[string]any{
					"power_guide_hash": "the-power-guide-hash-from-login",
					"active":           true,
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origReadPassword := readPassword
			defer func() { readPassword = origReadPassword }()
			readPassword = func() (string, error) {
				return tt.password, nil
			}
			ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/token.json":
					if r.Method == http.MethodGet {
						tt.tokenResp.Respond(w)
						return
					}
				case "/zapi/v3/session/hello":
					if r.Method == http.MethodPost {
						if err := r.ParseForm(); nil != err {
							t.Errorf("failed to parse body: %v", err)
						}
						if r.FormValue("lang") != "en" {
							t.Errorf(`lang expected "en", got %q`, r.FormValue("lang"))
						}
						if r.FormValue("format") != "json" {
							t.Errorf(`format expected "json", got %q`, r.FormValue("format"))
						}
						if r.FormValue("app_version") != "3.2533.0" {
							t.Errorf(`app_version expected "3.2533.0", got %q`, r.FormValue("app_version"))
						}
						if r.FormValue("client_app_token") != "the-session-token-from-token.json" {
							t.Errorf(`client_app_token expected "the-session-token-from-token.json", got %q`, r.FormValue("client_app_token"))
						}
						tt.sessionResp.Respond(w)
						return
					}
				case "/zapi/v3/account/login":
					if r.Method == http.MethodPost {
						if err := r.ParseForm(); nil != err {
							t.Errorf("failed to parse body: %v", err)
						}
						if r.FormValue("login") != "successful-login" {
							t.Errorf(`login expected "successful-login", got %q`, r.FormValue("login"))
						}
						if r.FormValue("password") != "zxcasd," {
							t.Errorf(`password expected "zxcasd,", got %q`, r.FormValue("password"))
						}
						if r.FormValue("remember") != "true" {
							t.Errorf(`remember expected "true", got %q`, r.FormValue("remember"))
						}
						if r.FormValue("format") != "json" {
							t.Errorf(`format expected "json", got %q`, r.FormValue("format"))
						}
						tt.loginResp.Respond(w)
						return
					}
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			}))
			defer ts.Close()
			client := ts.Client()
			origHttpClientFactory := httpClientFactory
			defer func() { httpClientFactory = origHttpClientFactory }()
			httpClientFactory = func() *http.Client { return client }
			tsUrl, _ := url.Parse(ts.URL)

			a := NewAccount(tt.email, tsUrl.Host)
			gotErr := a.Login()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Login() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Login() succeeded unexpectedly")
			}
		})
	}
}

func TestAccount_Login_ReadPasswordError(t *testing.T) {
	origReadPassword := readPassword
	defer func() { readPassword = origReadPassword }()
	readPassword = func() (string, error) {
		return "", errors.New("injected error")
	}

	a := NewAccount("user@foo.com", "test.com")
	gotErr := a.Login()
	if nil == gotErr {
		t.Errorf("Login() expected error, got nil")
	}
}

func TestAccount_GetAllRecordings(t *testing.T) {
	r := recording{
		Id:           123,
		ProgramId:    456,
		ChannelId:    "test",
		Level:        "hd",
		Title:        "A Test Tale",
		EpisodeTitle: "Unit Tests",
		Start:        time.Date(2025, 9, 26, 12, 13, 00, 0, time.UTC),
		End:          time.Date(2025, 9, 26, 14, 13, 00, 0, time.UTC),
	}
	tests := []struct {
		name    string // description of this test case
		want    []recording
		wantErr bool

		playlistResp test.HttpResponse
	}{
		{
			name:    "Failure",
			wantErr: true,
		},
		{
			name: "Success",
			want: []recording{r},

			playlistResp: test.HttpResponse{
				StatusCode: 200,
				Body: test.MakeJson(map[string]any{
					"success":    true,
					"recordings": []any{r},
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origReadPassword := readPassword
			defer func() { readPassword = origReadPassword }()
			readPassword = func() (string, error) {
				return "blah", nil
			}
			ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/zapi/v2/playlist":
					if r.Method == http.MethodGet {
						tt.playlistResp.Respond(w)
						return
					}
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			}))
			defer ts.Close()
			client := ts.Client()
			origHttpClientFactory := httpClientFactory
			defer func() { httpClientFactory = origHttpClientFactory }()
			httpClientFactory = func() *http.Client { return client }
			tsUrl, _ := url.Parse(ts.URL)

			a := NewAccount("user@test.com", tsUrl.Host)
			a.s = &session{client: client}
			got, gotErr := a.GetAllRecordings()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetAllRecordings() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetAllRecordings() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllRecordings() = %v, want %v", got, tt.want)
			}
		})
	}
}
