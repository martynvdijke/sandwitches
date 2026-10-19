package middleware

import (
	"strings"
	"sync"
	"time"
)

// LoginLimiter tracks failed login attempts per client IP and per account.
// A successful login clears both counters. Entries are pruned lazily so a
// long-running process cannot accumulate unbounded state.
type LoginLimiter struct {
	mu          sync.Mutex
	entries     map[string]*loginEntry
	maxFailures int
	window      time.Duration
	lockout     time.Duration
	now         func() time.Time
}

type loginEntry struct {
	failures  int
	since     time.Time
	lockedTo  time.Time
	lastSeen  time.Time
}

// NewLoginLimiter creates a limiter that locks a key for lockout after
// maxFailures failed attempts within window.
func NewLoginLimiter(maxFailures int, window, lockout time.Duration) *LoginLimiter {
	if maxFailures < 1 {
		maxFailures = 1
	}
	return &LoginLimiter{
		entries:     make(map[string]*loginEntry),
		maxFailures: maxFailures,
		window:      window,
		lockout:     lockout,
		now:         time.Now,
	}
}

func loginKeys(ip, username string) []string {
	username = strings.ToLower(strings.TrimSpace(username))
	keys := make([]string, 0, 2)
	if ip != "" {
		keys = append(keys, "ip:"+ip)
	}
	if username != "" {
		keys = append(keys, "user:"+username)
	}
	return keys
}

// Allow reports whether a login attempt may proceed and, when it may not, how
// long the caller should wait before retrying.
func (l *LoginLimiter) Allow(ip, username string) (bool, time.Duration) {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(now)

	for _, key := range loginKeys(ip, username) {
		e, ok := l.entries[key]
		if !ok {
			continue
		}
		e.lastSeen = now
		if now.Before(e.lockedTo) {
			return false, e.lockedTo.Sub(now)
		}
	}
	return true, 0
}

// Failure records a failed attempt for both the IP and the account.
func (l *LoginLimiter) Failure(ip, username string) {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(now)

	for _, key := range loginKeys(ip, username) {
		e, ok := l.entries[key]
		if !ok {
			e = &loginEntry{since: now}
			l.entries[key] = e
		}
		if now.Sub(e.since) > l.window {
			e.failures = 0
			e.since = now
		}
		e.failures++
		e.lastSeen = now
		if e.failures >= l.maxFailures {
			e.lockedTo = now.Add(l.lockout)
		}
	}
}

// Success clears the counters for the IP and account after a valid login.
func (l *LoginLimiter) Success(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, key := range loginKeys(ip, username) {
		delete(l.entries, key)
	}
}

// Reset clears all state. It is primarily useful in tests.
func (l *LoginLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make(map[string]*loginEntry)
}

// pruneLocked drops entries that are neither locked nor recent. Callers must
// hold l.mu.
func (l *LoginLimiter) pruneLocked(now time.Time) {
	if len(l.entries) < 1024 {
		return
	}
	cutoff := l.window
	if l.lockout > cutoff {
		cutoff = l.lockout
	}
	for key, e := range l.entries {
		if now.Before(e.lockedTo) {
			continue
		}
		if now.Sub(e.lastSeen) > cutoff {
			delete(l.entries, key)
		}
	}
}
