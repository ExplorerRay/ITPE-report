package promclient

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// apiClient is a package-level variable to store the Prometheus API client
var apiClient v1.API

func Init(prometheus_url string) {
	client, err := api.NewClient(api.Config{
		Address: prometheus_url,
	})

	if err != nil {
		panic(err)
	}

	apiClient = v1.NewAPI(client)
}

func Query(name string, timestamp time.Time) {
	if apiClient == nil {
		fmt.Println("Prometheus API client is not initialized")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := apiClient.Query(ctx, name, timestamp, v1.WithTimeout(2*time.Second))
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	// Handle vector metric
	fmt.Println("Vector result:")
	for _, sample := range result.(model.Vector) {
		fmt.Printf("  Metric: %v, Value: %v, Timestamp: %v\n",
			sample.Metric, sample.Value, sample.Timestamp)
	}
}
