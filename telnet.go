package telnetsession

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// TelnetSession manages a telnet connection and executes session actions
type TelnetSession struct {
	conn       net.Conn
	reader     *bufio.Reader
	session    *Session
	output     bytes.Buffer
	ctx        context.Context
	state      State
	moreBuffer bytes.Buffer
}

// New creates a new TelnetSession with the provided session configuration
func New(session *Session) *TelnetSession {
	return &TelnetSession{
		session: session,
		ctx:     context.Background(),
		state:   StateDisconnected,
	}
}

// transition updates the current state of the session
func (t *TelnetSession) transition(s State) {
	t.state = s
}

// GetOutput returns the accumulated output from the session, with ANSI codes and duplicate empty lines removed
func (t *TelnetSession) GetOutput() string {
	output := t.output.String()

	// 1. Remove ANSI escape codes (colors, cursor movements, etc.)
	ansi := regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	output = ansi.ReplaceAllString(output, "")

	// 2. Replace multiple consecutive Enter characters with single Enter
	re := regexp.MustCompile(fmt.Sprintf("(%s){2,}", regexp.QuoteMeta(t.session.Enter)))
	return re.ReplaceAllString(output, t.session.Enter)
}

// setReadDeadline sets a deadline for read operations if timeout is configured
func (t *TelnetSession) setReadDeadline() error {
	timeout := t.session.ReadTimeout
	if timeout == 0 {
		timeout = t.session.Timeout
	}

	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	// Check if context has a shorter deadline
	if ctxDeadline, ok := t.ctx.Deadline(); ok {
		if deadline.IsZero() || ctxDeadline.Before(deadline) {
			deadline = ctxDeadline
		}
	}

	if !deadline.IsZero() {
		if err := t.conn.SetReadDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}
	}
	return nil
}

// setWriteDeadline sets a deadline for write operations if timeout is configured
func (t *TelnetSession) setWriteDeadline() error {
	timeout := t.session.WriteTimeout
	if timeout == 0 {
		timeout = t.session.Timeout
	}

	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	// Check if context has a shorter deadline
	if ctxDeadline, ok := t.ctx.Deadline(); ok {
		if deadline.IsZero() || ctxDeadline.Before(deadline) {
			deadline = ctxDeadline
		}
	}

	if !deadline.IsZero() {
		if err := t.conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set write deadline: %w", err)
		}
	}
	return nil
}

// send writes a message to the remote device and flushes the input buffer
func (t *TelnetSession) send(msg string) error {
	if t.session.Debug {
		fmt.Printf(">>> SEND: %q\n", msg)
	}
	// Set write deadline for this operation
	if err := t.setWriteDeadline(); err != nil {
		return err
	}

	// Discard any buffered input before sending
	_, errDiscard := t.reader.Discard(t.reader.Buffered())
	if errDiscard != nil {
		return fmt.Errorf("failed to discard buffered input: %w", errDiscard)
	}

	// Send the message with the configured line ending
	_, err := t.conn.Write([]byte(msg + t.session.Enter))
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("%w: %w", ErrTimeout, err)
		}
		return fmt.Errorf("%w: %w", ErrWriteFailed, err)
	}

	return nil
}

// readByte handles Telnet IAC (Interpret As Command) sequences and Auto-More detection
func (t *TelnetSession) readByte() (byte, error) {
	for {
		b, err := t.reader.ReadByte()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return 0, fmt.Errorf("%w: %w", ErrTimeout, err)
			}
			return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
		}

		if t.session.Debug {
			fmt.Printf("<<< RECV: %q\n", b)
		}

		if b == 255 {
			// It's an IAC sequence
			next, err := t.reader.ReadByte()
			if err != nil {
				return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
			}

			switch next {
			case 255: // Escaped IAC
				b = 255
			case 251, 252, 253, 254: // WILL, WONT, DO, DONT
				// Read the option byte and discard
				_, err = t.reader.ReadByte()
				if err != nil {
					return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
				}
				// We could respond here if we wanted to negotiate,
				// but for now we just ignore and continue reading
				continue
			case 250: // SB (Subnegotiation)
				// Read until SE (240)
				for {
					bSub, err := t.reader.ReadByte()
					if err != nil {
						return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
					}
					if bSub == 240 {
						break
					}
				}
				continue
			default:
				// Other 2-byte commands
				continue
			}
		}

		// Pagination and Error detection (using a sliding window in moreBuffer)
		if t.session.MorePattern != nil || len(t.session.ErrorPatterns) > 0 {
			t.moreBuffer.WriteByte(b)
			// Limit moreBuffer size to 256 bytes to cover longer error messages
			if t.moreBuffer.Len() > 256 {
				t.moreBuffer.Next(1)
			}

			// 1. Check for Error Patterns first (High priority)
			for _, errPattern := range t.session.ErrorPatterns {
				if errPattern.Match(t.moreBuffer.Bytes()) {
					t.transition(StateErrorDetected)
					return 0, fmt.Errorf("%w: %s", ErrDetectedError, errPattern.String())
				}
			}

			// 2. Check for Pagination (Auto-More)
			if t.session.MorePattern != nil && t.session.MorePattern.Match(t.moreBuffer.Bytes()) {
				// Set write deadline for the response
				if errWriteTimeout := t.setWriteDeadline(); errWriteTimeout == nil {
					// Send the response directly to connection
					if _, errWrite := t.conn.Write([]byte(t.session.MoreResponse)); errWrite != nil {
						return 0, fmt.Errorf("failed to send pagination response: %w", errWrite)
					}
					// Clear the buffer so we don't match again immediately
					t.moreBuffer.Reset()
				}
			}
		}

		return b, nil
	}
}

