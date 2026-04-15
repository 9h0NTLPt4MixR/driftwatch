package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/driftwatch/internal/drift"
)

const defaultDir = ".driftwatch/snapshots"

// Entry represents a saved snapshot of drift results at a point in time.
type Entry struct {
	Timestamp time.Time           `json:"timestamp"`
	Results   []drift.Result      `json:"results"`
}

// Save writes drift results to a timestamped JSON file in the snapshot directory.
func Save(results []drift.Result, dir string) (string, error) {
	if dir == "" {
		dir = defaultDir
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("snapshot: create dir: %w", err)
	}

	entry := Entry{
		Timestamp: time.Now().UTC(),
		Results:   results,
	}

	filename := entry.Timestamp.Format("20060102T150405Z") + ".json"
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entry); err != nil {
		return "", fmt.Errorf("snapshot: encode: %w", err)
	}

	return path, nil
}

// Load reads a snapshot entry from the given file path.
func Load(path string) (*Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open: %w", err)
	}
	defer f.Close()

	var entry Entry
	if err := json.NewDecoder(f).Decode(&entry); err != nil {
		return nil, fmt.Errorf("snapshot: decode: %w", err)
	}
	return &entry, nil
}
