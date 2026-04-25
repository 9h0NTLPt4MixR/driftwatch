package throttle

import (
	"testing"
	"time"
)

func TestNew_ValidConfig(t *testing.T) {
	th, err := New(Config{Delay: 100 * time.Millisecond, Burst: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th == nil {
		t.Fatal("expected non-nil Throttle")
	}
}

func TestNew_NegativeDelay(t *testing.T) {
	_, err := New(Config{Delay: -1 * time.Millisecond, Burst: 1})
	if err == nil {
		t.Fatal("expected error for negative delay")
	}
}

func TestNew_ZeroBurst(t *testing.T) {
	_, err := New(Config{Delay: 0, Burst: 0})
	if err == nil {
		t.Fatal("expected error for zero burst")
	}
}

func TestWait_WithinBurst_NoSleep(t *testing.T) {
	th, _ := New(Config{Delay: 10 * time.Second, Burst: 3})
	slept := false
	th.sleepFn = func(d time.Duration) { slept = true }

	th.Wait()
	th.Wait()
	th.Wait()

	if slept {
		t.Error("expected no sleep within burst window")
	}
}

func TestWait_ExceedsBurst_Sleeps(t *testing.T) {
	th, _ := New(Config{Delay: 50 * time.Millisecond, Burst: 2})
	sleepCount := 0
	th.sleepFn = func(d time.Duration) { sleepCount++ }

	th.Wait() // 1 — within burst
	th.Wait() // 2 — within burst
	th.Wait() // 3 — exceeds burst, should sleep
	th.Wait() // 4 — exceeds burst, should sleep

	if sleepCount != 2 {
		t.Errorf("expected 2 sleeps, got %d", sleepCount)
	}
}

func TestReset_ClearsCount(t *testing.T) {
	th, _ := New(Config{Delay: 50 * time.Millisecond, Burst: 1})
	sleptAfterReset := false
	th.sleepFn = func(d time.Duration) { sleptAfterReset = true }

	th.Wait() // count=1, within burst
	th.Reset()
	th.Wait() // count=1 again, should not sleep

	if sleptAfterReset {
		t.Error("expected no sleep after reset")
	}
}

func TestStats(t *testing.T) {
	th, _ := New(Config{Delay: 0, Burst: 5})
	th.sleepFn = func(d time.Duration) {}

	th.Wait()
	th.Wait()

	count, burst := th.Stats()
	if count != 2 {
		t.Errorf("expected count=2, got %d", count)
	}
	if burst != 5 {
		t.Errorf("expected burst=5, got %d", burst)
	}
}
