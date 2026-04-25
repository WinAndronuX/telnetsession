package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/WinAndronuX/telnetsession"
)

func TestPublicServerTelehackComplex(t *testing.T) {
	session, err := telnetsession.NewBuilder().
		WithTimeout(15*time.Second).
		SetPrompt(`(?m)^\.$`).
		Send("date").
		SendAndDo("uptime", func(output string) error {
			if !strings.Contains(strings.ToLower(output), "load") && !strings.Contains(strings.ToLower(output), "up") {
				return errors.New("uptime failed")
			}
			return nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	ts := telnetsession.New(session)
	err = ts.Run("telehack.com", 23, "", "")
	if err != nil {
		t.Skipf("Telehack unreachable: %v", err)
		return
	}

	if !strings.Contains(ts.GetOutput(), "uptime") {
		t.Error("Incomplete output")
	}
}

func TestPublicServerFreeChess(t *testing.T) {
	session, err := telnetsession.NewBuilder().
		WithTimeout(15*time.Second).
		SetPrompt(`fics% `).
		Expect("Free Internet Chess Server").
		Send("uptime").
		Build()

	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	ts := telnetsession.New(session)
	err = ts.Run("freechess.org", 5000, "", "")
	if err != nil {
		t.Skipf("Freechess unreachable: %v", err)
		return
	}

	if !strings.Contains(ts.GetOutput(), "uptime") {
		t.Error("Incomplete output")
	}
}
