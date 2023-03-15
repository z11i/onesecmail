package onesecmail_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/z11i/onesecmail"
)

type ClientMock struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return c.DoFunc(req)
}

func Test_NewMailbox(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		expErr bool
	}{
		{name: "valid domain", domain: "1secmail.com"},
		{name: "invalid domain", domain: "foobar.com", expErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mailbox, err := onesecmail.NewMailbox("", test.domain, nil)
			if (err == nil) != !test.expErr {
				t.Fatal("should not error")
			}
			if !test.expErr && mailbox == nil {
				t.Fatal("mailbox should not be nil")
			}
		})
	}
}

func Test_NewMailboxWithAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		expErr  bool
	}{
		{name: "valid address", address: "foo@1secmail.com"},
		{name: "invalid address", address: "foobar.com", expErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mailbox, err := onesecmail.NewMailboxWithAddress(test.address, nil)
			if (err == nil) != !test.expErr {
				t.Fatal("should not error")
			}
			if !test.expErr && mailbox == nil {
				t.Fatal("mailbox should not be nil")
			}
		})
	}
}

func Test_CheckInbox(t *testing.T) {
	tests := []struct {
		name     string
		respBody string
		respCode int
		expErr   string
		expLen   int
	}{
		{
			"valid response",
			`[{"id":639,"from":"someone@example.com","subject":"Some subject","date":"2018-06-08 14:33:55"},{"id":640,"from":"someoneelse@example.com","subject":"Other subject","date":"2018-06-08 14:40:55"}]`,
			200,
			"",
			2,
		},
		{
			"invalid json response",
			`[{"id":639,"from":"someone@example.com","subject":"Some subject","date":"2018]`,
			200,
			"decode JSON failed",
			0,
		},
		{
			"empty response",
			``,
			200,
			"decode JSON failed",
			0,
		},
		{
			"server error code",
			``,
			500,
			"check inbox failed",
			0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := ioutil.NopCloser(bytes.NewReader([]byte(test.respBody)))
			client := &ClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					var err error = nil
					if test.expErr != "" {
						err = errors.New(test.expErr)
					}
					return &http.Response{
						StatusCode: test.respCode, Body: r,
					}, err
				},
			}
			mailbox, err := onesecmail.NewMailbox("foo", "1secmail.org", client)
			if err != nil {
				t.Fatal("should not error")
			}
			gotMails, err := mailbox.CheckInbox()
			if (err == nil) != (test.expErr == "") {
				t.Fatal("should not error")
			}
			if err != nil && !strings.Contains(err.Error(), test.expErr) {
				t.Fatalf("error expected: %s, got: %s", test.expErr, err.Error())
			}
			if len(gotMails) != test.expLen {
				t.Fatal("len not expected")
			}
		})
	}
}

func Test_ReadMessage(t *testing.T) {
	tests := []struct {
		name     string
		respBody string
		respCode int
		respErr  string
		expErr   string
	}{
		{
			name:     "valid response",
			respBody: `{"id":639,"from":"email","subject":"subject","date":"2018-06-08 14:33:55","text":"text","html":"html"}`,
			respErr:  "",
		}, {
			name:     "error response",
			respBody: `{"id":639,"from":"email","subject":"subject","date":"2018-06-08 14:33:55","text":"text","html":"html"}`,
			respCode: 500,
			expErr:   "read message failed",
		}, {
			name:     "unknown http error",
			respBody: `{"id":639,"from":"email","subject":"subject","date":"2018-06-08 14:33:55","text":"text","html":"html"}`,
			respErr:  "unknown error",
			expErr:   "unknown error",
		}, {
			name:     "json decode error",
			respBody: `{"id":639,"from":"`,
			expErr:   "decode JSON failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := ioutil.NopCloser(bytes.NewReader([]byte(test.respBody)))
			client := &ClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					var err error = nil
					if test.respErr != "" {
						err = errors.New(test.respErr)
					}
					code := test.respCode
					if code == 0 {
						code = 200
					}
					return &http.Response{StatusCode: code, Body: r}, err
				},
			}
			mailbox, err := onesecmail.NewMailbox("foo", "1secmail.org", client)
			if err != nil {
				t.Fatal("should not error")
			}
			_, err = mailbox.ReadMessage(1)
			if (err == nil) != (test.expErr == "") {
				t.Fatal("should not error")
			}
			if err != nil && !strings.Contains(err.Error(), test.expErr) {
				t.Fatalf("error expected: %s, got: %s", test.expErr, err.Error())
			}
		})
	}

}
