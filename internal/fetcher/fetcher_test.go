package fetcher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetch_Success(t *testing.T) {
	expected := map[string]interface{}{"replicas": float64(3), "image": "nginx:1.25"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := New(5 * time.Second)
	state, err := client.Fetch("svc-a", ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.ServiceName != "svc-a" {
		t.Errorf("expected service name 'svc-a', got '%s'", state.ServiceName)
	}
	if state.Values["image"] != "nginx:1.25" {
		t.Errorf("expected image 'nginx:1.25', got '%v'", state.Values["image"])
	}
}

func TestFetch_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := New(5 * time.Second)
	_, err := client.Fetch("svc-b", ts.URL)
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestFetch_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	client := New(5 * time.Second)
	_, err := client.Fetch("svc-c", ts.URL)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestFetch_Unreachable(t *testing.T) {
	client := New(1 * time.Second)
	_, err := client.Fetch("svc-d", "http://127.0.0.1:19999/no-server")
	if err == nil {
		t.Fatal("expected error for unreachable endpoint, got nil")
	}
}
