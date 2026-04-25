package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/WinAndronuX/telnetsession"
)

func TestContainerlabSwitch1(t *testing.T) {
	session, err := telnetsession.NewBuilder().
		WithTimeout(5*time.Second).
		SetEnter("\r").
		SetPrompt(`(?m)[#$]\s*$`).
		WithErrors(`(?i)Login incorrect`).
		Expect(`[#$]\s*$`).
		Send("ls -la /").
		SendAndDo("uptime", func(output string) error {
			if !strings.Contains(output, "up") && !strings.Contains(output, "load") {
				return errors.New("uptime output failed")
			}
			return nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build session: %v", err)
	}

	ts := telnetsession.New(session)
	err = ts.Run("172.30.30.2", 23, "", "")
	if err != nil {
		t.Skipf("Containerlab not running: %v", err)
		return
	}

	output := ts.GetOutput()
	if !strings.Contains(output, "etc") {
		t.Error("Output incomplete")
	}
}

func TestComplexFinalFlow(t *testing.T) {
	session, err := telnetsession.NewBuilder().
		WithTimeout(10*time.Second).
		SetEnter("\r").
		SetPrompt(`(?m)[#$]\s*$`).
		WithErrors(`(?i)command not found`, `(?i)no such file`).
		Expect(`[#$]\s*$`).
		SendTempl("echo 'vlan {{.id}}' > /tmp/config_{{.id}}", map[string]any{"id": 100}).
		Send("cat /tmp/config_100").
		Send("rm -i /tmp/config_100").
		Confirm("remove", "y").
		Send("seq 1 5").
		Build()

	if err != nil {
		t.Fatalf("Builder failed: %v", err)
	}

	ts := telnetsession.New(session)
	err = ts.Run("172.30.30.2", 23, "", "")
	if err != nil {
		t.Skipf("Containerlab not running: %v", err)
		return
	}

	if !strings.Contains(ts.GetOutput(), "vlan 100") {
		t.Error("Template-based content not found")
	}
}
