// Package normalizer provides utilities for normalizing drift result values
// before comparison or display, including type coercion and whitespace trimming.
package normalizer

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Options controls how normalization is applied.
type Options struct {
	// TrimSpace removes leading/trailing whitespace from string values.
	TrimSpace bool
	// LowercaseKeys normalizes all config keys to lowercase.
	LowercaseKeys bool
	// CoerceBooleans converts "true"/"false" string values to canonical form.
	CoerceBooleans bool
}

// Apply normalizes a slice of drift results according to the given options.
// It returns a new slice; the originals are not modified.
func Apply(results []drift.Result, opts Options) []drift.Result {
	normalized := make([]drift.Result, 0, len(results))
	for _, r := range results {
		normalized = append(normalized, normalizeResult(r, opts))
	}
	return normalized
}

func normalizeResult(r drift.Result, opts Options) drift.Result {
	out := drift.Result{
		Service:  r.Service,
		Endpoint: r.Endpoint,
		Drifted:  r.Drifted,
		ScannedAt: r.ScannedAt,
		Diffs:    make([]drift.Diff, 0, len(r.Diffs)),
	}
	for _, d := range r.Diffs {
		out.Diffs = append(out.Diffs, normalizeDiff(d, opts))
	}
	return out
}

func normalizeDiff(d drift.Diff, opts Options) drift.Diff {
	key := d.Key
	if opts.LowercaseKeys {
		key = strings.ToLower(key)
	}
	want := normalizeValue(d.Want, opts)
	got := normalizeValue(d.Got, opts)
	return drift.Diff{Key: key, Want: want, Got: got}
}

func normalizeValue(v string, opts Options) string {
	if opts.TrimSpace {
		v = strings.TrimSpace(v)
	}
	if opts.CoerceBooleans {
		switch strings.ToLower(v) {
		case "true", "1", "yes":
			v = "true"
		case "false", "0", "no":
			v = "false"
		}
	}
	return v
}
