package resolver_test

import (
	"testing"

	"github.com/driftwatch/internal/resolver"
)

func TestResolve_FromBase(t *testing.T) {
	r := resolver.New(map[string]string{"api": "http://api.local"})
	ep, err := r.Resolve("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep != "http://api.local" {
		t.Errorf("got %q, want %q", ep, "http://api.local")
	}
}

func TestResolve_EnvOverridesBase(t *testing.T) {
	t.Setenv("DRIFTWATCH_API_URL", "http://env.api.local")
	r := resolver.New(map[string]string{"api": "http://api.local"})
	ep, err := r.Resolve("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep != "http://env.api.local" {
		t.Errorf("got %q, want %q", ep, "http://env.api.local")
	}
}

func TestResolve_ProgrammaticOverrideWins(t *testing.T) {
	t.Setenv("DRIFTWATCH_API_URL", "http://env.api.local")
	r := resolver.New(map[string]string{"api": "http://api.local"})
	r.Override("api", "http://override.api.local")
	ep, err := r.Resolve("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep != "http://override.api.local" {
		t.Errorf("got %q, want %q", ep, "http://override.api.local")
	}
}

func TestResolve_UnknownService(t *testing.T) {
	r := resolver.New(nil)
	_, err := r.Resolve("unknown")
	if err == nil {
		t.Fatal("expected error for unknown service, got nil")
	}
}

func TestResolve_HyphenatedServiceName(t *testing.T) {
	t.Setenv("DRIFTWATCH_MY_SERVICE_URL", "http://hyphen.local")
	r := resolver.New(nil)
	ep, err := r.Resolve("my-service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep != "http://hyphen.local" {
		t.Errorf("got %q, want %q", ep, "http://hyphen.local")
	}
}

func TestResolveAll_ReturnsBothSources(t *testing.T) {
	r := resolver.New(map[string]string{
		"alpha": "http://alpha.local",
		"beta":  "http://beta.local",
	})
	r.Override("gamma", "http://gamma.local")

	entries := r.ResolveAll()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestResolveAll_SourceLabels(t *testing.T) {
	r := resolver.New(map[string]string{"api": "http://api.local"})
	r.Override("api", "http://override.local")

	entries := r.ResolveAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Source != "override" {
		t.Errorf("expected source 'override', got %q", entries[0].Source)
	}
}
