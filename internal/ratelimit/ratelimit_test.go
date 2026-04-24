package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/ratelimit"
)

func TestNew_InvalidRate(t *testing.T) {
	_, err := ratelimit.New(0, time.Second)
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
}

func TestNew_InvalidWindow(t *testing.T) {
	_, err := ratelimit.New(1, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestAllow_UnderLimit(t *testing.T) {
	l, err := ratelimit.New(3, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 3; i++ {
		if !l.Allow("svc-a") {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	l, _ := ratelimit.New(2, time.Minute)

	l.Allow("svc-b")
	l.Allow("svc-b")

	if l.Allow("svc-b") {
		t.Fatal("third call should be denied")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	l, _ := ratelimit.New(1, time.Minute)

	if !l.Allow("svc-x") {
		t.Fatal("first call for svc-x should be allowed")
	}
	if !l.Allow("svc-y") {
		t.Fatal("first call for svc-y should be allowed")
	}
	if l.Allow("svc-x") {
		t.Fatal("second call for svc-x should be denied")
	}
}

func TestRemaining_DecreasesWithCalls(t *testing.T) {
	l, _ := ratelimit.New(5, time.Minute)

	if got := l.Remaining("svc-c"); got != 5 {
		t.Fatalf("expected 5 remaining, got %d", got)
	}

	l.Allow("svc-c")
	l.Allow("svc-c")

	if got := l.Remaining("svc-c"); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	l, _ := ratelimit.New(1, 50*time.Millisecond)

	if !l.Allow("svc-d") {
		t.Fatal("first call should be allowed")
	}
	if l.Allow("svc-d") {
		t.Fatal("second call within window should be denied")
	}

	time.Sleep(60 * time.Millisecond)

	if !l.Allow("svc-d") {
		t.Fatal("call after window expiry should be allowed")
	}
}
