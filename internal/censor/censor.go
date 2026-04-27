// Package censor masks or removes fields from drift results based on
// configurable patterns before results are written to any output sink.
package censor

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Options controls which fields are censored.
type Options struct {
	// Patterns is a list of substring patterns; any key containing one of
	// these patterns (case-insensitive) will be censored.
	Patterns []string
	// Replacement is the string substituted for censored values.
	// Defaults to "***" when empty.
	Replacement string
}

const defaultReplacement = "***"

// Apply returns a new slice of results with matching keys censored.
// The original results are not modified.
func Apply(results []drift.Result, opts Options) []drift.Result {
	replacement := opts.Replacement
	if replacement == "" {
		replacement = defaultReplacement
	}

	out := make([]drift.Result, len(results))
	for i, r := range results {
		out[i] = censorResult(r, opts.Patterns, replacement)
	}
	return out
}

func censorResult(r drift.Result, patterns []string, replacement string) drift.Result {
	copy := r
	copy.Diffs = make([]drift.Diff, 0, len(r.Diffs))
	for _, d := range r.Diffs {
		if matchesAny(d.Key, patterns) {
			d.Expected = replacement
			d.Actual = replacement
		}
		copy.Diffs = append(copy.Diffs, d)
	}
	return copy
}

func matchesAny(key string, patterns []string) bool {
	lower := strings.ToLower(key)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
