/*
Package telnetsession provides a fluent interface for creating and executing telnet sessions with remote devices.

The library is designed to simplify telnet automation by providing a builder pattern for defining session actions
and a clean API for executing them.

# Basic Usage

Create a session using the builder pattern:

	session, err := telnetsession.NewBuilder().
		WithTimeout(5 * time.Second).
		SetEnter("\n").
		SetLoginExpr("Login:", "Password:").
		SetPrompt(">").
		Send("show version").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	device := telnetsession.New(session)
	err = device.Run("192.168.1.1", 23, "admin", "password")

# Session Configuration

The SessionBuilder provides several methods for configuring the session:

- WithTimeout(duration): Sets connection timeout
- WithReadTimeout(duration): Sets the read timeout for individual read operations
- WithWriteTimeout(duration): Sets the write timeout for individual write operations
- SetEnter(string): Sets line ending character(s)
- SetPrompt(string): Sets prompt character to wait for after commands
- SetLoginExpr(username, password): Sets text patterns for login prompts

# Actions

The library supports two types of actions:

## Expect Actions
Wait for specific text to appear in the session:

	builder.Expect("Welcome")
	builder.ExpectAndDo("Error", func(output string) error {
		// Handle error condition
		return nil
	})

## Send Actions
Send commands to the remote device:

	builder.Send("configure terminal")
	builder.SendAndDo("show version", func(output string) error {
		// Process output
		return nil
	})

## Template Actions
Send templated commands with dynamic data:

	data := map[string]any{
		"interfaces": []string{"eth0", "eth1", "eth2"},
	}

	builder.SendTempl(`

{{ range .interfaces }}
interface {{ . }}
show interface
{{ end }}`, data)

# Error Handling

The library provides comprehensive error handling:

- Connection errors are wrapped with context
- Template parsing errors are collected during building
- Action execution errors include action index and context
- Validation errors for invalid parameters

# Output Processing

The TelnetSession accumulates all output and provides a GetOutput() method that:

- Removes duplicate empty lines
- Preserves command output structure
- Returns clean, formatted output

# Examples

See the examples/ directory for complete working examples:

- examples/vsol/: Basic session with VSOL device
- examples/vsoltempl/: Session using templates for dynamic commands
*/
package telnetsession
