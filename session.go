package telnetsession

import "time"

// Session represents a telnet session configuration with a sequence of actions to perform
type Session struct {
	// Actions is the sequence of actions to execute during the session
	Actions []Action
	// Enter is the line ending character(s) to use when sending commands
	Enter string
	// Prompt is the character to wait for after sending commands (e.g., ">", "#", "$")
	Prompt string
	// ExprUser is the text pattern to expect when prompting for username
	ExprUser string
	// ExprPass is the text pattern to expect when prompting for password
	ExprPass string
	// Timeout is the connection timeout duration. Use 0 for no timeout
	Timeout time.Duration
	// ReadTimeout is the timeout for read operations. Use 0 for no timeout
	ReadTimeout time.Duration
	// WriteTimeout is the timeout for write operations. Use 0 for no timeout
	WriteTimeout time.Duration
}
