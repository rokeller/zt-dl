package zattoo

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"

	"golang.org/x/term"
)

var httpClientFactory func(transport http.RoundTripper) *http.Client = func(transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: transport,
	}
}

var readPassword func() (string, error) = func() (string, error) {
	if data, err := term.ReadPassword(int(os.Stdin.Fd())); nil != err {
		return "", err
	} else {
		return string(data), nil
	}
}

type Account struct {
	email    string
	password string
	language string
	domain   string

	s *session
}

func NewAccount(email, domain string) *Account {
	return &Account{
		email:    email,
		language: "en",
		domain:   domain,
	}
}

func (a *Account) Login() error {
	if a.password == "" {
		if err := a.readPassword(); nil != err {
			return err
		}
	}

	jar, err := cookiejar.New(nil)
	if nil != err {
		return fmt.Errorf("failed init cookie jar: %w", err)
	}

	client := httpClientFactory(
		&defaultHeadersRoundTripper{
			domain: a.domain,
			T:      http.DefaultTransport,
		})
	client.Jar = jar

	a.s = &session{
		client: client,
	}

	if err := a.s.load(*a); nil != err {
		return err
	}

	return nil
}

func (a *Account) GetAllRecordings() ([]recording, error) {
	return a.s.getPlaylist(*a)
}

func (a *Account) GetProgramDetails(id int64) (programDetails, error) {
	return a.s.getProgramDetails(*a, id)
}

func (a *Account) GetRecordingStreamUrl(id int64) (string, error) {
	stream, err := a.s.getRecording(*a, id)
	if nil != err {
		return "", err
	}
	return stream.Url, nil
}

func (a *Account) readPassword() error {
	fmt.Println("Please enter your password:")
	password, err := readPassword()

	if nil != err {
		return fmt.Errorf("failed to read password: %w", err)
	}
	a.password = password

	return nil
}
