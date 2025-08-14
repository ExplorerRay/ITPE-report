package promclient

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// apiClient is a package-level variable to store the Prometheus API client
var apiClient v1.API

// QueryInfo holds query details
type QueryInfo struct {
	Name      string
	Timestamp time.Time
}

// QueryResult holds the result of a single Prometheus query
type QueryResult struct {
	Name      string            // Query name
	Metric    model.Metric      // Metric labels (e.g., {container_name="ollama"})
	Value     model.SampleValue // Sample value
	Timestamp model.Time        // Sample timestamp
}

// QueryResponse holds the full response for a query
type QueryResponse struct {
	Results  []QueryResult // List of results for the query
	Warnings []string      // Any warnings from the query
	Error    error         // Any error from the query
}

// Init initializes the Prometheus API client with the given URL
func Init(prometheus_url string) error {
	client, err := api.NewClient(api.Config{
		Address: prometheus_url,
	})

	if err != nil {
		return err
	}

	apiClient = v1.NewAPI(client)
	return nil
}

func MakeKeplerQueryInfo(timestamp time.Time, containerName string) []QueryInfo {
	baseQueryList := []string{
		"kepler_node_platform_joules_total",
		"kepler_node_gpu_joules_total",
		"kepler_node_package_joules_total",
		"kepler_node_dram_joules_total",
		"kepler_node_other_joules_total",
		"kepler_container_gpu_joules_total{container_name=\"" + containerName + "\"}",
		"kepler_container_dram_joules_total{container_name=\"" + containerName + "\"}",
		"kepler_container_package_joules_total{container_name=\"" + containerName + "\"}",
		"kepler_container_platform_joules_total{container_name=\"" + containerName + "\"}",
		"kepler_container_other_joules_total{container_name=\"" + containerName + "\"}",
	}
	var queryInfos []QueryInfo
	for _, q := range baseQueryList {
		queryInfos = append(queryInfos, QueryInfo{Name: q, Timestamp: timestamp})
	}
	return queryInfos
}

func SumResults(results []QueryResult) float64 {
	var sum float64
	for _, res := range results {
		sum += float64(res.Value)
	}
	return sum
}

// Query executes a single Prometheus query and returns structured results
func Query(name string, timestamp time.Time) QueryResponse {
	if apiClient == nil {
		return QueryResponse{Error: fmt.Errorf("prometheus API client is not initialized")}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := apiClient.Query(ctx, name, timestamp, v1.WithTimeout(2*time.Second))
	if err != nil {
		return QueryResponse{Error: fmt.Errorf("querying Prometheus for %s: %v", name, err)}
	}

	response := QueryResponse{
		Warnings: warnings,
	}

	// Handle vector metric
	if vector, ok := result.(model.Vector); ok {
		if len(vector) == 0 {
			return response // Empty results
		}
		response.Results = make([]QueryResult, len(vector))
		for i, sample := range vector {
			response.Results[i] = QueryResult{
				Name:      name,
				Metric:    sample.Metric,
				Value:     sample.Value,
				Timestamp: sample.Timestamp,
			}
		}
	} else {
		// Non-vector results (e.g., Scalar, String) not expected, return empty
		return response
	}

	return response
}

// MultiQuery executes multiple Prometheus queries and returns structured results
func MultiQuery(queries []QueryInfo) []QueryResponse {
	if apiClient == nil {
		return []QueryResponse{{
			Error: fmt.Errorf("prometheus API client is not initialized"),
		}}
	}

	responses := make([]QueryResponse, len(queries))
	for i, q := range queries {
		responses[i] = Query(q.Name, q.Timestamp)
	}
	return responses
}
