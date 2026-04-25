package telnetsession

import "errors"

var (
	// ErrHostEmpty indicates the provided host is empty
	ErrHostEmpty = errors.New("host cannot be empty")
	// ErrInvalidPort indicates the provided port is out of range
	ErrInvalidPort = errors.New("port must be between 1 and 65535")
	// ErrConnectionFailed indicates the TCP connection could not be established
	ErrConnectionFailed = errors.New("failed to connect to remote host")
	// ErrLoginFailed indicates authentication with the remote host failed
	ErrLoginFailed = errors.New("authentication failed")
	// ErrActionFailed indicates an action execution failed
	ErrActionFailed = errors.New("action execution failed")
	// ErrReadFailed indicates a read operation from the connection failed
	ErrReadFailed = errors.New("failed to read from connection")
	// ErrWriteFailed indicates a write operation to the connection failed
	ErrWriteFailed = errors.New("failed to write to connection")
	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timed out")
	// ErrDetectedError indicates that a configured error pattern was matched in the device output
	ErrDetectedError = errors.New("device error detected")
)

