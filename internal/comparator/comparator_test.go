package comparator_test

import (
	"testing"

	"github.com/driftwatch/internal/comparator"
)

func TestCompare_ExactMatch_NoDrift(t *testing.T) {
	c := comparator.New(nil)
	results := c.Compare(
		map[string]string{"PORT": "8080"},
		map[string]string{"PORT": "8080"},
	)
	if len(results) != 1 || results[0].Drifted {
		t.Fatalf("expected no drift, got %+v", results)
	}
}

func TestCompare_ExactMismatch_Drift(t *testing.T) {
	c := comparator.New(nil)
	results := c.Compare(
		map[string]string{"PORT": "8080"},
		map[string]string{"PORT": "9090"},
	)
	if len(results) != 1 || !results[0].Drifted {
		t.Fatalf("expected drift, got %+v", results)
	}
}

func TestCompare_IgnoreStrategy_NoDrift(t *testing.T) {
	rules := []comparator.Rule{{KeyPattern: "timestamp", Strategy: comparator.StrategyIgnore}}
	c := comparator.New(rules)
	results := c.Compare(
		map[string]string{"timestamp": "2024-01-01"},
		map[string]string{"timestamp": "2025-06-01"},
	)
	if len(results) != 1 || results[0].Drifted {
		t.Fatalf("expected ignored key to have no drift, got %+v", results)
	}
	if results[0].Reason != "ignored" {
		t.Errorf("expected reason 'ignored', got %q", results[0].Reason)
	}
}

func TestCompare_PrefixStrategy_Match(t *testing.T) {
	rules := []comparator.Rule{{KeyPattern: "IMAGE", Strategy: comparator.StrategyPrefix}}
	c := comparator.New(rules)
	results := c.Compare(
		map[string]string{"IMAGE": "myapp"},
		map[string]string{"IMAGE": "myapp:v1.2.3"},
	)
	if len(results) != 1 || results[0].Drifted {
		t.Fatalf("expected prefix match, got %+v", results)
	}
}

func TestCompare_PrefixStrategy_Mismatch(t *testing.T) {
	rules := []comparator.Rule{{KeyPattern: "IMAGE", Strategy: comparator.StrategyPrefix}}
	c := comparator.New(rules)
	results := c.Compare(
		map[string]string{"IMAGE": "myapp"},
		map[string]string{"IMAGE": "otherapp:v1"},
	)
	if len(results) != 1 || !results[0].Drifted {
		t.Fatalf("expected drift on prefix mismatch, got %+v", results)
	}
}

func TestCompare_MissingKey_CountedAsDrift(t *testing.T) {
	c := comparator.New(nil)
	results := c.Compare(
		map[string]string{"PORT": "8080", "HOST": "localhost"},
		map[string]string{"PORT": "8080"},
	)
	drifted := 0
	for _, r := range results {
		if r.Drifted {
			drifted++
		}
	}
	if drifted != 1 {
		t.Errorf("expected 1 drifted key, got %d", drifted)
	}
}

func TestCompare_MultipleRules_FirstMatchWins(t *testing.T) {
	rules := []comparator.Rule{
		{KeyPattern: "SECRET", Strategy: comparator.StrategyIgnore},
		{KeyPattern: "SECRET", Strategy: comparator.StrategyExact},
	}
	c := comparator.New(rules)
	results := c.Compare(
		map[string]string{"SECRET_KEY": "abc"},
		map[string]string{"SECRET_KEY": "xyz"},
	)
	if results[0].Drifted {
		t.Error("first matching rule (ignore) should have prevented drift")
	}
}
