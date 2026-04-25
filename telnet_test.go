package telnetsession

import (
	"regexp"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestSessionBuilder(t *testing.T) {
	builder := NewBuilder()

	// Test basic configuration
	session, err := builder.
		WithTimeout(5*time.Second).
		SetEnter("\n").
		SetPrompt(">").
		SetLoginExpr("Login:", "Password:").
		Expect("Welcome").
		Send("show version").
		Build()

	if err != nil {
		t.Fatalf("Failed to build session: %v", err)
	}

	if session.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", session.Timeout)
	}

	if session.Enter != "\n" {
		t.Errorf("Expected enter \\n, got %q", session.Enter)
	}

	if session.Prompt.String() != ">" {
		t.Errorf("Expected prompt >, got %q", session.Prompt.String())
	}

	if len(session.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(session.Actions))
	}
}

func TestActionTypes(t *testing.T) {
	// Test ExpectAction
	expectAction := &ExpectAction{pattern: regexp.MustCompile("test"), OnSuccessFunc: nil}
	if expectAction.GetType() != ActionExpect {
		t.Errorf("Expected ActionExpect, got %v", expectAction.GetType())
	}

	// Test SendAction
	sendAction := &SendAction{templ: nil, data: nil, prompt: regexp.MustCompile(">"), onSuccessFunc: nil}
	if sendAction.GetType() != ActionSend {
		t.Errorf("Expected ActionSend, got %v", sendAction.GetType())
	}
}

func TestSendActionTemplate(t *testing.T) {
	templ, err := template.New("test").Parse("Hello {{.name}}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]any{"name": "World"}
	sendAction := &SendAction{
		templ:         templ,
		data:          data,
		prompt:        regexp.MustCompile(">"),
		onSuccessFunc: nil,
	}

	text, err := sendAction.GetText()
	if err != nil {
		t.Fatalf("Failed to get text: %v", err)
	}

	expected := "Hello World"
	if text != expected {
		t.Errorf("Expected %q, got %q", expected, text)
	}
}

func TestTelnetSessionValidation(t *testing.T) {
	session, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("Failed to build session: %v", err)
	}

	telnet := New(session)

	// Test empty host
	err = telnet.Run("", 23, "", "")
	if err == nil {
		t.Error("Expected error for empty host")
	}

	// Test invalid port
	err = telnet.Run("localhost", 0, "", "")
	if err == nil {
		t.Error("Expected error for invalid port")
	}

	err = telnet.Run("localhost", 70000, "", "")
	if err == nil {
		t.Error("Expected error for invalid port")
	}
}

func TestGetOutput(t *testing.T) {
	session, err := NewBuilder().SetEnter("\n").Build()
	if err != nil {
		t.Fatalf("Failed to build session: %v", err)
	}

	telnet := New(session)

	// Simulate some output with consecutive empty lines
	telnet.output.WriteString("Line 1\n")
	telnet.output.WriteString("Line 2\n")
	telnet.output.WriteString("\n")
	telnet.output.WriteString("\n")
	telnet.output.WriteString("Line 3\n")

	output := telnet.GetOutput()

	// Debug: print the output
	t.Logf("Output: %q", output)
	t.Logf("Contains \\n\\n\\n: %v", strings.Contains(output, "\n\n\n"))

	// Verify that consecutive empty lines are reduced to single empty lines
	// The output should contain only one empty line between Line 2 and Line 3
	if strings.Contains(output, "\n\n\n") {
		t.Error("Output contains consecutive empty lines, should be reduced to single empty lines")
	}

	// Verify the content is preserved
	if !strings.Contains(output, "Line 1") || !strings.Contains(output, "Line 2") || !strings.Contains(output, "Line 3") {
		t.Error("Output is missing expected content")
	}
}

func TestBuilderErrorHandling(t *testing.T) {
	builder := NewBuilder()

	// Test invalid template
	builder.SendTempl("{{invalid template", map[string]any{})

	_, err := builder.Build()
	if err == nil {
		t.Error("Expected error for invalid template")
	}
}
