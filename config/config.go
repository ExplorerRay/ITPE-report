package config

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		PrometheusURL string
		GenAIPerf     GenAIPerf
		GenAIConfPath string
		GenAIProfPath string
		GenAIArtfDir  string
	}
)

func DefaultConfig() *Config {
	return &Config{
		PrometheusURL: "http://localhost:9090",
		GenAIConfPath: "config.yml",
		GenAIProfPath: "profile_export.json",
		GenAIArtfDir:  "/artifacts",
	}
}

func RegisterFlags(app *kingpin.Application, config *Config) {
	app.Flag("prom-url", "Prometheus URL").StringVar(&config.PrometheusURL)
	app.Flag("genai-conf-path", "Path to GenAI-Perf config.yml").StringVar(&config.GenAIConfPath)
	app.Flag("genai-prof-path", "Path to GenAI-Perf profile_export.json").StringVar(&config.GenAIProfPath)
	app.Flag("genai-artf-dir", "Path to GenAI-Perf artifacts directory").StringVar(&config.GenAIArtfDir)
}

func loadGenAIConf(path string) (*GenAIPerf, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML config file %s: %v", path, err)
	}

	var gaip GenAIPerf
	if err := yaml.Unmarshal(data, &gaip); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config file %s: %v", path, err)
	}

	return &gaip, nil
}

func ParseArgsAndConfig() *Config {
	const appName = "itpe-report"
	app := kingpin.New(appName, "ITPE report tool - Used to fetch power info from existing exporter and generate report for perf & energy")

	config := DefaultConfig()

	RegisterFlags(app, config)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	gaip, err := loadGenAIConf(config.GenAIConfPath)
	if err != nil {
		fmt.Printf("Warning: failed to load GenAIPerf config: %v\n", err)
		return config
	}
	config.GenAIPerf = *gaip

	return config
}
