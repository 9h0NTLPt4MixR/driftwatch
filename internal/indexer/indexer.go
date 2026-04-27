// Package indexer builds an in-memory index of drift results for fast
// lookup by service name, drift status, or key.
package indexer

import (
	"fmt"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Index holds pre-computed lookup structures over a slice of drift results.
type Index struct {
	ByService map[string]*drift.Result
	Drifted   []*drift.Result
	Clean     []*drift.Result
	ByKey     map[string][]*drift.Result
}

// Build constructs an Index from the provided results.
func Build(results []drift.Result) (*Index, error) {
	if results == nil {
		return nil, fmt.Errorf("indexer: results must not be nil")
	}

	idx := &Index{
		ByService: make(map[string]*drift.Result),
		ByKey:     make(map[string][]*drift.Result),
	}

	for i := range results {
		r := &results[i]
		key := strings.ToLower(r.Service)
		idx.ByService[key] = r

		if r.Drifted {
			idx.Drifted = append(idx.Drifted, r)
		} else {
			idx.Clean = append(idx.Clean, r)
		}

		for _, d := range r.Diffs {
			k := strings.ToLower(d.Key)
			idx.ByKey[k] = append(idx.ByKey[k], r)
		}
	}

	return idx, nil
}

// LookupService returns the result for a given service name (case-insensitive).
func (idx *Index) LookupService(name string) (*drift.Result, bool) {
	r, ok := idx.ByService[strings.ToLower(name)]
	return r, ok
}

// LookupKey returns all results that contain a drift diff for the given key.
func (idx *Index) LookupKey(key string) []*drift.Result {
	return idx.ByKey[strings.ToLower(key)]
}

// Stats returns a brief summary string of the index contents.
func (idx *Index) Stats() string {
	return fmt.Sprintf("services=%d drifted=%d clean=%d keys=%d",
		len(idx.ByService), len(idx.Drifted), len(idx.Clean), len(idx.ByKey))
}
