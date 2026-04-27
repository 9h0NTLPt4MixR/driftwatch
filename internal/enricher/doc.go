// Package enricher provides utilities for attaching external metadata to
// drift scan results. Metadata such as environment name, service owner, and
// tier classification can be supplied via an Options struct and are merged
// onto each drift.Result to produce an enricher.Result.
//
// Typical usage:
//
//	opts := enricher.Options{
//		DefaultEnvironment: "production",
//		ServiceMeta: map[string]enricher.Meta{
//			"api-gateway": {Owner: "platform", Tier: "critical"},
//		},
//	}
//	enriched := enricher.Apply(results, opts)
package enricher
