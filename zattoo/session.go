package zattoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type session struct {
	client *http.Client

	sessionToken   string
	powerGuideHash string
}

type tokenResponse struct {
	SessionToken string `json:"session_token"`
	Success      bool   `json:"success"`
}

type sessionInfoResponse struct {
	Active         bool   `json:"active"`
	PowerGuideHash string `json:"power_guide_hash"`
	Success        bool   `json:"success"`
}

type loginResponse struct {
	Active         bool   `json:"active"`
	PowerGuideHash string `json:"power_guide_hash"`
}

func (s *session) load(a Account) error {
	if err := s.fetchSessionToken(a); nil != err {
		return err
	}
	if err := s.fetchSession(a); nil != err {
		return err
	}
	if err := s.login(a); nil != err {
		return err
	}

	return nil
}

func (s *session) fetchSessionToken(a Account) error {
	resp, err := s.client.Get(fmt.Sprintf("https://%s/token.json", a.domain))
	if nil != err {
		return err
	}

	defer resp.Body.Close()
	var res tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); nil != err {
		return err
	}

	if !res.Success {
		return errors.New("failed to fetch session token")
	}
	s.sessionToken = res.SessionToken

	return nil
}

func (s *session) fetchSession(a Account) error {
	uuid, err := uuid.NewRandom()
	if nil != err {
		return err
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", a.domain))
	if nil != err {
		return err
	}
	cookies := s.client.Jar.Cookies(u)
	cookies = append(cookies, &http.Cookie{
		Name:   "uuid",
		Value:  uuid.String(),
		Domain: a.domain,
	})
	s.client.Jar.SetCookies(u, cookies)

	data := url.Values{}
	data.Set("uuid", uuid.String())
	data.Set("lang", a.language)
	data.Set("format", "json")
	data.Set("app_version", "3.2533.0")
	data.Set("client_app_token", s.sessionToken)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://%s/zapi/v3/session/hello", a.domain),
		strings.NewReader(data.Encode()))
	if nil != err {
		return err
	}
	resp, err := s.client.Do(req)
	if nil != err {
		return err
	}

	defer resp.Body.Close()
	var res sessionInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); nil != err {
		return err
	}

	if !(res.Active || res.Success) {
		return errors.New("failed to initialize session (hello)")
	}
	s.powerGuideHash = res.PowerGuideHash

	return nil
}

func (s *session) login(a Account) error {
	data := url.Values{}
	data.Set("login", a.email)
	data.Set("password", a.password)
	data.Set("remember", "true")
	data.Set("format", "json")

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://%s/zapi/v3/account/login", a.domain),
		strings.NewReader(data.Encode()))
	if nil != err {
		return err
	}
	resp, err := s.client.Do(req)
	if nil != err {
		return err
	}

	defer resp.Body.Close()
	var res loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); nil != err {
		return err
	}

	if !res.Active {
		return errors.New("failed to login")
	}

	return nil
}
