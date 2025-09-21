package zattoo

import (
	"fmt"
	"net/http"
)

type defaultHeadersRoundTripper struct {
	domain string
	T      http.RoundTripper
}

func (t *defaultHeadersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Referer", fmt.Sprintf("https://%s/client", t.domain))
	req.Header.Add("Origin", fmt.Sprintf("https://%s", t.domain))

	return t.T.RoundTrip(req)
}
