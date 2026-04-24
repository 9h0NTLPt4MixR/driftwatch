// Package audit provides a structured audit log for drift scan events,
// recording who triggered a scan, when, and what was found.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp  time.Time `json:"timestamp"`
	User       string    `json:"user"`
	Command    string    `json:"command"`
	TotalSvcs  int       `json:"total_services"`
	DriftCount int       `json:"drift_count"`
	Services   []string  `json:"services"`
}

// Record appends a new audit entry to the given file path.
func Record(path, user, command string, results []drift.Result) error {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		User:      user,
		Command:   command,
		TotalSvcs: len(results),
	}

	for _, r := range results {
		if r.HasDrift {
			entry.DriftCount++
		}
		entry.Services = append(entry.Services, r.Service)
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit: open file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", line)
	if err != nil {
		return fmt.Errorf("audit: write entry: %w", err)
	}
	return nil
}

// LoadAll reads all audit entries from the given file path.
func LoadAll(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("audit: read file: %w", err)
	}

	var entries []Entry
	decoder := json.NewDecoder(
		// wrap bytes in a reader line-by-line
		newLineReader(data),
	)
	for decoder.More() {
		var e Entry
		if err := decoder.Decode(&e); err != nil {
			return nil, fmt.Errorf("audit: decode entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// newLineReader wraps a byte slice so json.Decoder can stream NDJSON.
func newLineReader(data []byte) *os.File {
	// Use a pipe trick via bytes.Reader — return as io.Reader via a temp approach.
	// Since os.File is needed for the decoder, we write to a temp file.
	tmp, _ := os.CreateTemp("", "audit-read-*")
	_ = os.WriteFile(tmp.Name(), data, 0o600)
	f, _ := os.Open(tmp.Name())
	_ = os.Remove(tmp.Name())
	return f
}
