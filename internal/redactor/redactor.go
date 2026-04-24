// Package redactor provides utilities for masking sensitive values
// in drift scan results before they are displayed or exported.
package redactor

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// DefaultSensitiveKeys contains common key patterns considered sensitive.
var DefaultSensitiveKeys = []string{
	"password",
	"secret",
	"token",
	"apikey",
	"api_key",
	"private_key",
	"credentials",
	"auth",
}

const redactedValue = "[REDACTED]"

// Options configures redaction behaviour.
type Options struct {
	// SensitiveKeys overrides the default list of sensitive key patterns.
	// Matching is case-insensitive substring match.
	SensitiveKeys []string
}

// Apply masks sensitive values inside each DriftResult's Expected, Actual,
// and per-key Differences, returning a new slice of results.
func Apply(results []drift.DriftResult, opts Options) []drift.DriftResult {
	keys := opts.SensitiveKeys
	if len(keys) == 0 {
		keys = DefaultSensitiveKeys
	}

	out := make([]drift.DriftResult, len(results))
	for i, r := range results {
		out[i] = redactResult(r, keys)
	}
	return out
}

func redactResult(r drift.DriftResult, sensitiveKeys []string) drift.DriftResult {
	r.Expected = redactMap(r.Expected, sensitiveKeys)
	r.Actual = redactMap(r.Actual, sensitiveKeys)

	redacted := make([]drift.Diff, len(r.Diffs))
	for i, d := range r.Diffs {
		if isSensitive(d.Key, sensitiveKeys) {
			d.Expected = redactedValue
			d.Actual = redactedValue
		}
		redacted[i] = d
	}
	r.Diffs = redacted
	return r
}

func redactMap(m map[string]string, sensitiveKeys []string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if isSensitive(k, sensitiveKeys) {
			out[k] = redactedValue
		} else {
			out[k] = v
		}
	}
	return out
}

func isSensitive(key string, patterns []string) bool {
	lower := strings.ToLower(key)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
