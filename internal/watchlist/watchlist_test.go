package watchlist_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/driftwatch/internal/watchlist"
)

func tmpPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "watchlist.json")
}

func TestAdd_And_List(t *testing.T) {
	p := tmpPath(t)
	if err := watchlist.Add(p, "svc-a", "high churn", 2); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := watchlist.Add(p, "svc-b", "critical path", 5); err != nil {
		t.Fatalf("Add: %v", err)
	}
	entries, err := watchlist.List(p)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Service != "svc-b" {
		t.Errorf("expected svc-b first (higher priority), got %s", entries[0].Service)
	}
}

func TestAdd_UpdatesExisting(t *testing.T) {
	p := tmpPath(t)
	_ = watchlist.Add(p, "svc-a", "initial", 1)
	_ = watchlist.Add(p, "svc-a", "updated", 3)
	entries, _ := watchlist.List(p)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after update, got %d", len(entries))
	}
	if entries[0].Reason != "updated" {
		t.Errorf("expected updated reason, got %s", entries[0].Reason)
	}
	if entries[0].Priority != 3 {
		t.Errorf("expected priority 3, got %d", entries[0].Priority)
	}
}

func TestRemove(t *testing.T) {
	p := tmpPath(t)
	_ = watchlist.Add(p, "svc-a", "reason", 1)
	_ = watchlist.Add(p, "svc-b", "reason", 2)
	if err := watchlist.Remove(p, "svc-a"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	entries, _ := watchlist.List(p)
	if len(entries) != 1 || entries[0].Service != "svc-b" {
		t.Errorf("expected only svc-b remaining")
	}
}

func TestList_MissingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.json")
	entries, err := watchlist.List(p)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d", len(entries))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(p, []byte("not-json"), 0644)
	_, err := watchlist.List(p)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
