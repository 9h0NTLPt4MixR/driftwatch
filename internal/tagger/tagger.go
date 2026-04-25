// Package tagger provides functionality for tagging drift results
// with user-defined labels for grouping, filtering, and reporting.
package tagger

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// TagMap maps service names to a list of string tags.
type TagMap map[string][]string

// Apply attaches tags to each DriftResult based on the provided TagMap.
// Results for services not present in the map are returned unmodified.
func Apply(results []drift.Result, tags TagMap) []drift.Result {
	out := make([]drift.Result, len(results))
	for i, r := range results {
		if t, ok := tags[r.Service]; ok {
			r.Tags = dedupe(t)
		}
		out[i] = r
	}
	return out
}

// FilterByTag returns only those results that carry the specified tag.
func FilterByTag(results []drift.Result, tag string) []drift.Result {
	var out []drift.Result
	for _, r := range results {
		for _, t := range r.Tags {
			if t == tag {
				out = append(out, r)
				break
			}
		}
	}
	return out
}

// LoadTagMap reads a JSON file mapping service names to tag slices.
func LoadTagMap(path string) (TagMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tagger: read file: %w", err)
	}
	var tm TagMap
	if err := json.Unmarshal(data, &tm); err != nil {
		return nil, fmt.Errorf("tagger: parse tag map: %w", err)
	}
	return tm, nil
}

func dedupe(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out
}
