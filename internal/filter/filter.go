package filter

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Options holds filtering criteria for drift results.
type Options struct {
	// OnlyDrifted filters results to only services with drift.
	OnlyDrifted bool
	// Services is an optional list of service names to include.
	// If empty, all services are included.
	Services []string
	// Keys is an optional list of config keys to restrict comparisons to.
	// If empty, all keys are included.
	Keys []string
}

// Apply filters a slice of drift.Result according to the provided Options.
func Apply(results []drift.Result, opts Options) []drift.Result {
	filtered := make([]drift.Result, 0, len(results))

	for _, r := range results {
		if !matchesService(r.Service, opts.Services) {
			continue
		}

		if len(opts.Keys) > 0 {
			r = filterKeys(r, opts.Keys)
		}

		if opts.OnlyDrifted && !r.HasDrift {
			continue
		}

		filtered = append(filtered, r)
	}

	return filtered
}

// matchesService returns true if name is in the allowed list, or the list is empty.
func matchesService(name string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, s := range allowed {
		if strings.EqualFold(s, name) {
			return true
		}
	}
	return false
}

// filterKeys returns a copy of r with Diffs restricted to the specified keys.
func filterKeys(r drift.Result, keys []string) drift.Result {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[strings.ToLower(k)] = struct{}{}
	}

	filtered := make([]drift.Diff, 0, len(r.Diffs))
	for _, d := range r.Diffs {
		if _, ok := keySet[strings.ToLower(d.Key)]; ok {
			filtered = append(filtered, d)
		}
	}

	r.Diffs = filtered
	r.HasDrift = len(filtered) > 0
	return r
}
