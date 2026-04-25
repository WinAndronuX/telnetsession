package telnetsession

import "fmt"

// State represents the current state of a Telnet session
type State int

const (
	// StateDisconnected indicates the session is not connected
	StateDisconnected State = iota
	// StateConnecting indicates the session is establishing a TCP connection
	StateConnecting
	// StateAuthenticating indicates the session is performing login
	StateAuthenticating
	// StateReady indicates the session is connected and ready for actions
	StateReady
	// StateExecuting indicates an action is currently being executed
	StateExecuting
	// StateClosed indicates the session has been closed
	StateClosed
	// StateError indicates the session encountered a terminal network or system error
	StateError
	// StateErrorDetected indicates a configured error pattern was matched
	StateErrorDetected
)

func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateAuthenticating:
		return "Authenticating"
	case StateReady:
		return "Ready"
	case StateExecuting:
		return "Executing"
	case StateClosed:
		return "Closed"
	case StateError:
		return "Error"
	case StateErrorDetected:
		return "ErrorDetected"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}
