// Package grouper provides functionality for grouping drift results
// by arbitrary dimensions such as environment, team, or region.
package grouper

import (
	"fmt"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// GroupBy defines the dimension to group results by.
type GroupBy string

const (
	GroupByService GroupBy = "service"
	GroupByStatus  GroupBy = "status"
	GroupByTag     GroupBy = "tag"
)

// Group holds a named collection of drift results.
type Group struct {
	Name    string
	Results []drift.Result
}

// Options configures grouping behaviour.
type Options struct {
	By      GroupBy
	TagKey  string // used when By == GroupByTag
}

// Compute groups the provided results according to opts.
// Results that do not match any group bucket are placed under "(untagged)".
func Compute(results []drift.Result, opts Options) []Group {
	buckets := make(map[string][]drift.Result)

	for _, r := range results {
		key := bucketKey(r, opts)
		buckets[key] = append(buckets[key], r)
	}

	groups := make([]Group, 0, len(buckets))
	for name, res := range buckets {
		groups = append(groups, Group{Name: name, Results: res})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	return groups
}

// Summary returns a human-readable summary line per group.
func Summary(groups []Group) []string {
	lines := make([]string, 0, len(groups))
	for _, g := range groups {
		drifted := 0
		for _, r := range g.Results {
			if r.HasDrift {
				drifted++
			}
		}
		lines = append(lines, fmt.Sprintf("%s: %d/%d drifted", g.Name, drifted, len(g.Results)))
	}
	return lines
}

func bucketKey(r drift.Result, opts Options) string {
	switch opts.By {
	case GroupByStatus:
		if r.HasDrift {
			return "drifted"
		}
		return "clean"
	case GroupByTag:
		if v, ok := r.Tags[opts.TagKey]; ok && v != "" {
			return v
		}
		return "(untagged)"
	default: // GroupByService
		return r.Service
	}
}
