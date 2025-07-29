package telnetsession

import (
	"errors"
	"text/template"
	"time"
)

// SessionBuilder provides a fluent interface for building Session configurations
type SessionBuilder struct {
	actions      []Action
	enter        string
	prompt       string
	exprUser     string
	exprPass     string
	timeout      time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	errors       []error
}

// NewBuilder creates a new SessionBuilder instance with default settings
func NewBuilder() *SessionBuilder {
	return &SessionBuilder{timeout: 0, readTimeout: 0, writeTimeout: 0, enter: "\n"}
}

// WithTimeout sets the connection timeout for the session
func (s *SessionBuilder) WithTimeout(timeout time.Duration) *SessionBuilder {
	s.timeout = timeout
	return s
}

// WithReadTimeout sets the read timeout for individual read operations
func (s *SessionBuilder) WithReadTimeout(timeout time.Duration) *SessionBuilder {
	s.readTimeout = timeout
	return s
}

// WithWriteTimeout sets the write timeout for individual write operations
func (s *SessionBuilder) WithWriteTimeout(timeout time.Duration) *SessionBuilder {
	s.writeTimeout = timeout
	return s
}

// SetEnter sets the line ending character(s) to use when sending commands
func (s *SessionBuilder) SetEnter(enter string) *SessionBuilder {
	s.enter = enter
	return s
}

// SetPrompt sets the prompt character to wait for after sending commands
func (s *SessionBuilder) SetPrompt(value string) *SessionBuilder {
	s.prompt = value
	return s
}

// SetLoginExpr sets the text patterns to expect for username and password prompts
func (s *SessionBuilder) SetLoginExpr(username, password string) *SessionBuilder {
	s.exprUser = username
	s.exprPass = password
	return s
}

// Expect adds an action that waits for the specified text to appear
func (s *SessionBuilder) Expect(text string) *SessionBuilder {
	s.actions = append(s.actions, &ExpectAction{text: text, OnSuccessFunc: nil})
	return s
}

// ExpectAndDo adds an action that waits for the specified text and executes a callback on success
func (s *SessionBuilder) ExpectAndDo(text string, onSuccess OnSuccessFunc) *SessionBuilder {
	s.actions = append(s.actions, &ExpectAction{text: text, OnSuccessFunc: onSuccess})
	return s
}

// Send adds an action that sends the specified text to the remote device
func (s *SessionBuilder) Send(text string) *SessionBuilder {
	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: nil, prompt: s.prompt, onSuccessFunc: nil})
	return s
}

// SendAndDo adds an action that sends text and executes a callback on success
func (s *SessionBuilder) SendAndDo(text string, onSuccess OnSuccessFunc) *SessionBuilder {
	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: nil, prompt: s.prompt, onSuccessFunc: onSuccess})
	return s
}

// SendTempl adds an action that sends templated text with the provided data
func (s *SessionBuilder) SendTempl(text string, data map[string]any) *SessionBuilder {
	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: data, prompt: s.prompt, onSuccessFunc: nil})
	return s
}

// SendTemplAndDo adds an action that sends templated text and executes a callback on success
func (s *SessionBuilder) SendTemplAndDo(text string, data map[string]any, onSuccess OnSuccessFunc) *SessionBuilder {
	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: data, prompt: s.prompt, onSuccessFunc: onSuccess})
	return s
}

// Build creates a Session instance from the builder configuration
// Returns an error if any template parsing errors occurred during building
func (s *SessionBuilder) Build() (*Session, error) {
	if len(s.errors) > 0 {
		return nil, errors.Join(s.errors...)
	}

	return &Session{
		Actions:      s.actions,
		Enter:        s.enter,
		Prompt:       s.prompt,
		ExprUser:     s.exprUser,
		ExprPass:     s.exprPass,
		Timeout:      s.timeout,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
	}, nil
}