// expectRegex reads from the connection until the specified regular expression matches the accumulated output
func (t *TelnetSession) expectRegex(re *regexp.Regexp) (string, error) {
	var response bytes.Buffer

	if re == nil {
		return "", fmt.Errorf("regular expression cannot be nil")
	}

	// Set read deadline for this operation
	if err := t.setReadDeadline(); err != nil {
		return "", err
	}

	for {
		b, err := t.readByte()
		if err != nil {
			return "", err
		}

		response.WriteByte(b)

		if re.Match(response.Bytes()) {
			break
		}

		// Reset deadline for next read operation
		if errReset := t.setReadDeadline(); errReset != nil {
			return "", errReset
		}
	}

	result := response.Bytes()
	t.output.Write(result)

	return string(result), nil
}

// Run establishes a telnet connection and executes the configured session actions
func (t *TelnetSession) Run(host string, port int, user, pass string) error {
	return t.RunWithContext(context.Background(), host, port, user, pass)
}

// RunWithContext establishes a telnet connection and executes actions with a context
func (t *TelnetSession) RunWithContext(ctx context.Context, host string, port int, user, pass string) error {
	t.ctx = ctx
	t.transition(StateConnecting)

	// Validate input parameters
	if host == "" {
		t.transition(StateError)
		return ErrHostEmpty
	}
	if port <= 0 || port > 65535 {
		t.transition(StateError)
		return fmt.Errorf("%w: got %d", ErrInvalidPort, port)
	}

	// Establish connection with timeout if configured
	var err error
	dialer := net.Dialer{}
	if t.session.Timeout > 0 {
		dialer.Timeout = t.session.Timeout
	}

	t.conn, err = dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		t.transition(StateError)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("%w: %w", ErrTimeout, err)
		}
		return fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}

	defer func() {
		_ = t.conn.Close()
		if t.state != StateError && t.state != StateErrorDetected {
			t.transition(StateClosed)
		}
	}()

	t.reader = bufio.NewReader(t.conn)

	// Execute Initial Actions (Pre-login, e.g., for Nokia TL1)
	for i, action := range t.session.InitialActions {
		t.transition(StateExecuting)
		if err := t.executeAction(action); err != nil {
			t.transition(StateError)
			return fmt.Errorf("initial action %d failed: %w", i+1, err)
		}
	}

	// Handle login if credentials are provided
	if user != "" && pass != "" {
		t.transition(StateAuthenticating)
		if errLogin := t.handleLogin(user, pass); errLogin != nil {
			t.transition(StateError)
			return fmt.Errorf("%w: %w", ErrLoginFailed, errLogin)
		}
	}

	t.transition(StateReady)

	// Execute all configured actions
	if errExecActions := t.executeActions(); errExecActions != nil {
		t.transition(StateError)
		return fmt.Errorf("%w: %w", ErrActionFailed, errExecActions)
	}

	return nil
}

// handleLogin performs the login sequence with username and password
func (t *TelnetSession) handleLogin(user, pass string) error {
	// Wait for username prompt
	if t.session.ExprUser != nil {
		if _, err := t.expectRegex(t.session.ExprUser); err != nil {
			return fmt.Errorf("failed to wait for username prompt: %w", err)
		}
	}

	// Send username
	if err := t.send(user); err != nil {
		return fmt.Errorf("failed to send username: %w", err)
	}

	// Wait for password prompt
	if t.session.ExprPass != nil {
		if _, err := t.expectRegex(t.session.ExprPass); err != nil {
			return fmt.Errorf("failed to wait for password prompt: %w", err)
		}
	}

	// Send password
	if err := t.send(pass); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}

	// Wait for final prompt after login if configured
	if t.session.Prompt != nil {
		if _, err := t.expectRegex(t.session.Prompt); err != nil {
			return fmt.Errorf("failed to wait for prompt after login: %w", err)
		}
	}

	return nil
}

// executeActions runs all configured actions in sequence
func (t *TelnetSession) executeActions() error {
	for i, action := range t.session.Actions {
		t.transition(StateExecuting)
		if err := t.executeAction(action); err != nil {
			return fmt.Errorf("action %d failed: %w", i+1, err)
		}
		t.transition(StateReady)
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
		return t.executeExpectAction(action)
	case ActionSend:
		return t.executeSendAction(action, text)
	default:
		return fmt.Errorf("unknown action type: %d", action.GetType())
	}
}

// executeExpectAction handles expect-type actions
func (t *TelnetSession) executeExpectAction(action Action) error {
	expectAction, ok := action.(*ExpectAction)
	if !ok {
		return fmt.Errorf("action is not an ExpectAction")
	}

	out, err := t.expectRegex(expectAction.GetPattern())
	if err != nil {
		return fmt.Errorf("failed to expect pattern '%s': %w", expectAction.GetPattern().String(), err)
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
		if prompt != nil {
			out, err := t.expectRegex(prompt)
			if err != nil {
				return fmt.Errorf("failed to wait for prompt '%s': %w", prompt.String(), err)
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
