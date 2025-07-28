package telnetsession

import "time"

type Session struct {
	Actions  []Action
	Enter    string
	Prompt   string
	ExprUser string
	ExprPass string
	Timeout  time.Duration
}
