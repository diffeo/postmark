package postmark

import (
	"path"
	"time"

	"golang.org/x/net/context"
)

// Emails defines the functionality of the emails resource
type Emails interface {
	// Email sends a single email with custom content.
	// http://developer.postmarkape.com/developer-api-email.html#send-email
	Email(ctx context.Context, email *Email) (*EmailResponse, error)

	// EmailWithTemplate sends an email with templated content.
	// http://developer.postmarkape.com/developer-api-templates.html#email-with-template
	EmailWithTemplate(ctx context.Context, email *EmailWithTemplate) (*EmailResponse, error)
}

type emails struct {
	pm *postmark
}

var _ Emails = (*emails)(nil)

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
}

// Email defines an email object within the Postmark API
type Email struct {
	BaseEmail

	Subject  string
	HTMLBody string `json:"HtmlBody"`
	TextBody string
}

func (e *emails) Email(ctx context.Context, email *Email) (*EmailResponse, error) {
	er := new(EmailResponse)
	_, err := e.pm.Exec(ctx, &Request{
		Method:  "POST",
		Path:    "email",
		Payload: email,
		Target:  er,
	})
	if err != nil {
		return nil, err
	}
	return er, nil
}

// EmailWithTemplate defines a templated email to the postmark API
type EmailWithTemplate struct {
	BaseEmail

	TemplateID    string `json:"TemplateId"`
	TemplateModel map[string]interface{}
	InlineCSS     bool `json:"InlineCss,omitempty"`
}

func (e *emails) EmailWithTemplate(ctx context.Context, email *EmailWithTemplate) (*EmailResponse, error) {
	er := new(EmailResponse)
	_, err := e.pm.Exec(ctx, &Request{
		Method:  "POST",
		Path:    path.Join("email", "withTemplate"),
		Payload: email,
		Target:  er,
	})
	if err != nil {
		return nil, err
	}
	return er, nil
}
