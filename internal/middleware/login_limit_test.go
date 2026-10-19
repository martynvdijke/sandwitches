package middleware

import (
	"testing"
	"time"
)

func TestLoginLimiterPerIPAndAccount(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	l := NewLoginLimiter(3, 10*time.Minute, 5*time.Minute)
	l.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if ok, _ := l.Allow("1.2.3.4", "alice"); !ok {
			t.Fatalf("attempt %d should be allowed", i)
		}
		l.Failure("1.2.3.4", "alice")
	}

	if ok, wait := l.Allow("1.2.3.4", "alice"); ok || wait <= 0 {
		t.Fatalf("expected lockout, ok=%v wait=%v", ok, wait)
	}
	// The account is locked from any IP, and the IP is locked for any account.
	if ok, _ := l.Allow("9.9.9.9", "alice"); ok {
		t.Error("account lockout must apply across IPs")
	}
	if ok, _ := l.Allow("1.2.3.4", "bob"); ok {
		t.Error("IP lockout must apply across accounts")
	}
	// A different account from a different IP is unaffected.
	if ok, _ := l.Allow("9.9.9.9", "bob"); !ok {
		t.Error("unrelated keys should not be locked")
	}

	now = now.Add(5*time.Minute + time.Second)
	if ok, _ := l.Allow("1.2.3.4", "alice"); !ok {
		t.Error("lockout should expire")
	}

	// Success clears counters so a legitimate retry is not penalised.
	l.Failure("1.2.3.4", "carol")
	l.Success("1.2.3.4", "carol")
	if ok, _ := l.Allow("1.2.3.4", "carol"); !ok {
		t.Error("successful login should clear failures")
	}
}

func TestLoginLimiterWindowResetsFailures(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	l := NewLoginLimiter(2, time.Minute, time.Minute)
	l.now = func() time.Time { return now }

	l.Failure("ip", "user")
	now = now.Add(2 * time.Minute)
	l.Failure("ip", "user")
	if ok, _ := l.Allow("ip", "user"); !ok {
		t.Error("failures outside the window should not accumulate")
	}
}
