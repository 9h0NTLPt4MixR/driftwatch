// Package patcher applies a set of patch operations to drift results,
// allowing callers to override, suppress, or annotate specific drifted keys
// before downstream processing.
package patcher

import (
	"strings"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Op describes a single patch operation.
type Op struct {
	// Service restricts the patch to a specific service name (empty = all).
	Service string
	// Key is the config key to target (exact match).
	Key string
	// Action is one of: "suppress", "override".
	Action string
	// Value is used when Action is "override".
	Value string
}

// Result wraps a drift.Result with patch metadata.
type Result struct {
	drift.Result
	PatchedAt time.Time
	PatchedKeys []string
}

// Apply runs the given patch ops against each drift result and returns
// patched copies. Original results are not modified.
func Apply(results []drift.Result, ops []Op) []Result {
	out := make([]Result, 0, len(results))
	for _, r := range results {
		pr := Result{Result: r, PatchedAt: time.Now()}
		pr.Diffs = patchDiffs(r.Service, r.Diffs, ops, &pr.PatchedKeys)
		pr.Drifted = len(pr.Diffs) > 0
		out = append(out, pr)
	}
	return out
}

func patchDiffs(service string, diffs []drift.Diff, ops []Op, patched *[]string) []drift.Diff {
	result := make([]drift.Diff, 0, len(diffs))
	for _, d := range diffs {
		kept, key := applyOps(service, d, ops)
		if !kept {
			*patched = append(*patched, key)
			continue
		}
		result = append(result, d)
	}
	return result
}

func applyOps(service string, d drift.Diff, ops []Op) (keep bool, suppressedKey string) {
	for _, op := range ops {
		if op.Service != "" && !strings.EqualFold(op.Service, service) {
			continue
		}
		if !strings.EqualFold(op.Key, d.Key) {
			continue
		}
		switch strings.ToLower(op.Action) {
		case "suppress":
			return false, d.Key
		case "override":
			d.Actual = op.Value
			d.Drifted = d.Actual != d.Expected
			return true, ""
		}
	}
	return true, ""
}
