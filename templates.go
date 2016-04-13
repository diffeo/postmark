package postmark

import (
	"net/url"
	"path"
	"strconv"

	"golang.org/x/net/context"
)

// Templates defines the functionality of the template resource
type Templates interface {
	// Get retrieves an individual template
	// http://developer.postmarkapp.com/developer-api-templates.html#get-template
	Get(ctx context.Context, id int64) (*Template, error)

	// Create creates a new template within Postmark
	// http://developer.postmarkapp.com/developer-api-templates.html#create-template
	Create(ctx context.Context, tmpl *Template) (*TemplateResp, error)

	// Edit modifies an existing template
	// http://developer.postmarkapp.com/developer-api-templates.html#edit-template
	Edit(ctx context.Context, id int64, tmpl *Template) (*TemplateResp, error)

	// List returns a list of all existing templates
	// http://developer.postmarkapp.com/developer-api-templates.html#template-list
	List(ctx context.Context, count, offset int) (*TemplateList, error)

	// Delete permanently deletes a template from Postmark
	// http://developer.postmarkapp.com/developer-api-templates.html#delete-template
	Delete(ctx context.Context, id int64) (*TemplateResp, error)

	// Validate allows a template's validity to be checked without sending the template
	// http://developer.postmarkapp.com/developer-api-templates.html#validate-template
	Validate(ctx context.Context, tmpl *TemplateValidation) (*TemplateValidationResp, error)

	// Email sends an email with the given template. This is a wrapper around the `EmailWithTemplate`
	// method that lives on the `Emails` resource, but lives here in order to match the Postmark docs.
	// http://developer.postmarkapp.com/developer-api-templates.html#email-with-template
	Email(ctx context.Context, email *EmailWithTemplate) (*EmailResponse, error)
}

type templates struct {
	pm *postmark
}

var _ Templates = (*templates)(nil)

// Template defines the template entities within postmark
type Template struct {
	TemplateID         int64 `json:"TemplateId"`
	Name               string
	Subject            string
	HTMLBody           string `json:"HtmlBody"`
	TextBody           string
	AssociatedServerID int64 `json:"AssociatedServerId"`
	Active             bool
}

func (t *templates) Get(ctx context.Context, id int64) (*Template, error) {
	tmpl := new(Template)
	_, err := t.pm.Exec(ctx, &Request{
		Method: "GET",
		Path:   path.Join("templates", i64toa(id)),
		Target: tmpl,
	})
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// TemplateResp defines the response after template manipulation
type TemplateResp struct {
	TemplateID int64 `json:"TemplateId"`
	Name       string
	Active     bool

	ErrorCode int
	Message   string
}

func (t *templates) Create(ctx context.Context, tmpl *Template) (*TemplateResp, error) {
	tmplResp := new(TemplateResp)
	_, err := t.pm.Exec(ctx, &Request{
		Method:  "POST",
		Path:    "templates",
		Payload: tmpl,
		Target:  tmplResp,
	})
	if err != nil {
		return nil, err
	}
	return tmplResp, nil
}

func (t *templates) Edit(ctx context.Context, id int64, tmpl *Template) (*TemplateResp, error) {
	tmplResp := new(TemplateResp)
	_, err := t.pm.Exec(ctx, &Request{
		Method:  "PUT",
		Path:    path.Join("templates", i64toa(id)),
		Payload: tmpl,
		Target:  tmplResp,
	})
	if err != nil {
		return nil, err
	}
	return tmplResp, nil
}

// TemplateList defines a list of template entities
type TemplateList struct {
	TemplateCount int64
	Templates     []*Template
}

func (t *templates) List(ctx context.Context, count, offset int) (*TemplateList, error) {
	tmplList := new(TemplateList)
	_, err := t.pm.Exec(ctx, &Request{
		Method: "GET",
		Path:   "templates",
		Params: url.Values{
			"count":  {strconv.Itoa(count)},
			"offset": {strconv.Itoa(offset)},
		},
		Target: tmplList,
	})
	if err != nil {
		return nil, err
	}
	return tmplList, nil
}

func (t *templates) Delete(ctx context.Context, id int64) (*TemplateResp, error) {
	tmplResp := new(TemplateResp)
	_, err := t.pm.Exec(ctx, &Request{
		Method: "DELETE",
		Path:   path.Join("templates", i64toa(id)),
		Target: tmplResp,
	})
	if err != nil {
		return nil, err
	}
	return tmplResp, nil
}

// TemplateValidation defines the structure of a template sent for validation with the API
type TemplateValidation struct {
	Subject                    string
	HTMLBody                   string `json:"HtmlBody"`
	TextBody                   string
	TestRenderModel            map[string]interface{}
	InlineCSSForHTMLTestRender bool `json:"InlineCssForHtmlTestRender"`
}

// TemplateValidationResp defines the result of a template validation call
type TemplateValidationResp struct {
	AllContentIsValid      bool
	ContentIsValid         bool
	ValidationErrors       []TemplateValidationErr
	RenderedContent        string
	SuggestedTemplateModel map[string]interface{}
}

// TemplateValidationErr defines an error that occured while validating a template
type TemplateValidationErr struct {
	Message           string
	Line              int
	CharacterPosition int
}

func (t *templates) Validate(ctx context.Context, tmpl *TemplateValidation) (*TemplateValidationResp, error) {
	tmplResp := new(TemplateValidationResp)
	_, err := t.pm.Exec(ctx, &Request{
		Method:  "POST",
		Path:    path.Join("templates", "validate"),
		Payload: tmpl,
		Target:  tmplResp,
	})
	if err != nil {
		return nil, err
	}
	return tmplResp, nil
}

func (t *templates) Email(ctx context.Context, email *EmailWithTemplate) (*EmailResponse, error) {
	return t.pm.Emails().EmailWithTemplate(ctx, email)
}

// helpers

// cleans up syntax for putting ids into URLs
func i64toa(i int64) string {
	return strconv.FormatInt(i, 10)
}
