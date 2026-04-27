// Package digester computes deterministic SHA-256 fingerprints for drift scan
// results.
//
// Use Compute to generate a []Result where each entry pairs a service name
// with its current digest. Use Changed to compare two digest snapshots and
// identify which services have changed between scans.
//
// Digests are stable: diffs are sorted by key before hashing so that
// insertion-order differences in maps do not produce false positives.
package digester
