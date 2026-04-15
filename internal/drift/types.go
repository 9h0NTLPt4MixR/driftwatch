package drift

// Diff represents a single field-level discrepancy between expected and actual state.
type Diff struct {
	// Key is the configuration field name that differs.
	Key string
	// Expected is the declared value from version control.
	Expected string
	// Actual is the value observed from the live service.
	Actual string
}

// Result holds the drift analysis outcome for a single service.
type Result struct {
	// Service is the name of the service that was scanned.
	Service string
	// Endpoint is the URL that was queried.
	Endpoint string
	// Diffs contains all detected discrepancies. Empty means no drift.
	Diffs []Diff
	// Error holds any fetch or parse error encountered during scanning.
	Error error
}

// IsDrifted returns true when the result contains at least one diff
// and no fetch error occurred.
func (r Result) IsDrifted() bool {
	return r.Error == nil && len(r.Diffs) > 0
}

// IsHealthy returns true when the service was reachable and matches
// its declared configuration exactly.
func (r Result) IsHealthy() bool {
	return r.Error == nil && len(r.Diffs) == 0
}
