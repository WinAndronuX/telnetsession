package telnetsession

import (
	"regexp"
	"time"
)

// Session represents a telnet session configuration with a sequence of actions to perform
type Session struct {
	// Actions is the sequence of actions to execute during the session
	Actions []Action
	// Enter is the line ending character(s) to use when sending commands
	Enter string
	// Prompt is the regular expression to wait for after sending commands (e.g., ">", "#", "$")
	Prompt *regexp.Regexp
	// ExprUser is the regular expression pattern to expect when prompting for username
	ExprUser *regexp.Regexp
	// ExprPass is the regular expression pattern to expect when prompting for password
	ExprPass *regexp.Regexp
	// Timeout is the connection timeout duration. Use 0 for no timeout
	Timeout time.Duration
	// ReadTimeout is the timeout for read operations. Use 0 for no timeout
	ReadTimeout time.Duration
	// WriteTimeout is the timeout for write operations. Use 0 for no timeout
	WriteTimeout time.Duration
	// MorePattern is a regular expression to detect pagination prompts (e.g., "--More--")
	MorePattern *regexp.Regexp
	// MoreResponse is the string to send when a pagination prompt is detected
	MoreResponse string
	// ErrorPatterns is a list of regular expressions that indicate a device error or failure
	ErrorPatterns []*regexp.Regexp
	// InitialActions are actions executed immediately after connection, before login
	InitialActions []Action
	// Debug enables verbose logging of sent/received bytes
	Debug bool
}
