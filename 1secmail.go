package onesecmail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type mailboxAction int

const (
	getMessages mailboxAction = iota
	readMessage
	download
	genRandomMailbox
	getDomainList
)

func (m mailboxAction) String() string {
	return [...]string{
		"getMessages", "readMessage", "download", "genRandomMailbox", "getDomainList",
	}[m]
}

// Mail represents a mail in a 1secmail inbox.
type Mail struct {
	ID          int          `json:"id"`
	From        string       `json:"from"`
	Subject     string       `json:"subject"`
	Date        string       `json:"date"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Body        *string      `json:"body,omitempty"`
	TextBody    *string      `json:"textBody,omitempty"`
	HTMLBody    *string      `json:"htmlBody,omitempty"`
}

// Attachment represents an attachment in a 1secmail mail.
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
}

// HTTPClient is an interface that makes an HTTP request.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// API manages communication with the 1secmail's APIs that do not belong to a specific mailbox.
type API struct {
	client HTTPClient
}

// NewAPI returns a new API. If nil httpClient is provided, a new http.Client will be created.
func NewAPI(httpClient HTTPClient) API {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return API{client: httpClient}
}

func (a API) RandomAddresses(count int) ([]string, error) {
	req := a.constructRequest("GET", genRandomMailbox, map[string]string{
		"count": strconv.Itoa(count),
	})
	resp, err := a.client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode != 200) {
		return nil, fmt.Errorf("generate random mailbox failed: %w", err)
	}
	defer resp.Body.Close()

	var list []string
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}
	return list, nil
}

func (a API) Domains() ([]string, error) {
	req := a.constructRequest("GET", getDomainList, nil)
	resp, err := a.client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode != 200) {
		return nil, fmt.Errorf("get domain list failed: %w", err)
	}
	defer resp.Body.Close()

	var list []string
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}
	return list, nil
}

// Mailbox manages communication with the 1secmail's APIs that belong to a specific mailbox.
type Mailbox struct {
	Login  string
	Domain string
	API
}

// Address returns the email address of a Mailbox.
func (m Mailbox) Address() string {
	return fmt.Sprintf("%s@%s", m.Login, m.Domain)
}

// NewMailbox returns a new Mailbox. Use login and domain for the email
// handler that you intend to use. Login is the email username.
// If nil httpClient is provided, a new http.Client will be created.
func NewMailbox(login, domain string, httpClient HTTPClient) (Mailbox, error) {
	if _, ok := Domains[domain]; !ok {
		return Mailbox{}, fmt.Errorf("invalid domain: %s", domain)
	}
	return Mailbox{
		API:    NewAPI(httpClient),
		Domain: domain,
		Login:  login,
	}, nil
}

// NewMailboxWithAddress returns a new Mailbox. It accepts an email address
// that refers to a 1secmail mailbox. This is easier to use than NewMailbox
// if you already have an email address. If nil httpClient is provided, a
// new http.Client will be created.
func NewMailboxWithAddress(address string, httpClient HTTPClient) (Mailbox, error) {
	login, domain, ok := strings.Cut(address, "@")
	if !ok || login == "" || domain == "" {
		return Mailbox{}, fmt.Errorf("invalid email address: %s", address)
	}
	return NewMailbox(login, domain, httpClient)
}

// CheckInbox checks the inbox of a mailbox, and returns a list of mails.
func (m Mailbox) CheckInbox() ([]*Mail, error) {
	req := m.constructRequest("GET", getMessages, map[string]string{
		"login":  m.Login,
		"domain": m.Domain,
	})
	resp, err := m.client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode != 200) {
		return nil, fmt.Errorf("check inbox failed: %w, error code: %v", err, resp.StatusCode)
	}
	defer resp.Body.Close()

	var mails []*Mail
	if err := json.NewDecoder(resp.Body).Decode(&mails); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}
	return mails, nil
}

// ReadMessage retrieves a particular mail from the inbox of a mailbox.
func (m Mailbox) ReadMessage(messageID int) (*Mail, error) {
	req := m.constructRequest("GET", readMessage, map[string]string{
		"login":  m.Login,
		"domain": m.Domain,
		"id":     strconv.Itoa(messageID),
	})
	resp, err := m.client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode != 200) {
		return nil, fmt.Errorf("read message failed: %w", err)
	}
	defer resp.Body.Close()

	var mail *Mail
	if err := json.NewDecoder(resp.Body).Decode(&mail); err != nil {
		return nil, fmt.Errorf("decode JSON failed: %w", err)
	}

	return mail, nil
}

func (a API) constructRequest(method string, action mailboxAction, args map[string]string) *http.Request {
	const apiBase = "https://www.1secmail.com/api/v1/"

	req, _ := http.NewRequest(method, apiBase, nil)
	query := req.URL.Query()
	query.Add("action", fmt.Sprint(action))
	for k, v := range args {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()
	return req
}
