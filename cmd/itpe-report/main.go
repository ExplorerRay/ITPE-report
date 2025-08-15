package main

import (
	"fmt"
	"os"

	"github.com/explorerray/itpe-report/config"
	"github.com/explorerray/itpe-report/internal/client/promclient"
	"github.com/explorerray/itpe-report/internal/exporter/stdout"
	"github.com/explorerray/itpe-report/internal/input"
)

func main() {
	c := config.ParseArgsAndConfig()

	// Initialize Prometheus client
	if err := promclient.Init(c.PrometheusURL); err != nil {
		fmt.Printf("Failed to initialize Prometheus client: %v\n", err)
		os.Exit(1)
	}

	// Read and process GenAI-Perf JSON
	profile, err := input.ParseGenAIPerfJSON(c.GenAIProfPath)
	if err != nil {
		fmt.Printf("Failed to read profile-export.json: %v\n", err)
		return
	}

	// Compute and display metrics for the first experiment (extend for multiple if needed)
	if len(profile.Experiments) > 0 {
		metrics := input.ComputeMetrics(profile.Experiments[0])
		stdout.MetricsToTableOut(metrics)
	} else {
		fmt.Println("No experiments found in profile-export.json")
	}

	// Display metrics for power
	if len(profile.Experiments) > 0 {
		powerMetrics := input.GetPowerMetrics(profile.Experiments[0])
		stdout.PowerMetricsToTableOut(powerMetrics)
	} else {
		fmt.Println("No experiments found in profile-export.json")
	}
}
