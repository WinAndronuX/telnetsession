package telnetsession

import (
	"errors"
	"regexp"
	"text/template"
	"time"
)

// SessionBuilder provides a fluent interface for building Session configurations
type SessionBuilder struct {
	actions      []Action
	enter        string
	prompt       *regexp.Regexp
	exprUser     *regexp.Regexp
	exprPass     *regexp.Regexp
	timeout      time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	morePattern  *regexp.Regexp
	moreResponse string
	errorPatterns []*regexp.Regexp
	initialActions []Action
	debug        bool
	errors       []error
}

// NewBuilder creates a new SessionBuilder instance with default settings
func NewBuilder() *SessionBuilder {
	return &SessionBuilder{timeout: 0, readTimeout: 0, writeTimeout: 0, enter: "\n", debug: false}
}

// WithDebug enables verbose logging for the session
func (s *SessionBuilder) WithDebug() *SessionBuilder {
	s.debug = true
	return s
}

// WithErrors adds regular expression patterns that, if found in the device output, will abort the session
func (s *SessionBuilder) WithErrors(patterns ...string) *SessionBuilder {
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			s.errors = append(s.errors, err)
			continue
		}
		s.errorPatterns = append(s.errorPatterns, re)
	}
	return s
}

// SendInitial adds an action to be executed immediately after connecting, before login
func (s *SessionBuilder) SendInitial(text string) *SessionBuilder {
	templ, err := template.New("").Parse(text)
	if err != nil {
		s.errors = append(s.errors, err)
	}
	s.initialActions = append(s.initialActions, &SendAction{templ: templ, data: nil, prompt: nil, onSuccessFunc: nil})
	return s
}

// ExpectInitial adds an action to wait for a pattern before the login process starts
func (s *SessionBuilder) ExpectInitial(pattern string) *SessionBuilder {
	re, err := regexp.Compile(pattern)
	if err != nil {
		s.errors = append(s.errors, err)
		return s
	}
	s.initialActions = append(s.initialActions, &ExpectAction{pattern: re, OnSuccessFunc: nil})
	return s
}

// WithPagination sets a regular expression to detect pagination prompts and the response to send
func (s *SessionBuilder) WithPagination(pattern, response string) *SessionBuilder {
	re, err := regexp.Compile(pattern)
	if err != nil {
		s.errors = append(s.errors, err)
		return s
	}
	s.morePattern = re
	s.moreResponse = response
	return s
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

// SetPrompt sets the prompt regular expression to wait for after sending commands
func (s *SessionBuilder) SetPrompt(value string) *SessionBuilder {
	re, err := regexp.Compile(value)
	if err != nil {
		s.errors = append(s.errors, err)
		return s
	}
	s.prompt = re
	return s
}

// SetLoginExpr sets the regular expression patterns to expect for username and password prompts
func (s *SessionBuilder) SetLoginExpr(username, password string) *SessionBuilder {
	reUser, errUser := regexp.Compile(username)
	if errUser != nil {
		s.errors = append(s.errors, errUser)
	} else {
		s.exprUser = reUser
	}

	rePass, errPass := regexp.Compile(password)
	if errPass != nil {
		s.errors = append(s.errors, errPass)
	} else {
		s.exprPass = rePass
	}

	return s
}

// Expect adds an action that waits for the specified regular expression to appear
func (s *SessionBuilder) Expect(pattern string) *SessionBuilder {
	re, err := regexp.Compile(pattern)
	if err != nil {
		s.errors = append(s.errors, err)
		return s
	}
	s.actions = append(s.actions, &ExpectAction{pattern: re, OnSuccessFunc: nil})
	return s
}

// ExpectAndDo adds an action that waits for the specified regular expression and executes a callback on success
func (s *SessionBuilder) ExpectAndDo(pattern string, onSuccess OnSuccessFunc) *SessionBuilder {
	re, err := regexp.Compile(pattern)
	if err != nil {
		s.errors = append(s.errors, err)
		return s
	}
	s.actions = append(s.actions, &ExpectAction{pattern: re, OnSuccessFunc: onSuccess})
	return s
}

// Enable adds actions to enter privileged EXEC mode (Cisco-style)
func (s *SessionBuilder) Enable(password string, challengePattern ...string) *SessionBuilder {
	pattern := "Password:"
	if len(challengePattern) > 0 {
		pattern = challengePattern[0]
	}

	// 1. Enviar comando enable sin esperar prompt (porque el prompt NO vendrá, vendrá el reto)
	cmdTempl, _ := template.New("").Parse("enable")
	s.actions = append(s.actions, &SendAction{templ: cmdTempl, data: nil, prompt: nil, onSuccessFunc: nil})

	// 2. Esperar el reto de password (usualmente "Password:")
	s.Expect(pattern)

	// 3. Enviar la contraseña.
	passTempl, _ := template.New("").Parse(password)
	s.actions = append(s.actions, &SendAction{templ: passTempl, data: nil, prompt: nil, onSuccessFunc: nil})

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

func (s *SessionBuilder) Confirm(promptPattern, message string) *SessionBuilder {
	if len(s.actions) == 0 || s.actions[len(s.actions)-1].GetType() != ActionSend {
		s.errors = append(s.errors, errors.New("confirm action requires a pre send action"))
		return s
	}

	rePrompt, errPrompt := regexp.Compile(promptPattern)
	if errPrompt != nil {
		s.errors = append(s.errors, errPrompt)
		return s
	}

	s.actions[len(s.actions)-1].SetPrompt(rePrompt)

	templ, err := template.New("").Parse(message)
	if err != nil {
		s.errors = append(s.errors, err)
	}

	s.actions = append(s.actions, &SendAction{templ: templ, data: nil, prompt: s.prompt, onSuccessFunc: nil})

	return s
}

// Build creates a Session instance from the builder configuration
// Returns an error if any template parsing or regex compilation errors occurred during building
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
		MorePattern:  s.morePattern,
		MoreResponse: s.moreResponse,
		ErrorPatterns: s.errorPatterns,
		InitialActions: s.initialActions,
		Debug:        s.debug,
	}, nil
}
