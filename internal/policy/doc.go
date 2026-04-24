// Package policy provides rule-based evaluation of drift scan results.
//
// Rules can restrict which services are evaluated, flag drift on protected
// configuration keys, and enforce a maximum drift percentage threshold.
//
// Example usage:
//
//	rules := []policy.Rule{
//		{
//			Name:            "no-secret-drift",
//			ProtectedKeys:   []string{"DB_PASSWORD", "API_KEY"},
//			MaxDriftPercent: 0,
//		},
//	}
//	violations := policy.Evaluate(results, rules)
//	for _, v := range violations {
//		fmt.Println(v)
//	}
package policy
