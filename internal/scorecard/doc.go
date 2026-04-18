// Package scorecard provides a health scoring system for configuration drift results.
//
// It computes a percentage of clean (non-drifted) services and assigns a letter
// grade (A–F) based on configurable thresholds. The scorecard is useful for
// surfacing overall drift health in CI pipelines and summary reports.
package scorecard
