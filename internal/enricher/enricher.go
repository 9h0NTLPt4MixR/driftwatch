// Package enricher attaches metadata to drift results from external or
// computed sources, such as environment labels, owner annotations, and
// service tier classifications.
package enricher

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Meta holds enrichment fields added to a result.
type Meta struct {
	Environment string            `json:"environment,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Tier        string            `json:"tier,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Options controls which enrichment sources are applied.
type Options struct {
	// ServiceMeta maps service name to a Meta struct.
	ServiceMeta map[string]Meta
	// DefaultEnvironment is used when no per-service environment is defined.
	DefaultEnvironment string
}

// Result wraps a drift result with enrichment metadata.
type Result struct {
	drift.Result
	Meta Meta `json:"meta"`
}

// Apply enriches each drift.Result with metadata from opts and returns a
// slice of enriched Results.
func Apply(results []drift.Result, opts Options) []Result {
	out := make([]Result, 0, len(results))
	for _, r := range results {
		meta := resolveMeta(r.Service, opts)
		out = append(out, Result{Result: r, Meta: meta})
	}
	return out
}

// FilterByEnvironment returns only those enriched results whose environment
// matches env (case-insensitive). If env is empty all results are returned.
func FilterByEnvironment(results []Result, env string) []Result {
	if env == "" {
		return results
	}
	env = strings.ToLower(env)
	out := make([]Result, 0)
	for _, r := range results {
		if strings.ToLower(r.Meta.Environment) == env {
			out = append(out, r)
		}
	}
	return out
}

// FilterByOwner returns only those enriched results whose owner matches
// the given string (case-insensitive). Empty owner returns all results.
func FilterByOwner(results []Result, owner string) []Result {
	if owner == "" {
		return results
	}
	owner = strings.ToLower(owner)
	out := make([]Result, 0)
	for _, r := range results {
		if strings.ToLower(r.Meta.Owner) == owner {
			out = append(out, r)
		}
	}
	return out
}

func resolveMeta(service string, opts Options) Meta {
	if m, ok := opts.ServiceMeta[service]; ok {
		if m.Environment == "" {
			m.Environment = opts.DefaultEnvironment
		}
		return m
	}
	return Meta{Environment: opts.DefaultEnvironment}
}
