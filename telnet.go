package telnetsession

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type TelnetSession struct {
	conn    net.Conn
	reader  *bufio.Reader
	session *Session
	output  strings.Builder
}

func New(session *Session) *TelnetSession {
	return &TelnetSession{session: session}
}

func (t *TelnetSession) GetOutput() string {
	lines := strings.Split(t.output.String(), t.session.Enter)

	var result []string
	for i, line := range lines {
		if line == "" && (i > 0 && lines[i-1] == "") {
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, t.session.Enter)
}

func (t *TelnetSession) send(msg string) error {

	_, errDiscard := t.reader.Discard(t.reader.Buffered())
	if errDiscard != nil {
		return errDiscard
	}

	_, err := t.conn.Write([]byte(msg + t.session.Enter))
	return err
}

func (t *TelnetSession) expect(c byte) (string, error) {

	var response strings.Builder

	for {
		b, err := t.reader.ReadByte()
		if err != nil {
			return "", err
		}

		response.WriteByte(b)

		if b == c {
			break
		}
	}

	var result = response.String()
	t.output.WriteString(result)

	return result, nil
}

func (t *TelnetSession) expectString(s string) (string, error) {

	var response strings.Builder

	for {
		b, err := t.reader.ReadByte()
		if err != nil {
			return "", err
		}

		response.WriteByte(b)

		if strings.Contains(response.String(), s) {
			break
		}
	}

	var result = response.String()
	t.output.WriteString(result)

	return result, nil
}

func (t *TelnetSession) Run(host string, port int, user, pass string) error {

	var err error
	if t.session.Timeout > -1 {
		t.conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), t.session.Timeout)
	} else {
		t.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	}
	if err != nil {
		return err
	}

	defer func(conn net.Conn) { _ = conn.Close() }(t.conn)

	t.reader = bufio.NewReader(t.conn)

	if user != "" && pass != "" {

		if _, errExpectU := t.expectString(t.session.ExprUser); errExpectU != nil {
			return errExpectU
		}

		if errSendU := t.send(user); errSendU != nil {
			return errSendU
		}

		if _, errExpectP := t.expectString(t.session.ExprPass); errExpectP != nil {
			return errExpectP
		}

		if errSendP := t.send(pass); errSendP != nil {
			return errSendP
		}
	}

	for _, action := range t.session.Actions {

		text, errText := action.GetText()
		if errText != nil {
			return errText
		}

		switch action.GetType() {
		case expect:

			var out, errExpect = t.expectString(text)
			if errExpect != nil {
				return errExpect
			}

			if fn := action.GetOnSuccessFunc(); fn != nil {
				if errFn := fn(out); errFn != nil {
					return errFn
				}
			}

		case send:

			var errSend = t.send(text)
			if errSend != nil {
				return errSend
			}

			var out, errExpect = t.expect(action.GetPrompt()[0])
			if errExpect != nil {
				return errExpect
			}

			if fn := action.GetOnSuccessFunc(); fn != nil {
				if errFn := fn(out); errFn != nil {
					return errFn
				}
			}
		}
	}

	return nil
}
