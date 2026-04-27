// Package alerter provides threshold-based alerting for drift scan results.
// It evaluates a set of alert rules against scan results and emits alerts
// when configured thresholds are exceeded.
package alerter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo     Level = "INFO"
	LevelWarning  Level = "WARNING"
	LevelCritical Level = "CRITICAL"
)

// Rule defines a threshold-based alerting rule.
type Rule struct {
	// Name is a human-readable identifier for the rule.
	Name string `json:"name"`
	// ServiceFilter restricts the rule to a specific service name (empty = all).
	ServiceFilter string `json:"service_filter,omitempty"`
	// MinDriftedKeys triggers the alert when drifted key count meets or exceeds this value.
	MinDriftedKeys int `json:"min_drifted_keys,omitempty"`
	// MinDriftPercent triggers the alert when drift percentage meets or exceeds this value.
	MinDriftPercent float64 `json:"min_drift_percent,omitempty"`
	// Level is the severity assigned to alerts fired by this rule.
	Level Level `json:"level"`
}

// Alert represents a fired alert for a single scan result.
type Alert struct {
	Rule      string    `json:"rule"`
	Service   string    `json:"service"`
	Level     Level     `json:"level"`
	Message   string    `json:"message"`
	FiredAt   time.Time `json:"fired_at"`
}

// Evaluate checks each result against the provided rules and returns any fired alerts.
func Evaluate(results []drift.Result, rules []Rule) []Alert {
	var alerts []Alert
	for _, result := range results {
		for _, rule := range rules {
			if rule.ServiceFilter != "" && !strings.EqualFold(rule.ServiceFilter, result.Service) {
				continue
			}
			if !result.Drifted {
				continue
			}

			driftedKeys := len(result.Diffs)
			totalKeys := result.TotalKeys
			var driftPct float64
			if totalKeys > 0 {
				driftPct = float64(driftedKeys) / float64(totalKeys) * 100.0
			}

			fire := false
			var reason string

			if rule.MinDriftedKeys > 0 && driftedKeys >= rule.MinDriftedKeys {
				fire = true
				reason = fmt.Sprintf("%d drifted keys (threshold: %d)", driftedKeys, rule.MinDriftedKeys)
			} else if rule.MinDriftPercent > 0 && driftPct >= rule.MinDriftPercent {
				fire = true
				reason = fmt.Sprintf("%.1f%% drift (threshold: %.1f%%)", driftPct, rule.MinDriftPercent)
			}

			if fire {
				alerts = append(alerts, Alert{
					Rule:    rule.Name,
					Service: result.Service,
					Level:   rule.Level,
					Message: fmt.Sprintf("[%s] service %q triggered rule %q: %s", rule.Level, result.Service, rule.Name, reason),
					FiredAt: time.Now().UTC(),
				})
			}
		}
	}
	return alerts
}

// Write outputs fired alerts to w. If no alerts were fired, a short summary is written instead.
func Write(alerts []Alert, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(alerts) == 0 {
		fmt.Fprintln(w, "alerter: no alerts fired")
		return
	}
	for _, a := range alerts {
		fmt.Fprintf(w, "%s  %s\n", a.FiredAt.Format(time.RFC3339), a.Message)
	}
	fmt.Fprintf(w, "\n%d alert(s) fired.\n", len(alerts))
}
