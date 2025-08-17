package main

import (
	"fmt"
	"os"

	"github.com/explorerray/itpe-report/config"
	"github.com/explorerray/itpe-report/internal/client/promclient"
	"github.com/explorerray/itpe-report/internal/exporter/plot"
	"github.com/explorerray/itpe-report/internal/input"
)

func main() {
	c := config.ParseArgsAndConfig()

	// Initialize Prometheus client
	if err := promclient.Init(c.PrometheusURL); err != nil {
		fmt.Printf("Failed to initialize Prometheus client: %v\n", err)
		os.Exit(1)
	}

	// Read & parse GenAIperf json, then generate experiment metrics mapping
	emp := input.GenExpMetricPair(*c)
	// Gen plots into png
	plotDir := plot.CreatePlotsSubdir(*c)
	if err := plot.GeneratePlots(emp, plotDir); err != nil {
		fmt.Printf("Failed to generate plots: %v\n", err)
		os.Exit(1)
	}
}
