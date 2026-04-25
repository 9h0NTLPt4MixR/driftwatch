// Package labeler provides rule-based label assignment for drift results.
//
// Labels are lightweight string tags attached to a Result that can be used
// by downstream components (notifier, exporter, policy) to route or filter
// results without re-evaluating the full drift logic.
//
// Example usage:
//
//	rules := []labeler.Rule{
//		{Label: "critical", KeyPrefix: "auth.", MinDrifted: 1},
//		{Label: "payments", Service: "payment-svc"},
//	}
//	labeled := labeler.Apply(results, rules)
//	critical := labeler.FilterByLabel(labeled, "critical")
package labeler
