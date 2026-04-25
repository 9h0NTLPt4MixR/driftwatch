// Package resolver maps service names to their live endpoint URLs,
// supporting static configuration and environment-variable overrides.
package resolver

import (
	"fmt"
	"os"
	"strings"
)

// Entry holds resolution metadata for a single service.
type Entry struct {
	Service  string
	Endpoint string
	Source   string // "config" | "env" | "override"
}

// Resolver resolves service names to endpoints.
type Resolver struct {
	base      map[string]string // from config
	overrides map[string]string // programmatic overrides
}

// New creates a Resolver seeded with base endpoints from config.
func New(base map[string]string) *Resolver {
	if base == nil {
		base = make(map[string]string)
	}
	return &Resolver{
		base:      base,
		overrides: make(map[string]string),
	}
}

// Override sets a programmatic endpoint for a service, taking highest precedence.
func (r *Resolver) Override(service, endpoint string) {
	r.overrides[service] = endpoint
}

// Resolve returns the endpoint for the given service name.
// Precedence: programmatic override > ENV var (DRIFTWATCH_<SERVICE>_URL) > config base.
func (r *Resolver) Resolve(service string) (string, error) {
	if ep, ok := r.overrides[service]; ok {
		return ep, nil
	}
	envKey := "DRIFTWATCH_" + strings.ToUpper(strings.ReplaceAll(service, "-", "_")) + "_URL"
	if ep := os.Getenv(envKey); ep != "" {
		return ep, nil
	}
	if ep, ok := r.base[service]; ok {
		return ep, nil
	}
	return "", fmt.Errorf("resolver: no endpoint found for service %q", service)
}

// ResolveAll returns an Entry slice for every known service.
// Services discovered only via ENV are not included; only those present
// in the base map or override map are enumerated.
func (r *Resolver) ResolveAll() []Entry {
	seen := make(map[string]struct{})
	var entries []Entry

	add := func(service string) {
		if _, ok := seen[service]; ok {
			return
		}
		seen[service] = struct{}{}
		ep, err := r.Resolve(service)
		if err != nil {
			return
		}
		src := r.sourceFor(service)
		entries = append(entries, Entry{Service: service, Endpoint: ep, Source: src})
	}

	for svc := range r.overrides {
		add(svc)
	}
	for svc := range r.base {
		add(svc)
	}
	return entries
}

func (r *Resolver) sourceFor(service string) string {
	if _, ok := r.overrides[service]; ok {
		return "override"
	}
	envKey := "DRIFTWATCH_" + strings.ToUpper(strings.ReplaceAll(service, "-", "_")) + "_URL"
	if os.Getenv(envKey) != "" {
		return "env"
	}
	return "config"
}
