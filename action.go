package telnetsession

import (
	"strings"
	"text/template"
)

// ActionType represents the type of action that can be performed in a telnet session
type ActionType int

// OnSuccessFunc is a callback function that gets executed when an action completes successfully
// The function receives the output value as a string and can return an error
type OnSuccessFunc func(value string) error

const (
	// ActionExpect represents an action that waits for specific text to appear
	ActionExpect ActionType = iota
	// ActionSend represents an action that sends text to the remote device
	ActionSend
)

// Action defines the interface that all session actions must implement
type Action interface {
	// GetType returns the type of this action
	GetType() ActionType
	// GetText returns the text associated with this action
	GetText() (string, error)
	// GetPrompt returns the prompt character to wait for after sending text
	GetPrompt() string
	// GetOnSuccessFunc returns the callback function to execute on success
	GetOnSuccessFunc() OnSuccessFunc
}

// ExpectAction represents an action that waits for specific text to appear in the session
type ExpectAction struct {
	text          string
	OnSuccessFunc OnSuccessFunc
}

// GetType returns ActionExpect for ExpectAction
func (e *ExpectAction) GetType() ActionType {
	return ActionExpect
}

// GetText returns the text to expect, always returns the text without error
func (e *ExpectAction) GetText() (string, error) {
	return e.text, nil
}

// GetPrompt returns an empty string as ExpectAction doesn't use prompts
func (e *ExpectAction) GetPrompt() string {
	return ""
}

// GetOnSuccessFunc returns the success callback function
func (e *ExpectAction) GetOnSuccessFunc() OnSuccessFunc {
	return e.OnSuccessFunc
}

// SendAction represents an action that sends templated text to the remote device
type SendAction struct {
	templ         *template.Template
	prompt        string
	data          map[string]any
	onSuccessFunc OnSuccessFunc
}

// GetType returns ActionSend for SendAction
func (s *SendAction) GetType() ActionType {
	return ActionSend
}

// GetText executes the template with the provided data and returns the resulting text
func (s *SendAction) GetText() (string, error) {
	var result strings.Builder
	if err := s.templ.Execute(&result, s.data); err != nil {
		return "", err
	}

	return result.String(), nil
}

// GetPrompt returns the prompt character to wait for after sending text
func (s *SendAction) GetPrompt() string {
	return s.prompt
}

// GetOnSuccessFunc returns the success callback function
func (s *SendAction) GetOnSuccessFunc() OnSuccessFunc {
	return s.onSuccessFunc
}
