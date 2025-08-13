package config

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
)

type Config struct {
	PrometheusURL string
}

func DefaultConfig() *Config {
	return &Config{
		PrometheusURL: "http://localhost:9090",
	}
}

func ParseArgs() *Config {
	const appName = "itpe-report"
	app := kingpin.New(appName, "ITPE report tool - Used to fetch power info from existing exporter and generate report for perf & energy")

	config := DefaultConfig()

	app.Flag("prom-url", "Prometheus URL").StringVar(&config.PrometheusURL)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	return config
}
