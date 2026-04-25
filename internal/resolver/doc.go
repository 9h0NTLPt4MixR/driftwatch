// Package resolver provides endpoint resolution for driftwatch services.
//
// Resolution follows a three-tier precedence model:
//
//  1. Programmatic overrides set via Resolver.Override — highest priority,
//     useful for tests or one-off CLI flags.
//  2. Environment variables of the form DRIFTWATCH_<SERVICE>_URL, where the
//     service name is upper-cased and hyphens are replaced with underscores.
//  3. Base endpoints declared in the driftwatch configuration file.
//
// Example:
//
//	r := resolver.New(cfg.EndpointMap())
//	ep, err := r.Resolve("payments-api")
package resolver
