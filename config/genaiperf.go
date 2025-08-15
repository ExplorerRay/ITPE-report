package config

type (
	TokenConf struct {
		Name   string `yaml:"name"`
		Mean   int    `yaml:"mean"`
		Stddev int    `yaml:"stddev"`
	}

	TokenConfs struct {
		Input  []TokenConf `yaml:"input"`
		Output []TokenConf `yaml:"output"`
	}

	Enabled struct {
		Stream     bool `yaml:"stream"`
		Checkpoint bool `yaml:"checkpoint"`
	}

	Requests struct {
		RunCount []int `yaml:"run_count"`
	}

	GenAIPerf struct {
		EndpointURL string     `yaml:"url"`
		Enabled     Enabled    `yaml:"enabled"`
		Models      []string   `yaml:"models"`
		Concurrency []int      `yaml:"concurrency"`
		Requests    Requests   `yaml:"requests"`
		TokenConfs  TokenConfs `yaml:"token_confs"`
	}
)
