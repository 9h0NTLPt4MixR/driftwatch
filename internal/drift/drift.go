package drift

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/user/driftwatch/internal/config"
)

// DriftResult captures the comparison outcome for a single service.
type DriftResult struct {
	Service  string
	Drifted  bool
	Details  []string
	Error    string
}

// liveState is a minimal representation of what the service endpoint reports.
type liveState struct {
	Image    string            `json:"image"`
	Replicas int               `json:"replicas"`
	Env      map[string]string `json:"env"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Scan compares each declared service against its live state.
func Scan(services map[string]config.Service, verbose bool) ([]DriftResult, error) {
	var results []DriftResult
	for name, svc := range services {
		result := compare(name, svc, verbose)
		results = append(results, result)
	}
	return results, nil
}

func compare(name string, declared config.Service, verbose bool) DriftResult {
	res := DriftResult{Service: name}

	resp, err := httpClient.Get(declared.Endpoint + "/_driftwatch/state")
	if err != nil {
		res.Error = fmt.Sprintf("unreachable: %v", err)
		res.Drifted = true
		return res
	}
	defer resp.Body.Close()

	var live liveState
	if err := json.NewDecoder(resp.Body).Decode(&live); err != nil {
		res.Error = fmt.Sprintf("decode response: %v", err)
		res.Drifted = true
		return res
	}

	if declared.Image != "" && live.Image != declared.Image {
		res.Details = append(res.Details, fmt.Sprintf("image: declared=%s live=%s", declared.Image, live.Image))
	}
	if declared.Replicas > 0 && live.Replicas != declared.Replicas {
		res.Details = append(res.Details, fmt.Sprintf("replicas: declared=%d live=%d", declared.Replicas, live.Replicas))
	}
	for k, v := range declared.Env {
		if live.Env[k] != v {
			res.Details = append(res.Details, fmt.Sprintf("env.%s: declared=%s live=%s", k, v, live.Env[k]))
		}
	}

	res.Drifted = len(res.Details) > 0
	return res
}

// HasDrift returns true if any result contains drift.
func HasDrift(results []DriftResult) bool {
	for _, r := range results {
		if r.Drifted {
			return true
		}
	}
	return false
}

// Report writes scan results to w in the requested format.
func Report(results []DriftResult, format string, w io.Writer) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	default:
		for _, r := range results {
			status := "OK"
			if r.Drifted {
				status = "DRIFT"
			}
			fmt.Fprintf(w, "[%s] %s\n", status, r.Service)
			if r.Error != "" {
				fmt.Fprintf(w, "  error: %s\n", r.Error)
			}
			for _, d := range r.Details {
				fmt.Fprintf(w, "  - %s\n", d)
			}
		}
		return nil
	}
}
