// Package pruner provides utilities for removing stale, clean, or explicitly
// excluded drift results from a result set.
//
// Use Prune for simple filtering or PruneWithStats when you need to report
// how many results were discarded and why.
package pruner
