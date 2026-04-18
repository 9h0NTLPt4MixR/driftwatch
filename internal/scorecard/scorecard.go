// Package scorecard computes a drift health score for scanned services.
package scorecard

import (
	"fmt"
	"math"

	"github.com/example/driftwatch/internal/drift"
)

// Grade represents a letter grade for drift health.
type Grade string

const (
	GradeA Grade = "A"
	GradeB Grade = "B"
	GradeC Grade = "C"
	GradeD Grade = "D"
	GradeF Grade = "F"
)

// Score holds the computed scorecard for a set of results.
type Score struct {
	Total     int     `json:"total"`
	Drifted   int     `json:"drifted"`
	Clean     int     `json:"clean"`
	Pct       float64 `json:"health_pct"`
	Grade     Grade   `json:"grade"`
	DriftKeys int     `json:"total_drift_keys"`
}

// Compute calculates a Score from a slice of drift results.
func Compute(results []drift.Result) Score {
	total := len(results)
	if total == 0 {
		return Score{Grade: GradeA, Pct: 100.0}
	}

	drifted := 0
	driftKeys := 0
	for _, r := range results {
		if r.HasDrift {
			drifted++
			driftKeys += len(r.Diffs)
		}
	}

	clean := total - drifted
	pct := math.Round((float64(clean)/float64(total))*10000) / 100

	return Score{
		Total:     total,
		Drifted:   drifted,
		Clean:     clean,
		Pct:       pct,
		Grade:     grade(pct),
		DriftKeys: driftKeys,
	}
}

// String returns a human-readable summary line.
func (s Score) String() string {
	return fmt.Sprintf("Health: %.2f%% (%d/%d clean) — Grade: %s | Drifted keys: %d",
		s.Pct, s.Clean, s.Total, s.Grade, s.DriftKeys)
}

func grade(pct float64) Grade {
	switch {
	case pct >= 95:
		return GradeA
	case pct >= 80:
		return GradeB
	case pct >= 65:
		return GradeC
	case pct >= 50:
		return GradeD
	default:
		return GradeF
	}
}
