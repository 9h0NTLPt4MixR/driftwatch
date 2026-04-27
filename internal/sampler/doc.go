// Package sampler provides configurable sampling of drift scan results.
//
// Three strategies are supported:
//
//   - random  — selects results using a seeded PRNG; reproducible across runs
//               when the same seed is provided.
//   - modulo  — selects every Nth result based on the inverse of the rate;
//               useful for uniform, deterministic thinning.
//   - head    — returns the first N results proportional to the rate;
//               useful for quick previews of large result sets.
//
// HashSample provides an alternative deterministic approach that hashes each
// service name, making per-service membership stable across independent runs
// without requiring a shared seed.
package sampler
