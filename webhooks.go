package postmark

import (
	"time"
)

// BounceWebhook defines the format of a webhook sent after an email bounced
// http://developer.postmarkapp.com/developer-bounce-webhook.html#data
type BounceWebhook struct {
	// you can use the ID to make different requests to the Bounce API.
	ID int64
	// the classification that Postmark assigned the bounce.
	Type        string
	TypeCode    int64
	Name        string
	Tag         string
	MessageID   string
	Description string
	Details     string
	// the email address that bounced.
	Email         string
	BouncedAt     time.Time
	DumpAvailable bool
	// lets you know if this bounce caused the email address to be deactivated.
	Inactive bool
	// lets you know if this address can be activated again.
	CanActivate bool
	Subject     string
}

// InboundWebhook defines the format of a webhook sent in response to an inbound email.
// http://developer.postmarkapp.com/developer-inbound-webhook.html#data
type InboundWebhook struct {
	FromName          string
	From              string
	FromFull          InboundEntity
	To                string
	ToFull            []InboundEntity
	Cc                string
	CcFull            []InboundEntity
	Bcc               string
	BccFull           []InboundEntity
	OriginalRecipient string
	Subject           string
	MessageID         string
	ReplyTo           string
	MailboxHash       string
	Date              string
	TextBody          string
	HTMLBody          string `json:"HtmlBody"`
	StrippedTextReply string
	Tag               string
	Headers           []InboundHeader
	Attachments       []InboundAttachment
}

// InboundEntity is an entity associated with an inbound webhook.
type InboundEntity struct {
	Email       string
	Name        string
	MailboxHash string
}

// InboundHeader is a header associated with an inbound webhook.
type InboundHeader struct {
	Name  string
	Value string
}

// InboundAttachment is an attachment associated with an inbound webhook.
type InboundAttachment struct {
	Name          string
	Content       string
	ContentType   string
	ContentLength int64
}

// OpenWebhook defines the format of a webhook sent in response to the opening of an email.
// http://developer.postmarkapp.com/developer-open-webhook.html#data
type OpenWebhook struct {
	FirstOpen   bool
	Client      OpenContext
	OS          OpenContext
	Platform    string
	UserAgent   string
	ReadSeconds float64
	Geo         OpenGeolocation
	MessageID   string
	ReceivedAt  time.Time
	Tag         string
	Recipient   string
}

// OpenContext defines the context within which an open occurred.
type OpenContext struct {
	Name    string
	Company string
	Family  string
}

// OpenGeolocation defines the geolocation for the opening of an email.
type OpenGeolocation struct {
	CountryISOCode string
	Country        string
	Region         string
	City           string
	Zip            string
	Coords         string
	IP             string
}
