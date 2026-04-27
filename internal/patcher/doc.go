// Package patcher provides patch operations that mutate or suppress specific
// drifted keys within a set of drift results.
//
// Supported actions:
//
//   - suppress: removes the diff entry entirely, marking the key as patched.
//   - override: replaces the Actual value with a caller-supplied value and
//     re-evaluates whether drift still exists.
//
// Patches can be scoped to a specific service or applied globally across all
// services when the Service field is left empty.
//
// Original drift.Result values are never modified; Apply always returns new
// patcher.Result wrappers with patch metadata attached.
package patcher
