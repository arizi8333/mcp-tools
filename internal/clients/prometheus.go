package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// PrometheusClient defines the interface for interacting with a Prometheus server.
type PrometheusClient interface {
	Query(ctx context.Context, query string) (string, error)
}

// RealPrometheus is the production client wrapper using the official Prometheus API.
type RealPrometheus struct {
	client v1.API
}

// NewRealPrometheus creates a new production Prometheus client.
func NewRealPrometheus(address string) (PrometheusClient, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}
	return &RealPrometheus{client: v1.NewAPI(client)}, nil
}

func (p *RealPrometheus) Query(ctx context.Context, query string) (string, error) {
	result, _, err := p.client.Query(ctx, query, time.Now())
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

// MockPrometheus is a simulation client for development and demo purposes.
type MockPrometheus struct {
	Endpoint string
}

func (m *MockPrometheus) Query(ctx context.Context, query string) (string, error) {
	return fmt.Sprintf("[Mock Response] Query: %s | Source: %s", query, m.Endpoint), nil
}
