package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/storage/memory"
)

func TestLimiter_BasicFlow(t *testing.T) {
	tests := []struct {
		name        string
		rule        Rule
		attempts    int // number of attempts to make
		wantAllow   bool
		wantAllowed int // number of attempts that should be allowed
	}{
		{
			name: "under limit",
			rule: Rule{
				Limit:  3,
				Window: time.Second,
			},
			attempts:    2,
			wantAllow:   true,
			wantAllowed: 2,
		},
		{
			name: "at limit",
			rule: Rule{
				Limit:  3,
				Window: time.Second,
			},
			attempts:    3,
			wantAllow:   true,
			wantAllowed: 3,
		},
		{
			name: "over limit",
			rule: Rule{
				Limit:  3,
				Window: time.Second,
			},
			attempts:    5,
			wantAllow:   false, // last attempt should be rejected
			wantAllowed: 3,
		},
		{
			name: "zero limit",
			rule: Rule{
				Limit:  0,
				Window: time.Second,
			},
			attempts:    1,
			wantAllow:   false,
			wantAllowed: 0,
		},
		{
			name: "zero window",
			rule: Rule{
				Limit:  5,
				Window: 0,
			},
			attempts:    10,
			wantAllow:   true, // per inmemorydb implementation, zero window = always allow
			wantAllowed: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := memory.NewInMemoryDB()
			l := NewInMemoryLimiter(db, "test-key", tt.rule)
			ctx := context.Background()

			allowed := 0
			var lastAllow bool
			var lastErr error

			// Make attempts
			for i := 0; i < tt.attempts; i++ {
				lastAllow, lastErr = l.Allow(ctx)
				if lastErr != nil {
					t.Fatalf("attempt %d: unexpected error: %v", i+1, lastErr)
				}
				if lastAllow {
					allowed++
				}
			}

			if lastAllow != tt.wantAllow {
				t.Errorf("last attempt: got allow=%v, want %v", lastAllow, tt.wantAllow)
			}
			if allowed != tt.wantAllowed {
				t.Errorf("total allowed: got %d, want %d", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestLimiter_SlidingWindow(t *testing.T) {
	db := memory.NewInMemoryDB()
	rule := Rule{
		Limit:  2,
		Window: 100 * time.Millisecond,
	}
	l := NewInMemoryLimiter(db, "sliding-key", rule)
	ctx := context.Background()

	// First attempt should be allowed
	allow, err := l.Allow(ctx)
	if err != nil || !allow {
		t.Fatalf("first attempt: got allow=%v err=%v, want allow=true err=nil", allow, err)
	}

	// Second attempt should be allowed
	allow, err = l.Allow(ctx)
	if err != nil || !allow {
		t.Fatalf("second attempt: got allow=%v err=%v, want allow=true err=nil", allow, err)
	}

	// Third attempt should be rejected (limit reached)
	allow, err = l.Allow(ctx)
	if err != nil || allow {
		t.Fatalf("third attempt: got allow=%v err=%v, want allow=false err=nil", allow, err)
	}

	// Wait for window to slide past first attempt
	time.Sleep(rule.Window + 10*time.Millisecond)

	// Should allow one more attempt now
	allow, err = l.Allow(ctx)
	if err != nil || !allow {
		t.Fatalf("after window: got allow=%v err=%v, want allow=true err=nil", allow, err)
	}
}

func TestLimiter_Reset(t *testing.T) {
	db := memory.NewInMemoryDB()
	rule := Rule{
		Limit:  2,
		Window: time.Minute,
	}
	l := NewInMemoryLimiter(db, "reset-key", rule)
	ctx := context.Background()

	// Up the limit
	for i := 0; i < rule.Limit; i++ {
		allow, err := l.Allow(ctx)
		if err != nil || !allow {
			t.Fatalf("attempt %d: got allow=%v err=%v, want allow=true err=nil", i+1, allow, err)
		}
	}

	// Next attempt should be rejected
	if allow, err := l.Allow(ctx); err != nil || allow {
		t.Fatalf("over-limit: got allow=%v err=%v, want allow=false err=nil", allow, err)
	}

	// Reset should clear the limit
	if err := l.Reset(ctx); err != nil {
		t.Fatalf("reset: unexpected error: %v", err)
	}

	// Should be allowed again after reset
	if allow, err := l.Allow(ctx); err != nil || !allow {
		t.Fatalf("after reset: got allow=%v err=%v, want allow=true err=nil", allow, err)
	}
}

func TestLimiter_MultipleKeysIsolation(t *testing.T) {
	db := memory.NewInMemoryDB()
	rule := Rule{
		Limit:  2,
		Window: time.Second,
	}

	// Create two limiters with different keys
	l1 := NewInMemoryLimiter(db, "key1", rule)
	l2 := NewInMemoryLimiter(db, "key2", rule)
	ctx := context.Background()

	// Up first limiter
	for i := 0; i < rule.Limit; i++ {
		allow, err := l1.Allow(ctx)
		if err != nil || !allow {
			t.Fatalf("l1 attempt %d: got allow=%v err=%v, want allow=true err=nil", i+1, allow, err)
		}
	}

	// First limiter should be blocked
	if allow, err := l1.Allow(ctx); err != nil || allow {
		t.Fatalf("l1 over-limit: got allow=%v err=%v, want allow=false err=nil", allow, err)
	}

	// Second limiter should still allow attempts
	for i := 0; i < rule.Limit; i++ {
		allow, err := l2.Allow(ctx)
		if err != nil || !allow {
			t.Fatalf("l2 attempt %d: got allow=%v err=%v, want allow=true err=nil", i+1, allow, err)
		}
	}
}
