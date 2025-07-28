package telnetsession

import (
	"errors"
	"text/template"
	"time"
)

type SessionBuilder struct {
	actions  []Action
	enter    string
	prompt   string
	exprUser string
	exprPass string
	timeout  time.Duration
	errors   []error
}

func NewBuilder() *SessionBuilder {
	return &SessionBuilder{timeout: -1}
}

func (s *SessionBuilder) WithTimeout(timeout time.Duration) *SessionBuilder {
	s.timeout = timeout
	return s
}

func (s *SessionBuilder) SetEnter(enter string) *SessionBuilder {
	s.enter = enter
	return s
}

func (s *SessionBuilder) SetPrompt(value string) *SessionBuilder {
	s.prompt = value
	return s
}

func (s *SessionBuilder) SetLoginExpr(username, password string) *SessionBuilder {
	s.exprUser = username
	s.exprPass = password
	return s
}

func (s *SessionBuilder) Expect(text string) *SessionBuilder {
	s.actions = append(s.actions, &ExpectAction{text: text, OnSuccessFunc: nil})
	return s
}

func (s *SessionBuilder) ExpectAndDo(text string, onSuccess OnSuccessFunc) *SessionBuilder {
	s.actions = append(s.actions, &ExpectAction{text: text, OnSuccessFunc: onSuccess})
	return s
}

func (s *SessionBuilder) Send(text string) *SessionBuilder {

	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: nil, prompt: s.prompt, onSuccessFunc: nil})
	return s
}

func (s *SessionBuilder) SendAndDo(text string, onSuccess OnSuccessFunc) *SessionBuilder {

	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: nil, prompt: s.prompt, onSuccessFunc: onSuccess})
	return s
}

func (s *SessionBuilder) SendTempl(text string, data map[string]any) *SessionBuilder {

	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: data, prompt: s.prompt, onSuccessFunc: nil})
	return s
}

func (s *SessionBuilder) SendTemplAndDo(text string, data map[string]any, onSuccess OnSuccessFunc) *SessionBuilder {

	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: data, prompt: s.prompt, onSuccessFunc: onSuccess})
	return s
}

func (s *SessionBuilder) Build() (*Session, error) {

	if len(s.errors) > 0 {
		return nil, errors.Join(s.errors...)
	}

	return &Session{
		Actions:  s.actions,
		Enter:    s.enter,
		Prompt:   s.prompt,
		ExprUser: s.exprUser,
		ExprPass: s.exprPass,
		Timeout:  s.timeout,
	}, nil
}
