package telnetsession

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// TelnetSession manages a telnet connection and executes session actions
type TelnetSession struct {
	conn    net.Conn
	reader  *bufio.Reader
	session *Session
	output  strings.Builder
}

// New creates a new TelnetSession with the provided session configuration
func New(session *Session) *TelnetSession {
	return &TelnetSession{session: session}
}

// GetOutput returns the accumulated output from the session, with duplicate empty lines removed
func (t *TelnetSession) GetOutput() string {
	output := t.output.String()

	// Replace multiple consecutive Enter characters with single Enter
	// Use a simple loop to replace all occurrences
	for {
		newOutput := strings.ReplaceAll(output, t.session.Enter+t.session.Enter, t.session.Enter)
		if newOutput == output {
			break
		}
		output = newOutput
	}

	return output
}

// setReadDeadline sets a deadline for read operations if timeout is configured
func (t *TelnetSession) setReadDeadline() error {
	timeout := t.session.ReadTimeout
	if timeout == 0 {
		timeout = t.session.Timeout
	}
	if timeout > 0 {
		deadline := time.Now().Add(timeout)
		return t.conn.SetReadDeadline(deadline)
	}
	return nil
}

// setWriteDeadline sets a deadline for write operations if timeout is configured
func (t *TelnetSession) setWriteDeadline() error {
	timeout := t.session.WriteTimeout
	if timeout == 0 {
		timeout = t.session.Timeout
	}
	if timeout > 0 {
		deadline := time.Now().Add(timeout)
		return t.conn.SetWriteDeadline(deadline)
	}
	return nil
}

// send writes a message to the remote device and flushes the input buffer
func (t *TelnetSession) send(msg string) error {
	// Set write deadline for this operation
	if err := t.setWriteDeadline(); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Discard any buffered input before sending
	_, errDiscard := t.reader.Discard(t.reader.Buffered())
	if errDiscard != nil {
		return fmt.Errorf("failed to discard buffered input: %w", errDiscard)
	}

	// Send the message with the configured line ending
	_, err := t.conn.Write([]byte(msg + t.session.Enter))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// expect reads from the connection until the specified character is encountered
func (t *TelnetSession) expect(c byte) (string, error) {
	var response strings.Builder

	// Set read deadline for this operation
	if err := t.setReadDeadline(); err != nil {
		return "", fmt.Errorf("failed to set read deadline: %w", err)
	}

	for {
		b, err := t.reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("failed to read byte: %w", err)
		}

		response.WriteByte(b)

		if b == c {
			break
		}

		// Reset deadline for next read operation
		if errReset := t.setReadDeadline(); errReset != nil {
			return "", fmt.Errorf("failed to reset read deadline: %w", errReset)
		}
	}

	result := response.String()
	t.output.WriteString(result)

	return result, nil
}

// expectString reads from the connection until the specified string is found
func (t *TelnetSession) expectString(s string) (string, error) {
	var response strings.Builder

	// Set read deadline for this operation
	if err := t.setReadDeadline(); err != nil {
		return "", fmt.Errorf("failed to set read deadline: %w", err)
	}

	for {
		b, err := t.reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("failed to read byte: %w", err)
		}

		response.WriteByte(b)

		if strings.Contains(response.String(), s) {
			break
		}

		// Reset deadline for next read operation
		if errReset := t.setReadDeadline(); errReset != nil {
			return "", fmt.Errorf("failed to reset read deadline: %w", errReset)
		}
	}

	result := response.String()
	t.output.WriteString(result)

	return result, nil
}

// Run establishes a telnet connection and executes the configured session actions
func (t *TelnetSession) Run(host string, port int, user, pass string) error {
	// Validate input parameters
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if port <= 0 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}

	// Establish connection with timeout if configured
	var err error
	if t.session.Timeout > 0 {
		t.conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), t.session.Timeout)
	} else {
		t.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	}
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%d: %w", host, port, err)
	}

	defer func(conn net.Conn) { _ = conn.Close() }(t.conn)

	t.reader = bufio.NewReader(t.conn)

	// Handle login if credentials are provided
	if user != "" && pass != "" {
		if errLogin := t.handleLogin(user, pass); errLogin != nil {
			return fmt.Errorf("login failed: %w", errLogin)
		}
	}

	// Execute all configured actions
	if errExecActions := t.executeActions(); errExecActions != nil {
		return fmt.Errorf("failed to execute actions: %w", errExecActions)
	}

	return nil
}

// handleLogin performs the login sequence with username and password
func (t *TelnetSession) handleLogin(user, pass string) error {
	// Wait for username prompt
	if _, err := t.expectString(t.session.ExprUser); err != nil {
		return fmt.Errorf("failed to wait for username prompt: %w", err)
	}

	// Send username
	if err := t.send(user); err != nil {
		return fmt.Errorf("failed to send username: %w", err)
	}

	// Wait for password prompt
	if _, err := t.expectString(t.session.ExprPass); err != nil {
		return fmt.Errorf("failed to wait for password prompt: %w", err)
	}

	// Send password
	if err := t.send(pass); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}

	return nil
}

// executeActions runs all configured actions in sequence
func (t *TelnetSession) executeActions() error {
	for i, action := range t.session.Actions {
		if err := t.executeAction(action); err != nil {
			return fmt.Errorf("action %d failed: %w", i+1, err)
		}
	}
	return nil
}

// executeAction executes a single action based on its type
func (t *TelnetSession) executeAction(action Action) error {
	text, err := action.GetText()
	if err != nil {
		return fmt.Errorf("failed to get action text: %w", err)
	}

	text = strings.ReplaceAll(text, "\n", t.session.Enter)

	switch action.GetType() {
	case ActionExpect:
		return t.executeExpectAction(action, text)
	case ActionSend:
		return t.executeSendAction(action, text)
	default:
		return fmt.Errorf("unknown action type: %d", action.GetType())
	}
}

// executeExpectAction handles expect-type actions
func (t *TelnetSession) executeExpectAction(action Action, text string) error {
	out, err := t.expectString(text)
	if err != nil {
		return fmt.Errorf("failed to expect text '%s': %w", text, err)
	}

	if fn := action.GetOnSuccessFunc(); fn != nil {
		if err := fn(out); err != nil {
			return fmt.Errorf("onSuccess callback failed: %w", err)
		}
	}

	return nil
}

// executeSendAction handles send-type actions
func (t *TelnetSession) executeSendAction(action Action, text string) error {
	var result strings.Builder

	for _, line := range strings.Split(text, t.session.Enter) {
		if line == "" {
			continue
		}

		if err := t.send(line); err != nil {
			return fmt.Errorf("failed to send line '%s': %w", line, err)
		}

		prompt := action.GetPrompt()
		if prompt != "" {
			out, err := t.expect(prompt[0])
			if err != nil {
				return fmt.Errorf("failed to wait for prompt '%s': %w", prompt, err)
			}
			result.WriteString(out)
		}
	}

	if fn := action.GetOnSuccessFunc(); fn != nil {
		if err := fn(result.String()); err != nil {
			return fmt.Errorf("onSuccess callback failed: %w", err)
		}
	}

	return nil
}
