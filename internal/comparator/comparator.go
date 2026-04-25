// Package comparator provides multi-strategy comparison of live vs declared config values.
package comparator

import (
	"fmt"
	"strings"
)

// Strategy defines how two values are compared.
type Strategy string

const (
	StrategyExact  Strategy = "exact"
	StrategyPrefix Strategy = "prefix"
	StrategyIgnore Strategy = "ignore"
)

// Rule maps a key pattern to a comparison strategy.
type Rule struct {
	KeyPattern string   `json:"key_pattern"`
	Strategy   Strategy `json:"strategy"`
}

// Result holds the outcome of comparing a single key.
type Result struct {
	Key      string
	Expected string
	Actual   string
	Drifted  bool
	Reason   string
}

// Comparator applies per-key strategies to detect drift.
type Comparator struct {
	rules []Rule
}

// New creates a Comparator with the given rules.
func New(rules []Rule) *Comparator {
	return &Comparator{rules: rules}
}

// Compare evaluates expected vs actual for a set of keys and returns per-key results.
func (c *Comparator) Compare(expected, actual map[string]string) []Result {
	keys := unionKeys(expected, actual)
	results := make([]Result, 0, len(keys))
	for _, k := range keys {
		exp := expected[k]
		act := actual[k]
		strategy := c.strategyFor(k)
		r := applyStrategy(k, exp, act, strategy)
		results = append(results, r)
	}
	return results
}

func (c *Comparator) strategyFor(key string) Strategy {
	for _, rule := range c.rules {
		if strings.Contains(key, rule.KeyPattern) {
			return rule.Strategy
		}
	}
	return StrategyExact
}

func applyStrategy(key, expected, actual string, s Strategy) Result {
	switch s {
	case StrategyIgnore:
		return Result{Key: key, Expected: expected, Actual: actual, Drifted: false, Reason: "ignored"}
	case StrategyPrefix:
		if strings.HasPrefix(actual, expected) {
			return Result{Key: key, Expected: expected, Actual: actual, Drifted: false}
		}
		return Result{Key: key, Expected: expected, Actual: actual, Drifted: true,
			Reason: fmt.Sprintf("actual %q does not start with expected prefix %q", actual, expected)}
	default: // exact
		if expected == actual {
			return Result{Key: key, Expected: expected, Actual: actual, Drifted: false}
		}
		return Result{Key: key, Expected: expected, Actual: actual, Drifted: true,
			Reason: fmt.Sprintf("expected %q but got %q", expected, actual)}
	}
}

func unionKeys(a, b map[string]string) []string {
	seen := make(map[string]struct{})
	var keys []string
	for k := range a {
		if _, ok := seen[k]; !ok {
			keys = append(keys, k)
			seen[k] = struct{}{}
		}
	}
	for k := range b {
		if _, ok := seen[k]; !ok {
			keys = append(keys, k)
			seen[k] = struct{}{}
		}
	}
	return keys
}
