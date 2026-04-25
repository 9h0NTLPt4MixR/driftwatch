// Package labeler assigns metadata labels to drift results based on
// configurable rules, enabling downstream filtering and routing.
package labeler

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Rule maps a label to a set of conditions that must all match.
type Rule struct {
	Label      string            `json:"label"`
	Service    string            `json:"service,omitempty"`
	KeyPrefix  string            `json:"key_prefix,omitempty"`
	MinDrifted int               `json:"min_drifted,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
}

// Apply evaluates each rule against the results and attaches matching
// labels. Labels are stored in Result.Labels (deduplicated).
func Apply(results []drift.Result, rules []Rule) []drift.Result {
	out := make([]drift.Result, len(results))
	for i, r := range results {
		for _, rule := range rules {
			if matches(r, rule) {
				r.Labels = appendUnique(r.Labels, rule.Label)
			}
		}
		out[i] = r
	}
	return out
}

// FilterByLabel returns only results that carry the given label.
func FilterByLabel(results []drift.Result, label string) []drift.Result {
	var out []drift.Result
	for _, r := range results {
		for _, l := range r.Labels {
			if l == label {
				out = append(out, r)
				break
			}
		}
	}
	return out
}

func matches(r drift.Result, rule Rule) bool {
	if rule.Service != "" && r.Service != rule.Service {
		return false
	}
	if rule.KeyPrefix != "" && !hasDriftedKeyWithPrefix(r, rule.KeyPrefix) {
		return false
	}
	if rule.MinDrifted > 0 && len(r.Diffs) < rule.MinDrifted {
		return false
	}
	return true
}

func hasDriftedKeyWithPrefix(r drift.Result, prefix string) bool {
	for _, d := range r.Diffs {
		if strings.HasPrefix(d.Key, prefix) {
			return true
		}
	}
	return false
}

func appendUnique(labels []string, label string) []string {
	for _, l := range labels {
		if l == label {
			return labels
		}
	}
	return append(labels, label)
}
