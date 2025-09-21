package zattoo

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"

	"golang.org/x/term"
)

type Account struct {
	email    string
	password string
	language string
	domain   string

	s *session
}

func NewAccount(email string) *Account {
	return &Account{
		email:    email,
		language: "en",
		domain:   "zattoo.com",
	}
}

func (a *Account) Login() error {
	if a.password == "" {
		a.readPassword()
	}

	jar, err := cookiejar.New(nil)
	if nil != err {
		return err
	}

	client := http.Client{
		Transport: &defaultHeadersRoundTripper{
			domain: a.domain,
			T:      http.DefaultTransport,
		},
		Jar: jar,
	}

	a.s = &session{
		client: &client,
	}

	if err := a.s.load(*a); nil != err {
		return err
	}

	return nil
}

func (a *Account) GetAllRecordings() error {
	err := a.s.getPlaylist(*a)
	if nil != err {
		return err
	}
	return nil
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
	fmt.Print("Please enter your password: ")
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if nil != err {
		return err
	}

	a.password = string(data)

	return nil
}
