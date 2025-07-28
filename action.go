package telnetsession

import (
	"strings"
	"text/template"
)

type ActionType int
type OnSuccessFunc func(value string) error

const expect, send ActionType = 0, 1

type Action interface {
	GetType() ActionType
	GetText() (string, error)
	GetPrompt() string
	GetOnSuccessFunc() OnSuccessFunc
}

type ExpectAction struct {
	text          string
	OnSuccessFunc OnSuccessFunc
}

func (e *ExpectAction) GetType() ActionType {
	return expect
}

func (e *ExpectAction) GetText() (string, error) {
	return e.text, nil
}

func (e *ExpectAction) GetPrompt() string {
	return ""
}

func (e *ExpectAction) GetOnSuccessFunc() OnSuccessFunc {
	return e.OnSuccessFunc
}

type SendAction struct {
	templ         *template.Template
	prompt        string
	data          map[string]any
	onSuccessFunc OnSuccessFunc
}

func (s *SendAction) GetType() ActionType {
	return send
}

func (s *SendAction) GetText() (string, error) {

	var result strings.Builder
	if err := s.templ.Execute(&result, s.data); err != nil {
		return "", err
	}

	return result.String(), nil
}

func (s *SendAction) GetPrompt() string {
	return s.prompt
}

func (s *SendAction) GetOnSuccessFunc() OnSuccessFunc {
	return s.onSuccessFunc
}
