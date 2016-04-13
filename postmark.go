package postmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
)

const (
	apiHost = "api.postmarkapp.com"

	serverTokenHeader  = "X-Postmark-Server-Token"
	accountTokenHeader = "X-Postmark-Account-Token"
)

// Postmark defines methods to interace with the Postmark API
type Postmark interface {
	SetClient(client *http.Client) Postmark

	// Templates returns a resource root object handling template interactions with Postmark
	Templates() Templates

	// Templates returns a resource root object handling template interactions with Postmark
	Emails() Emails
}

type postmark struct {
	serverToken  string
	accountToken string

	scheme string
	host   string

	client *http.Client
}

// Request is an general container for requests sent with Postmark
type Request struct {
	Method  string
	Path    string
	Params  url.Values
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
		scheme:       "https",
		host:         apiHost,
	}
}

func (p *postmark) Templates() Templates {
	return &templates{pm: p}
}

func (p *postmark) Emails() Emails {
	return &emails{pm: p}
}

func (p *postmark) Exec(ctx context.Context, req *Request) (*http.Response, error) {
	var payload io.Reader
	if req.Payload != nil {
		data, err := json.Marshal(req.Payload)
		if err != nil {
			return nil, err
		}

		payload = bytes.NewReader(data)
	}

	urlBuilder := url.URL{
		Scheme:   p.scheme,
		Host:     p.host,
		Path:     req.Path,
		RawQuery: req.Params.Encode(), // returns "" if nil
	}

	r, err := http.NewRequest(req.Method, urlBuilder.String(), payload)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")

	if req.AccountAuth {
		r.Header.Set(accountTokenHeader, p.accountToken)
	} else {
		r.Header.Set(serverTokenHeader, p.serverToken)
	}

	resp, err := p.httpclient().Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// for unsuccessful http status codes, unmarshal an error
	if resp.StatusCode/100 != 2 {
		pmerr := &Error{StatusCode: resp.StatusCode}

		// handle non-json responses
		if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			log.Printf("req: %+v", r)

			respData, _ := ioutil.ReadAll(resp.Body)
			pmerr.Message = string(respData)
			return resp, pmerr
		}

		if err := json.NewDecoder(resp.Body).Decode(pmerr); err != nil {
			return resp, err
		}
		if pmerr.IsError() {
			return resp, pmerr
		}
		return resp, fmt.Errorf("postmark call errored with status: %d", resp.StatusCode)
	}

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

// Error defines an error from the Postmark API
type Error struct {
	ErrorCode int
	Message   string

	// the HTTP status code of the response itself
	StatusCode int `json:"-"`
}

// IsError returns whether or not the response indicated an error
func (e *Error) IsError() bool {
	return e.ErrorCode != 0
}

func (e *Error) Error() string {
	if e.ErrorCode == 0 {
		return fmt.Sprintf("postmark HTTP error %d: %s", e.StatusCode, e.Message)
	}

	codeMeaning := "unknown"
	if meaning, ok := ErrorLookup[e.ErrorCode]; ok {
		codeMeaning = meaning
	}
	return fmt.Sprintf("postmark error %d %s: %s", e.ErrorCode, e.Message, codeMeaning)
}
