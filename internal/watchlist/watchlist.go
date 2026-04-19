// Package watchlist manages a prioritized list of services to monitor closely.
package watchlist

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// Entry represents a single watchlist entry.
type Entry struct {
	Service   string    `json:"service"`
	Reason    string    `json:"reason"`
	AddedAt   time.Time `json:"added_at"`
	Priority  int       `json:"priority"` // higher = more urgent
}

// Watchlist holds all monitored service entries.
type Watchlist struct {
	Entries []Entry `json:"entries"`
}

// Add inserts or updates a service entry in the watchlist.
func Add(path, service, reason string, priority int) error {
	wl, err := load(path)
	if err != nil {
		return err
	}
	for i, e := range wl.Entries {
		if e.Service == service {
			wl.Entries[i] = Entry{Service: service, Reason: reason, AddedAt: time.Now(), Priority: priority}
			return save(path, wl)
		}
	}
	wl.Entries = append(wl.Entries, Entry{
		Service:  service,
		Reason:   reason,
		AddedAt:  time.Now(),
		Priority: priority,
	})
	return save(path, wl)
}

// Remove deletes a service from the watchlist.
func Remove(path, service string) error {
	wl, err := load(path)
	if err != nil {
		return err
	}
	filtered := wl.Entries[:0]
	for _, e := range wl.Entries {
		if e.Service != service {
			filtered = append(filtered, e)
		}
	}
	wl.Entries = filtered
	return save(path, wl)
}

// List returns entries sorted by priority descending.
func List(path string) ([]Entry, error) {
	wl, err := load(path)
	if err != nil {
		return nil, err
	}
	sorted := make([]Entry, len(wl.Entries))
	copy(sorted, wl.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})
	return sorted, nil
}

func load(path string) (Watchlist, error) {
	var wl Watchlist
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return wl, nil
	}
	if err != nil {
		return wl, fmt.Errorf("watchlist: read: %w", err)
	}
	if err := json.Unmarshal(data, &wl); err != nil {
		return wl, fmt.Errorf("watchlist: parse: %w", err)
	}
	return wl, nil
}

func save(path string, wl Watchlist) error {
	data, err := json.MarshalIndent(wl, "", "  ")
	if err != nil {
		return fmt.Errorf("watchlist: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
