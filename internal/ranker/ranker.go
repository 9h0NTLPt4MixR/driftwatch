// Package ranker ranks drift scan results by severity, scoring each service
// based on the number of drifted keys and whether protected keys are affected.
package ranker

import (
	"sort"

	"github.com/driftwatch/internal/drift"
)

// RankedResult wraps a drift result with a computed severity score.
type RankedResult struct {
	drift.Result
	Score    int
	Severity string
}

// Options controls how ranking is computed.
type Options struct {
	// ProtectedKeys are weighted more heavily in the score.
	ProtectedKeys []string
	// ProtectedWeight is the multiplier applied to protected key drifts (default 3).
	ProtectedWeight int
}

// Rank scores and sorts results from most to least severe.
// Results with no drift receive a score of 0 and severity "clean".
func Rank(results []drift.Result, opts Options) []RankedResult {
	if opts.ProtectedWeight <= 0 {
		opts.ProtectedWeight = 3
	}
	protected := make(map[string]bool, len(opts.ProtectedKeys))
	for _, k := range opts.ProtectedKeys {
		protected[k] = true
	}

	ranked := make([]RankedResult, 0, len(results))
	for _, r := range results {
		score := computeScore(r, protected, opts.ProtectedWeight)
		ranked = append(ranked, RankedResult{
			Result:   r,
			Score:    score,
			Severity: severity(score),
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})
	return ranked
}

func computeScore(r drift.Result, protected map[string]bool, weight int) int {
	if !r.HasDrift {
		return 0
	}
	score := 0
	for _, d := range r.Diffs {
		if protected[d.Key] {
			score += weight
		} else {
			score++
		}
	}
	return score
}

func severity(score int) string {
	switch {
	case score == 0:
		return "clean"
	case score <= 3:
		return "low"
	case score <= 7:
		return "medium"
	case score <= 14:
		return "high"
	default:
		return "critical"
	}
}
