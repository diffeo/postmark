package postmark

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nu7hatch/gouuid"

	"golang.org/x/net/context"
)

type (
	// Parents are used for more realistic validation/tests on postmark servers (using POSTMARK_API_TEST).
	mock          struct{ parent Postmark }
	mockEmails    struct{ parent *mock }
	mockTemplates struct{ parent *mock }

	templateInfo struct {
		*Template
		Keys []string
	}
)

var (
	// Mock is a mock Postmark client.
	Mock = &mock{New("POSTMARK_API_TEST", "")}
	eml  = &mockEmails{parent: Mock}
	tmpl = &mockTemplates{parent: Mock}

	tmplCt int64 = 5

	tmplInfo = map[int64]templateInfo{
		1: {
			Template: &Template{
				TemplateID: 1,
				Name:       "Template 1",
				Subject:    "Subject 1",
				HTMLBody:   "",
				TextBody:   "",
				Active:     true,
			},
			Keys: []string{"val1", "val2", "val3"},
		},
		2: {
			Template: &Template{
				TemplateID: 2,
				Name:       "Template 2",
				Subject:    "Subject 2",
				HTMLBody:   "TEST BODY",
				TextBody:   "TEST TEXT BODY",
				Active:     false,
			},
			Keys: []string{"val2", "val5", "val4"},
		},
		3: {
			Template: &Template{
				TemplateID: 3,
				Name:       "Template 3",
				Subject:    "Subject 3",
				HTMLBody:   "ANOTHER TEST BODY",
				TextBody:   "ANOTHER TEST TEXT BODY",
				Active:     true,
			},
			Keys: []string{"val6", "val9", "val12"},
		},
		4: {
			Template: &Template{
				TemplateID: 4,
				Name:       "Template 4",
				Subject:    "Subject 4",
				HTMLBody:   "some random text",
				TextBody:   "qwweqwweqweqwe",
				Active:     false,
			},
			Keys: []string{"val11", "val12", "val13"},
		},
		5: {
			Template: &Template{
				TemplateID: 5,
				Name:       "Template 5",
				Subject:    "Subject 5",
				HTMLBody:   "",
				TextBody:   "",
				Active:     true,
			},
			Keys: []string{"val2", "val1", "val3"},
		},
	}
	defaultTmpls = map[int]templateInfo{
		1: {
			Template: &Template{
				TemplateID: 1,
				Name:       "Template 1",
				Subject:    "Subject 1",
				HTMLBody:   "",
				TextBody:   "",
				Active:     true,
			},
			Keys: []string{"val1", "val2", "val3"},
		},
		2: {
			Template: &Template{
				TemplateID: 2,
				Name:       "Template 2",
				Subject:    "Subject 2",
				HTMLBody:   "TEST BODY",
				TextBody:   "TEST TEXT BODY",
				Active:     false,
			},
			Keys: []string{"val2", "val5", "val4"},
		},
		3: {
			Template: &Template{
				TemplateID: 3,
				Name:       "Template 3",
				Subject:    "Subject 3",
				HTMLBody:   "ANOTHER TEST BODY",
				TextBody:   "ANOTHER TEST TEXT BODY",
				Active:     true,
			},
			Keys: []string{"val6", "val9", "val12"},
		},
		4: {
			Template: &Template{
				TemplateID: 4,
				Name:       "Template 4",
				Subject:    "Subject 4",
				HTMLBody:   "some random text",
				TextBody:   "qwweqwweqweqwe",
				Active:     false,
			},
			Keys: []string{"val11", "val12", "val13"},
		},
		5: {
			Template: &Template{
				TemplateID: 5,
				Name:       "Template 5",
				Subject:    "Subject 5",
				HTMLBody:   "",
				TextBody:   "",
				Active:     true,
			},
			Keys: []string{"val2", "val1", "val3"},
		},
	}
)

// MockTemplateKeys retreives the keys for a given mock template.
func MockTemplateKeys(id int64) []string {
	return tmplInfo[id].Keys
}

func (m *mock) Emails() Emails {
	return eml
}

func (m *mock) Templates() Templates {
	return tmpl
}

func (m *mock) SetClient(client *http.Client) Postmark {
	m.parent = m.parent.SetClient(client)
	return m
}

