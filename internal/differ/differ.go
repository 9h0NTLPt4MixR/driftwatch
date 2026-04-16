// Package differ provides utilities for computing human-readable diffs
// between two configuration snapshots.
package differ

import (
	"fmt"
	"sort"
	"strings"

	"github.com/user/driftwatch/internal/drift"
)

// DiffLine represents a single line in a diff output.
type DiffLine struct {
	Key      string
	Old      string
	New      string
	ChangeType string // "added", "removed", "changed"
}

// Diff holds the full diff for a single service.
type Diff struct {
	Service string
	Lines   []DiffLine
}

// Compute returns a slice of Diff entries derived from drift scan results.
func Compute(results []drift.Result) []Diff {
	var diffs []Diff
	for _, r := range results {
		if !r.HasDrift {
			continue
		}
		d := Diff{Service: r.Service}
		keys := sortedKeys(r.Diffs)
		for _, k := range keys {
			delta := r.Diffs[k]
			line := DiffLine{Key: k, Old: delta.Expected, New: delta.Actual}
			switch {
			case delta.Expected == "":
				line.ChangeType = "added"
			case delta.Actual == "":
				line.ChangeType = "removed"
			default:
				line.ChangeType = "changed"
			}
			d.Lines = append(d.Lines, line)
		}
		diffs = append(diffs, d)
	}
	return diffs
}

// Format renders a Diff slice as a unified-style text string.
func Format(diffs []Diff) string {
	var sb strings.Builder
	for _, d := range diffs {
		fmt.Fprintf(&sb, "--- %s\n", d.Service)
		for _, l := range d.Lines {
			switch l.ChangeType {
			case "added":
				fmt.Fprintf(&sb, "  + %s: %s\n", l.Key, l.New)
			case "removed":
				fmt.Fprintf(&sb, "  - %s: %s\n", l.Key, l.Old)
			default:
				fmt.Fprintf(&sb, "  ~ %s: %s -> %s\n", l.Key, l.Old, l.New)
			}
		}
	}
	return sb.String()
}

func sortedKeys(m map[string]drift.Delta) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
