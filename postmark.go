package postmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

const (
	pmRootEndpoint = "https://api.postmarkapp.com"

	pmServerTokenHeader  = "X-Postmark-Server-Token"
	pmAccountTokenHeader = "X-Postmark-Account-Token"
)

// Postmark defines methods to interace with the Postmark API
type Postmark interface {
	SetClient(client *http.Client) Postmark

	// Email sends a single email with custom content.
	// http://developer.postmarkapp.com/developer-api-email.html#send-email
	Email(ctx context.Context, email *Email) error

	// EmailWithTemplate sends an email with templated content.
	// http://developer.postmarkapp.com/developer-api-templates.html#email-with-template
	EmailWithTemplate(ctx context.Context, email *EmailWithTemplate) error
}

type postmark struct {
	serverToken  string
	accountToken string
	client       *http.Client
}

// Request is an general container for requests sent with Postmark
type Request struct {
	Method  string
	Path    string
	Payload interface{}
	Target  interface{}

	// Set this to true in order to use the account-wide API token
	AccountAuth bool
}

// New returns an initialized Postmark client
func New(serverToken, accountToken string) Postmark {
	return &postmark{
		serverToken:  serverToken,
		accountToken: accountToken,
	}
}

func (p *postmark) Exec(ctx context.Context, req *Request) (*http.Response, error) {
	data, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(req.Method, pmRootEndpoint+req.Path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")

	if req.AccountAuth {
		r.Header.Set("X-Postmark-Account-Token", p.accountToken)
	} else {
		r.Header.Set("X-Postmark-Server-Token", p.serverToken)
	}

	resp, err := p.httpclient().Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if req.Target != nil {
		if err := json.NewDecoder(resp.Body).Decode(req.Target); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (p *postmark) httpclient() *http.Client {
	if p.client != nil {
		return p.client
	}
	return http.DefaultClient
}

func (p *postmark) SetClient(client *http.Client) Postmark {
	p.client = client
	return p
}

// BaseEmail defines the fields common to all Postmark emails
type BaseEmail struct {
	From        string       `json:",omitempty"`
	To          string       `json:",omitempty"`
	Cc          string       `json:",omitempty"`
	Bcc         string       `json:",omitempty"`
	Tag         string       `json:",omitempty"`
	ReplyTo     string       `json:",omitempty"`
	Headers     []Header     `json:",omitempty"`
	TrackOpens  bool         `json:",omitempty"`
	Attachments []Attachment `json:",omitempty"`
}

// Header defines an email header within the Postmark API
type Header struct {
	Name  string
	Value string
}

// Attachment defines an email attachment within the Postmark API
type Attachment struct {
	Name        string
	Content     string
	ContentType string
}

// EmailResponse is the response from the postmark API after an email is sent.
// This can also be an error type for unsuccesful calls.
type EmailResponse struct {
	To          string
	SubmittedAt time.Time
	MessageID   string
	ErrorCode   int
	Message     string

	// the HTTP status code of the response itself
	StatusCode int `json:"-"`
}

// IsError returns whether or not the response indicated an error
func (e *EmailResponse) IsError() bool {
	return e.ErrorCode != 0
}

func (e *EmailResponse) Error() string {
	codeMeaning := "unknown"
	if meaning, ok := ErrorLookup[e.ErrorCode]; ok {
		codeMeaning = meaning
	}
	return fmt.Sprintf("postmark error %d %s: %s", e.ErrorCode, e.Message, codeMeaning)
}

// Email defines an email object within the Postmark API
type Email struct {
	BaseEmail

	Subject  string
	HTMLBody string `json:"HtmlBody"`
	TextBody string
}

func (p *postmark) Email(ctx context.Context, email *Email) error {
	er := new(EmailResponse)
	resp, err := p.Exec(ctx, &Request{
		Method:  "POST",
		Path:    "/email",
		Payload: email,
		Target:  er,
	})
	if err != nil {
		return err
	}

	er.StatusCode = resp.StatusCode

	if resp.StatusCode/100 != 2 {
		if er.IsError() {
			return er
		}
		return fmt.Errorf("postmark call errored with status: %d", resp.StatusCode)
	}

	if er.IsError() {
		return er
	}

	return nil
}

// EmailWithTemplate defines a templated email to the postmark API
type EmailWithTemplate struct {
	BaseEmail

	TemplateID    string `json:"TemplateId"`
	TemplateModel map[string]interface{}
	InlineCSS     bool `json:"InlineCss,omitempty"`
}

func (p *postmark) EmailWithTemplate(ctx context.Context, email *EmailWithTemplate) error {
	er := new(EmailResponse)
	resp, err := p.Exec(ctx, &Request{
		Method:  "POST",
		Path:    "/email/withTemplate/",
		Payload: email,
		Target:  er,
	})
	if err != nil {
		return err
	}

	er.StatusCode = resp.StatusCode

	if resp.StatusCode/100 != 2 {
		if er.IsError() {
			return er
		}
		return fmt.Errorf("postmark call errored with status: %d", resp.StatusCode)
	}

	if er.IsError() {
		return er
	}

	return nil
}
