package telnetsession

import (
	"context"
	"sync"
	"time"
)

// deviceKey uniquely identifies a device by host and port
type deviceKey struct {
	host string
	port int
}

type limiterEntry struct {
	ch       chan struct{}
	lastUsed time.Time
}

// Pool manages exclusive access to devices and maintains a registry of active sessions
type Pool struct {
	mu       sync.Mutex
	limiters map[deviceKey]*limiterEntry
	sessions map[deviceKey]*TelnetSession
}

// NewPool creates a new Pool instance with an automatic cleanup goroutine
func NewPool() *Pool {
	p := &Pool{
		limiters: make(map[deviceKey]*limiterEntry),
		sessions: make(map[deviceKey]*TelnetSession),
	}
	go p.cleanupLoop()
	return p
}

func (p *Pool) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		now := time.Now()
		for key, entry := range p.limiters {
			// If no active session and hasn't been used in 10 minutes, clean up
			if len(entry.ch) == 0 && now.Sub(entry.lastUsed) > 10*time.Minute {
				delete(p.limiters, key)
			}
		}
		p.mu.Unlock()
	}
}

// Do executes a telnet session ensuring that only one request accesses the device (host:port) at a time.
// If the device is already in use, the call blocks until it's available or the context is cancelled.
func (p *Pool) Do(ctx context.Context, s *Session, host string, port int, user, pass string) error {
	key := deviceKey{host: host, port: port}

	// Get or create the limiter for this specific device
	p.mu.Lock()
	entry, ok := p.limiters[key]
	if !ok {
		entry = &limiterEntry{
			ch: make(chan struct{}, 1),
		}
		p.limiters[key] = entry
	}
	entry.lastUsed = time.Now()
	p.mu.Unlock()

	// Try to acquire access to the device respecting the context
	select {
	case entry.ch <- struct{}{}:
		// Access granted
	case <-ctx.Done():
		return ctx.Err()
	}

	// Ensure release and metrics update upon completion
	defer func() {
		p.mu.Lock()
		if e, ok := p.limiters[key]; ok {
			e.lastUsed = time.Now()
		}
		p.mu.Unlock()
		<-entry.ch
	}()

	// Create the session and register it in the active sessions map
	ts := New(s)

	p.mu.Lock()
	p.sessions[key] = ts
	p.mu.Unlock()

	// Remove from registry when execution ends
	defer func() {
		p.mu.Lock()
		delete(p.sessions, key)
		p.mu.Unlock()
	}()

	// Execute original session logic
	return ts.RunWithContext(ctx, host, port, user, pass)
}

// GetActiveSession returns the TelnetSession instance currently operating on a device.
// Useful for monitoring or external cancellations. Returns nil if no active session exists.
func (p *Pool) GetActiveSession(host string, port int) *TelnetSession {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sessions[deviceKey{host: host, port: port}]
}

// CountActiveSessions returns the total number of currently active telnet sessions.
func (p *Pool) CountActiveSessions() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.sessions)
}
