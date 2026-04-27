// Package sampler provides deterministic sampling of drift results
// based on configurable rate and strategy (random, modulo, or head).
package sampler

import (
	"fmt"
	"hash/fnv"
	"math/rand"

	"github.com/driftwatch/internal/drift"
)

// Strategy controls how results are selected.
type Strategy string

const (
	StrategyRandom  Strategy = "random"
	StrategyModulo  Strategy = "modulo"
	StrategyHead    Strategy = "head"
)

// Options configures the sampler.
type Options struct {
	// Rate is the fraction of results to keep (0.0–1.0).
	Rate     float64
	Strategy Strategy
	// Seed is used for reproducible random sampling.
	Seed     int64
}

// Sample returns a subset of results according to opts.
// It returns an error if opts.Rate is outside [0, 1].
func Sample(results []drift.Result, opts Options) ([]drift.Result, error) {
	if opts.Rate < 0 || opts.Rate > 1 {
		return nil, fmt.Errorf("sampler: rate must be between 0.0 and 1.0, got %f", opts.Rate)
	}
	if len(results) == 0 {
		return results, nil
	}

	switch opts.Strategy {
	case StrategyHead:
		return sampleHead(results, opts.Rate), nil
	case StrategyModulo:
		return sampleModulo(results, opts.Rate), nil
	default:
		return sampleRandom(results, opts.Rate, opts.Seed), nil
	}
}

func sampleHead(results []drift.Result, rate float64) []drift.Result {
	n := int(float64(len(results)) * rate)
	if n > len(results) {
		n = len(results)
	}
	return results[:n]
}

func sampleModulo(results []drift.Result, rate float64) []drift.Result {
	step := 1
	if rate < 1.0 && rate > 0 {
		step = int(1.0 / rate)
	}
	out := make([]drift.Result, 0)
	for i, r := range results {
		if i%step == 0 {
			out = append(out, r)
		}
	}
	return out
}

func sampleRandom(results []drift.Result, rate float64, seed int64) []drift.Result {
	rng := rand.New(rand.NewSource(seed))
	out := make([]drift.Result, 0, len(results))
	for _, r := range results {
		if rng.Float64() < rate {
			out = append(out, r)
		}
	}
	return out
}

// HashSample selects results deterministically by hashing the service name.
// It is useful for stable per-service sampling across runs.
func HashSample(results []drift.Result, rate float64) ([]drift.Result, error) {
	if rate < 0 || rate > 1 {
		return nil, fmt.Errorf("sampler: rate must be between 0.0 and 1.0, got %f", rate)
	}
	threshold := uint32(rate * float64(^uint32(0)))
	out := make([]drift.Result, 0, len(results))
	for _, r := range results {
		h := fnv.New32a()
		_, _ = h.Write([]byte(r.Service))
		if h.Sum32() <= threshold {
			out = append(out, r)
		}
	}
	return out, nil
}
