package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		ConfigPath string
		ReportConf ReportConf `yaml:"itpe_report"`
		GenAIPerf  GenAIPerf  `yaml:"itpe_perf"`
	}
)

func DefaultConfig() *Config {
	return &Config{
		ConfigPath: "config.yaml",
		ReportConf: ReportConf{
			PrometheusURL: "http://localhost:9090",
			ArtfDir:       "/artifacts",
		},
		GenAIPerf: GenAIPerf{
			EndpointURL: "http://localhost:8000",
		},
	}
}

func RegisterFlags(app *kingpin.Application, config *Config) {
	app.Flag("config", "Path to config file").StringVar(&config.ConfigPath)
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %v", path, err)
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %v", path, err)
	}
	return &conf, nil
}

func ParseArgsAndConfig(logger *slog.Logger) *Config {
	const appName = "itpe-report"
	app := kingpin.New(appName, "ITPE report tool - Used to generate report for inference perf & energy")

	config := DefaultConfig()

	RegisterFlags(app, config)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	config, err := loadConfig(config.ConfigPath)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return DefaultConfig()
	}

	return config
}
