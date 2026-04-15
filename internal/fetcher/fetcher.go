package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ServiceState represents the live configuration state fetched from a service endpoint.
type ServiceState struct {
	ServiceName string
	Values      map[string]interface{}
}

// Client is an HTTP client used to fetch live service state.
type Client struct {
	httpClient *http.Client
}

// New creates a new fetcher Client with a default timeout.
func New(timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Fetch retrieves the live configuration from the given endpoint URL
// and returns a ServiceState populated with the parsed JSON response.
func (c *Client) Fetch(serviceName, endpoint string) (*ServiceState, error) {
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetcher: GET %s failed: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetcher: unexpected status %d from %s", resp.StatusCode, endpoint)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fetcher: reading body from %s: %w", endpoint, err)
	}

	var values map[string]interface{}
	if err := json.Unmarshal(body, &values); err != nil {
		return nil, fmt.Errorf("fetcher: parsing JSON from %s: %w", endpoint, err)
	}

	return &ServiceState{
		ServiceName: serviceName,
		Values:      values,
	}, nil
}
