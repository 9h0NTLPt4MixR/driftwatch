// Package ranker provides severity-based ranking of drift scan results.
//
// Results are scored by counting drifted keys, with optional weighted scoring
// for protected (sensitive) keys. Scores map to severity labels:
//
//   - clean    (score == 0)
//   - low      (score 1–3)
//   - medium   (score 4–7)
//   - high     (score 8–14)
//   - critical (score >= 15)
//
// Usage:
//
//	ranked := ranker.Rank(results, ranker.Options{
//		ProtectedKeys:   []string{"db_password", "api_secret"},
//		ProtectedWeight: 5,
//	})
package ranker