func (m *mock) Reset() {
	tmplInfo = map[int64]templateInfo{}
	for k, v := range defaultTmpls {
		tmplInfo[int64(k)] = v
	}
}

func (m *mockEmails) real() Emails {
	return m.parent.parent.Emails()
}

func (m *mockEmails) Email(ctx context.Context, email *Email) (*EmailResponse, error) {
	return m.real().Email(ctx, email)
}

func (m *mockEmails) EmailWithTemplate(_ context.Context, email *EmailWithTemplate) (*EmailResponse, error) {
	id, err := strconv.Atoi(email.TemplateID)
	if err != nil {
		return nil, err
	}

	t, ok := tmplInfo[int64(id)]
	if !ok {
		return nil, &Error{
			ErrorCode: 1101,
			Message:   ErrorLookup[1101],
		}
	}

	if len(email.TemplateModel) == 0 {
		return nil, &Error{
			ErrorCode: 1109,
			Message:   ErrorLookup[1109],
		}
	}

	for _, k := range t.Keys {
		if _, ok := email.TemplateModel[k]; !ok {
			return nil, &Error{
				ErrorCode: 1120,
				Message:   ErrorLookup[1120],
			}
		}
	}

outerLoop:
	for k := range email.TemplateModel {
		for _, v := range t.Keys {
			if v == k {
				continue outerLoop
			}
		}
		return nil, &Error{
			ErrorCode: 1123,
			Message:   ErrorLookup[1123],
		}
	}

	guid, err := uuid.NewV4()

	return &EmailResponse{
		To:          email.To,
		SubmittedAt: time.Now(),
		MessageID:   guid.String(),
		ErrorCode:   0,
		Message:     "OK",
	}, err
}

func (m *mockTemplates) Get(_ context.Context, id int64) (*Template, error) {
	ret, ok := tmplInfo[id]
	if !ok {
		return nil, &Error{
			ErrorCode: 1101,
			Message:   ErrorLookup[1101],
		}
	}

	return ret.Template, nil
}

func (m *mockTemplates) Create(ctx context.Context, tmpl *Template) (*TemplateResp, error) {
	tmplCt++
	tmpl.TemplateID = tmplCt
	tmplInfo[tmplCt] = templateInfo{
		Template: tmpl,
	}

	return &TemplateResp{
		TemplateID: tmpl.TemplateID,
		Name:       tmpl.Name,
		Active:     tmpl.Active,
		Message:    "OK",
	}, nil
}

func (m *mockTemplates) Edit(_ context.Context, id int64, tmpl *Template) (*TemplateResp, error) {
	if _, ok := tmplInfo[tmpl.TemplateID]; !ok {
		return nil, &Error{
			ErrorCode: 1101,
			Message:   ErrorLookup[1101],
		}
	}
	tmplInfo[tmpl.TemplateID].Template = tmpl
	return &TemplateResp{
		TemplateID: tmpl.TemplateID,
		Name:       tmpl.Name,
		Active:     tmpl.Active,
		Message:    "OK",
	}, nil
}

func (m *mockTemplates) List(_ context.Context, count, offset int) (*TemplateList, error) {
	t := &TemplateList{}
	ct := 0

	if offset > tmplCt {
		return t, nil
	}

	for i := offset + 1; ct < count && i < 1000; i++ {
		tmpl, ok := tmplInfo[i]
		if !ok {
			continue
		}

		ct++
		t.Templates = append(t.Templates, tmpl)
	}

	t.TemplateCount = ct

	return t, nil
}

func (m *mockTemplates) Delete(_ context.Context, id int64) (*TemplateResp, error) {
	delete(tmplInfo, id)

	return &TemplateResp{
		Message: fmt.Sprintf("Template %v removed", id),
	}, nil
}

func (m *mockTemplates) Validate(_ context.Context, tmpl *TemplateValidation) (*TemplateValidationResp, error) {
	return &TemplateValidationResp{
		ValidationErrors: []TemplateValidationErr{
			Message: "Template validation is not supported in postmark.Mock.",
		},
	}, errors.New("Template validation not supported in postmark.Mock.")
}

func (m *mockTemplates) Email(ctx context.Context, email *EmailWithTemplate) (*EmailResponse, error) {
	return eml.EmailWithTemplate(ctx, email)
}
