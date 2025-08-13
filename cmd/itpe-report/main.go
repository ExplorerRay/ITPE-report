package main

import (
	"time"

	"github.com/explorerray/itpe-report/config"
	"github.com/explorerray/itpe-report/internal/promclient"
)

func main() {
	c := config.ParseArgs()

	promclient.Init(c.PrometheusURL)
	promclient.Query("kepler_container_platform_joules_total{container_name=\"ollama\"}", time.Now().Add(-1*time.Hour))
	promclient.Query("kepler_node_platform_joules_total", time.Now().Add(-1*time.Hour))
}
