package telnetsession

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func setupTestServer(t *testing.T, handler func(net.Conn)) (string, int) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	go func() {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handler(conn)
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port
}

func TestFullSessionIntegration(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()
		reader := bufio.NewReader(conn)
		conn.Write([]byte{255, 251, 1})
		fmt.Fprint(conn, "Login: ")
		reader.ReadString('\n')
		fmt.Fprint(conn, "Password: ")
		reader.ReadString('\n')
		fmt.Fprint(conn, "Switch>")
		line, _ := reader.ReadString('\n')
		if strings.Contains(line, "show version") {
			fmt.Fprint(conn, "v1.0.0\nSwitch>")
		}
	}

	host, port := setupTestServer(t, handler)
	session, _ := NewBuilder().
		SetLoginExpr("Login:", "Password:").
		SetPrompt(`Switch>`).
		SendAndDo("show version", func(output string) error {
			if !strings.Contains(output, "v1.0.0") {
				return fmt.Errorf("missing version")
			}
			return nil
		}).
		Build()

	ts := New(session)
	if err := ts.Run(host, port, "admin", "pass"); err != nil {
		t.Errorf("Failed: %v", err)
	}
}

func TestPaginationIntegration(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()
		reader := bufio.NewReader(conn)
		reader.ReadString('\n')
		fmt.Fprint(conn, "Part 1\n--More--")
		buf := make([]byte, 1)
		reader.Read(buf)
		if buf[0] == ' ' {
			fmt.Fprint(conn, "Part 2\nSwitch>")
		}
	}

	host, port := setupTestServer(t, handler)
	session, _ := NewBuilder().
		SetPrompt(`Switch>`).
		WithPagination(`--More--`, " ").
		Send("show long").
		Build()

	ts := New(session)
	ts.Run(host, port, "", "")
	output := ts.GetOutput()
	if !strings.Contains(output, "Part 1") || !strings.Contains(output, "Part 2") {
		t.Errorf("Incomplete: %q", output)
	}
}

func TestErrorDetection(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()
		fmt.Fprint(conn, "Switch>")
		bufio.NewReader(conn).ReadString('\n')
		fmt.Fprint(conn, "% Invalid command\nSwitch>")
	}

	host, port := setupTestServer(t, handler)
	session, _ := NewBuilder().
		SetPrompt("Switch>").
		Expect("Switch>").
		WithErrors("Invalid command").
		Send("bad").
		Build()

	ts := New(session)
	err := ts.Run(host, port, "", "")
	if err == nil || !strings.Contains(err.Error(), "device error detected") {
		t.Errorf("Expected detection error, got: %v", err)
	}
}

func TestNokiaStyleInitialActions(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()
		reader := bufio.NewReader(conn)

		fmt.Fprint(conn, "Press 'T': ")
		reader.ReadString('\n')
		fmt.Fprint(conn, "Confirm? [y/n]: ")
		reader.ReadString('\n')

		// Login con ANSI
		fmt.Fprint(conn, "\x1b[32mLogin:\x1b[0m ")
		reader.ReadString('\n')
		fmt.Fprint(conn, "Password: ")
		reader.ReadString('\n')

		fmt.Fprint(conn, "NOKIA-OLT# ")
		time.Sleep(50 * time.Millisecond)
	}

	host, port := setupTestServer(t, handler)
	session, _ := NewBuilder().
		SendInitial("T").
		ExpectInitial("Confirm?").
		SendInitial("y").
		SetLoginExpr("Login:", "Password:").
		SetPrompt("NOKIA-OLT#").
		Build()

	ts := New(session)
	if err := ts.Run(host, port, "admin", "pass"); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	output := ts.GetOutput()
	if strings.Contains(output, "\x1b[32m") {
		t.Error("ANSI codes were not cleaned")
	}
}
