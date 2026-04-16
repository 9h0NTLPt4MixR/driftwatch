// Package baseline provides functionality for saving and comparing
// configuration baselines for driftwatch services.
//
// A baseline captures the expected and actual configuration state of each
// service at a point in time. Subsequent scans can be compared against a
// saved baseline to surface only new or changed drift, rather than all
// currently divergent keys.
//
// Typical usage:
//
//	// Save a baseline after a known-good deployment:
//	_ = baseline.Save("baseline.json", scanResults)
//
//	// Later, compare live state against the baseline:
//	base, _ := baseline.Load("baseline.json")
//	diffs := baseline.Compare(base, newScanResults)
package baseline
