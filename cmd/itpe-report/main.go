package main

import (
	"os"

	"github.com/explorerray/itpe-report/config"
	"github.com/explorerray/itpe-report/internal/client/promclient"
	"github.com/explorerray/itpe-report/internal/exporter/plot"
	"github.com/explorerray/itpe-report/internal/input"
	"github.com/explorerray/itpe-report/internal/logger"
)

func main() {
	logger := logger.NewLogger(logger.LogLevel(), os.Stdout)
	c := config.ParseArgsAndConfig(logger)

	// Initialize Prometheus client
	if err := promclient.Init(c.PrometheusURL); err != nil {
		logger.Error("Failed to initialize Prometheus client", "error", err)
		os.Exit(1)
	}

	// Read & parse GenAIperf json, then generate experiment metrics mapping
	emp := input.GenExpMetricPair(*c)
	logger.Info("Experiment metrics parsed")
	// Gen plots into png
	plotDir := plot.CreatePlotsSubdir(*c)
	if err := plot.GeneratePlots(emp, plotDir); err != nil {
		logger.Error("Failed to generate plots", "error", err)
		os.Exit(1)
	}
}
