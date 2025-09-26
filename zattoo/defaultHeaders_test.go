package zattoo

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type testRoundTripper struct{}

func (testRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	resp := rec.Result()
	for key, vals := range r.Header {
		for _, val := range vals {
			resp.Header.Add(key, val)
		}
	}

	return resp, nil
}

func Test_defaultHeadersRoundTripper_RoundTrip(t *testing.T) {
	req, err := http.NewRequest("TEST", "https://localhost/test", nil)
	if nil != err {
		t.Fatalf("failed to create request: %v", err)
	}
	type fields struct {
		domain string
		T      http.RoundTripper
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    http.Header
		wantErr bool
	}{
		{
			name: "Default Headers",
			fields: fields{
				domain: "test.com",
				T:      testRoundTripper{},
			},
			args: args{
				req: req,
			},
			want: map[string][]string{
				"Accept":           {"application/json"},
				"User-Agent":       {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"},
				"Content-Type":     {"application/x-www-form-urlencoded"},
				"X-Requested-With": {"XMLHttpRequest"},
				"Referer":          {"https://test.com/client"},
				"Origin":           {"https://test.com"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &defaultHeadersRoundTripper{
				domain: tt.fields.domain,
				T:      tt.fields.T,
			}

			got, err := tr.RoundTrip(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultHeadersRoundTripper.RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Header, tt.want) {
				t.Errorf("defaultHeadersRoundTripper.RoundTrip() = %v, want %v", got.Header, tt.want)
			}
		})
	}
}
