// Package policy evaluates drift results against user-defined rules,
// flagging violations when drift exceeds configured thresholds or
// touches protected keys.
package policy

import (
	"fmt"
	"strings"

	"github.com/user/driftwatch/internal/drift"
)

// Rule defines a single policy constraint.
type Rule struct {
	// Name is a human-readable label for the rule.
	Name string `yaml:"name" json:"name"`
	// ProtectedKeys lists config keys that must never drift.
	ProtectedKeys []string `yaml:"protected_keys" json:"protected_keys"`
	// MaxDriftPercent is the maximum allowed percentage of drifted keys (0-100).
	// A value of 0 means any drift is a violation.
	MaxDriftPercent float64 `yaml:"max_drift_percent" json:"max_drift_percent"`
	// Services restricts the rule to specific service names; empty means all.
	Services []string `yaml:"services" json:"services"`
}

// Violation describes a policy rule that was breached.
type Violation struct {
	Rule    string
	Service string
	Reason  string
}

func (v Violation) String() string {
	return fmt.Sprintf("[%s] %s: %s", v.Rule, v.Service, v.Reason)
}

// Evaluate checks all results against the provided rules and returns any
// violations found. A nil or empty slice means no violations.
func Evaluate(results []drift.Result, rules []Rule) []Violation {
	var violations []Violation
	for _, rule := range rules {
		for _, result := range results {
			if !appliesToService(rule, result.Service) {
				continue
			}
			violations = append(violations, checkProtectedKeys(rule, result)...)
			if v := checkDriftPercent(rule, result); v != nil {
				violations = append(violations, *v)
			}
		}
	}
	return violations
}

func appliesToService(rule Rule, service string) bool {
	if len(rule.Services) == 0 {
		return true
	}
	for _, s := range rule.Services {
		if strings.EqualFold(s, service) {
			return true
		}
	}
	return false
}

func checkProtectedKeys(rule Rule, result drift.Result) []Violation {
	var vs []Violation
	for _, key := range rule.ProtectedKeys {
		for _, d := range result.Diffs {
			if strings.EqualFold(d.Key, key) {
				vs = append(vs, Violation{
					Rule:    rule.Name,
					Service: result.Service,
					Reason:  fmt.Sprintf("protected key %q has drifted (expected %q, got %q)", key, d.Expected, d.Actual),
				})
			}
		}
	}
	return vs
}

func checkDriftPercent(rule Rule, result drift.Result) *Violation {
	if len(result.Expected) == 0 {
		return nil
	}
	pct := float64(len(result.Diffs)) / float64(len(result.Expected)) * 100
	if pct > rule.MaxDriftPercent {
		return &Violation{
			Rule:    rule.Name,
			Service: result.Service,
			Reason:  fmt.Sprintf("drift %.1f%% exceeds allowed %.1f%%", pct, rule.MaxDriftPercent),
		}
	}
	return nil
}
