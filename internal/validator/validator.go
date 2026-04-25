// Package validator checks drift results against user-defined thresholds
// and emits structured violations when limits are exceeded.
package validator

import (
	"fmt"

	"github.com/driftwatch/internal/drift"
)

// Rule defines a single validation constraint.
type Rule struct {
	// MaxDriftedServices is the maximum number of services allowed to have drift.
	// -1 means unlimited.
	MaxDriftedServices int `json:"max_drifted_services"`

	// MaxDriftedKeys is the maximum number of drifted keys allowed per service.
	// -1 means unlimited.
	MaxDriftedKeys int `json:"max_drifted_keys"`

	// RequireCleanServices lists service names that must have zero drift.
	RequireCleanServices []string `json:"require_clean_services"`
}

// Violation describes a single rule breach.
type Violation struct {
	Service string
	Message string
}

// Result holds the outcome of a validation run.
type Result struct {
	Violations []Violation
	Passed     bool
}

// Validate applies r against results and returns a Result.
func Validate(results []drift.Result, r Rule) Result {
	var violations []Violation

	// Count drifted services.
	if r.MaxDriftedServices >= 0 {
		drifted := 0
		for _, res := range results {
			if res.Drifted {
				drifted++
			}
		}
		if drifted > r.MaxDriftedServices {
			violations = append(violations, Violation{
				Service: "*",
				Message: fmt.Sprintf(
					"drifted service count %d exceeds maximum %d",
					drifted, r.MaxDriftedServices,
				),
			})
		}
	}

	// Check per-service key limits.
	if r.MaxDriftedKeys >= 0 {
		for _, res := range results {
			if len(res.Diffs) > r.MaxDriftedKeys {
				violations = append(violations, Violation{
					Service: res.Service,
					Message: fmt.Sprintf(
						"drifted key count %d exceeds maximum %d",
						len(res.Diffs), r.MaxDriftedKeys,
					),
				})
			}
		}
	}

	// Enforce required-clean services.
	index := make(map[string]drift.Result, len(results))
	for _, res := range results {
		index[res.Service] = res
	}
	for _, svc := range r.RequireCleanServices {
		if res, ok := index[svc]; ok && res.Drifted {
			violations = append(violations, Violation{
				Service: svc,
				Message: "service is required to be clean but has drift",
			})
		}
	}

	return Result{
		Violations: violations,
		Passed:     len(violations) == 0,
	}
}
