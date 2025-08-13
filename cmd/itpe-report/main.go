package main

import (
	"fmt"
	"os"
	"time"

	"github.com/explorerray/itpe-report/config"
	"github.com/explorerray/itpe-report/internal/exporter/stdout"
	"github.com/explorerray/itpe-report/internal/promclient"
)

func main() {
	c := config.ParseArgs()

	// Initialize Prometheus client
	if err := promclient.Init(c.PrometheusURL); err != nil {
		fmt.Printf("Failed to initialize Prometheus client: %v\n", err)
		os.Exit(1)
	}

	// Define queries
	queries := []promclient.QueryInfo{
		{Name: "kepler_container_platform_joules_total{container_name=\"ollama\"}", Timestamp: time.Now().Add(-1 * time.Hour)},
		{Name: "kepler_container_dram_joules_total{container_name=\"ollama\"}", Timestamp: time.Now().Add(-1 * time.Hour)},
		{Name: "kepler_container_package_joules_total{container_name=\"ollama\"}", Timestamp: time.Now().Add(-1 * time.Hour)},
		{Name: "kepler_container_gpu_joules_total{container_name=\"ollama\"}", Timestamp: time.Now().Add(-1 * time.Hour)},
		{Name: "kepler_node_platform_joules_total", Timestamp: time.Now().Add(-1 * time.Hour)},
	}

	// Run queries and get results
	responses := promclient.MultiQuery(queries)

	// Print out results in table
	stdout.QueryToTableOut(queries, responses)
}
