// Package classifier categorises drift results by severity level
// based on the number of drifted keys and the presence of protected keys.
package classifier

import (
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Level represents a drift severity classification.
type Level string

const (
	LevelNone     Level = "none"
	LevelLow      Level = "low"
	LevelMedium   Level = "medium"
	LevelHigh     Level = "high"
	LevelCritical Level = "critical"
)

// ClassifiedResult pairs a drift result with its computed severity level.
type ClassifiedResult struct {
	drift.Result
	Level Level
}

// Options controls classification thresholds and protected key list.
type Options struct {
	// ProtectedKeys are key names whose drift always elevates to Critical.
	ProtectedKeys []string
	// LowThreshold is the minimum drifted-key count to be considered Low (default 1).
	LowThreshold int
	// MediumThreshold triggers Medium classification (default 3).
	MediumThreshold int
	// HighThreshold triggers High classification (default 6).
	HighThreshold int
}

func (o Options) withDefaults() Options {
	if o.LowThreshold == 0 {
		o.LowThreshold = 1
	}
	if o.MediumThreshold == 0 {
		o.MediumThreshold = 3
	}
	if o.HighThreshold == 0 {
		o.HighThreshold = 6
	}
	return o
}

// Classify assigns a severity Level to each drift.Result.
func Classify(results []drift.Result, opts Options) []ClassifiedResult {
	opts = opts.withDefaults()
	protected := make(map[string]bool, len(opts.ProtectedKeys))
	for _, k := range opts.ProtectedKeys {
		protected[k] = true
	}

	out := make([]ClassifiedResult, 0, len(results))
	for _, r := range results {
		out = append(out, ClassifiedResult{
			Result: r,
			Level:  classify(r, protected, opts),
		})
	}
	return out
}

// SortByLevel returns a copy of classified results ordered Critical → None.
func SortByLevel(results []ClassifiedResult) []ClassifiedResult {
	copy_ := make([]ClassifiedResult, len(results))
	copy(copy_, results)
	order := map[Level]int{LevelCritical: 0, LevelHigh: 1, LevelMedium: 2, LevelLow: 3, LevelNone: 4}
	sort.SliceStable(copy_, func(i, j int) bool {
		return order[copy_[i].Level] < order[copy_[j].Level]
	})
	return copy_
}

func classify(r drift.Result, protected map[string]bool, opts Options) Level {
	if !r.HasDrift {
		return LevelNone
	}
	for _, d := range r.Diffs {
		if protected[d.Key] {
			return LevelCritical
		}
	}
	n := len(r.Diffs)
	switch {
	case n >= opts.HighThreshold:
		return LevelHigh
	case n >= opts.MediumThreshold:
		return LevelMedium
	default:
		return LevelLow
	}
}
