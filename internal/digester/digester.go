// Package digester computes deterministic fingerprints (digests) for drift
// scan results, enabling change detection between successive scans.
package digester

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Result holds the digest for a single service scan result.
type Result struct {
	Service string `json:"service"`
	Digest  string `json:"digest"`
	Drifted bool   `json:"drifted"`
}

// Compute generates a SHA-256 digest for each drift.Result in the slice.
// Results are fingerprinted by service name, drift status, and sorted diffs
// so that the digest is stable regardless of map iteration order.
func Compute(results []drift.Result) ([]Result, error) {
	out := make([]Result, 0, len(results))
	for _, r := range results {
		d, err := digest(r)
		if err != nil {
			return nil, fmt.Errorf("digester: service %q: %w", r.Service, err)
		}
		out = append(out, Result{
			Service: r.Service,
			Digest:  d,
			Drifted: r.HasDrift,
		})
	}
	return out, nil
}

// Changed returns the service names whose digest differs between prev and next.
func Changed(prev, next []Result) []string {
	prevMap := make(map[string]string, len(prev))
	for _, r := range prev {
		prevMap[r.Service] = r.Digest
	}
	var changed []string
	for _, r := range next {
		if old, ok := prevMap[r.Service]; !ok || old != r.Digest {
			changed = append(changed, r.Service)
		}
	}
	sort.Strings(changed)
	return changed
}

// digest produces a stable SHA-256 hex string for a single drift.Result.
func digest(r drift.Result) (string, error) {
	type stable struct {
		Service  string          `json:"service"`
		HasDrift bool            `json:"has_drift"`
		Diffs    []drift.Diff    `json:"diffs"`
	}
	sorted := make([]drift.Diff, len(r.Diffs))
	copy(sorted, r.Diffs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})
	s := stable{Service: r.Service, HasDrift: r.HasDrift, Diffs: sorted}
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
