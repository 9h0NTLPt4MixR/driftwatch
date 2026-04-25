// Package comparator implements multi-strategy key comparison for drift detection.
//
// It supports three strategies:
//
//   - exact   (default): values must match exactly.
//   - prefix:            the live value must start with the declared value.
//   - ignore:            the key is excluded from drift evaluation entirely.
//
// Rules are matched by substring against the key name; the first matching rule wins.
// If no rule matches, the exact strategy is applied.
//
// Example usage:
//
//	rules := []comparator.Rule{
//		{KeyPattern: "IMAGE",     Strategy: comparator.StrategyPrefix},
//		{KeyPattern: "timestamp", Strategy: comparator.StrategyIgnore},
//	}
//	c := comparator.New(rules)
//	results := c.Compare(expected, actual)
package comparator
