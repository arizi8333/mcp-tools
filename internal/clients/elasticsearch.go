package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ElasticsearchClient defines the interface for Elasticsearch operations.
type ElasticsearchClient interface {
	Search(ctx context.Context, index string, query map[string]interface{}) ([]string, error)
	Ping(ctx context.Context) error
}

// RealElasticsearch implements ElasticsearchClient for real Elasticsearch.
type RealElasticsearch struct {
	BaseURL    string
	HTTPClient *http.Client
	Username   string
	Password   string
	APIKey     string // API Key for authentication (preferred over username/password)
}

// NewRealElasticsearch creates a new Elasticsearch client.
func NewRealElasticsearch(baseURL, username, password, apiKey string) (*RealElasticsearch, error) {
	return &RealElasticsearch{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Username: username,
		Password: password,
		APIKey:   apiKey,
	}, nil
}

// Ping checks if Elasticsearch is reachable.
func (e *RealElasticsearch) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", e.BaseURL, nil)
	if err != nil {
		return err
	}

	// Priority 1: API Key authentication (more secure)
	if e.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", e.APIKey))
	} else if e.Username != "" {
		// Priority 2: Basic Auth fallback
		req.SetBasicAuth(e.Username, e.Password)
	}

	resp, err := e.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("elasticsearch ping failed: status %d", resp.StatusCode)
	}

	return nil
}

// Search executes a search query against Elasticsearch.
func (e *RealElasticsearch) Search(ctx context.Context, index string, query map[string]interface{}) ([]string, error) {
	url := fmt.Sprintf("%s/%s/_search", e.BaseURL, index)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Priority 1: API Key authentication (more secure)
	if e.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", e.APIKey))
	} else if e.Username != "" {
		// Priority 2: Basic Auth fallback
		req.SetBasicAuth(e.Username, e.Password)
	}

	resp, err := e.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elasticsearch search failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var logs []string
	for _, hit := range result.Hits.Hits {
		// Try to extract message field, fallback to full JSON
		if msg, ok := hit.Source["message"].(string); ok {
			logs = append(logs, msg)
		} else {
			jsonBytes, _ := json.Marshal(hit.Source)
			logs = append(logs, string(jsonBytes))
		}
	}

	return logs, nil
}

// MockElasticsearch implements ElasticsearchClient for testing.
type MockElasticsearch struct {
	BaseURL string
}

func (m *MockElasticsearch) Ping(ctx context.Context) error {
	return nil
}

func (m *MockElasticsearch) Search(ctx context.Context, index string, query map[string]interface{}) ([]string, error) {
	return []string{
		"[MOCK] 2024-01-15T10:23:45Z ERROR Connection timeout to database",
		"[MOCK] 2024-01-15T10:24:12Z ERROR Failed to process request: invalid token",
		"[MOCK] 2024-01-15T10:25:33Z WARN High memory usage detected: 85%",
	}, nil
}
