package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/driftwatch/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	yaml := `
version: "1"
services:
  api:
    name: api
    endpoint: http://api.example.com
    image: api:latest
    replicas: 2
    env:
      LOG_LEVEL: info
`
	path := writeTemp(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(cfg.Services))
	}
	svc := cfg.Services["api"]
	if svc.Replicas != 2 {
		t.Errorf("expected replicas=2, got %d", svc.Replicas)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoServices(t *testing.T) {
	path := writeTemp(t, "version: \"1\"\nservices: {}\n")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty services")
	}
}

func TestLoad_MissingEndpoint(t *testing.T) {
	yaml := `version: "1"
services:
  worker:
    name: worker
    image: worker:latest
`
	path := writeTemp(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing endpoint")
	}
}
