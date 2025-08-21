package config

type ReportConf struct {
	PrometheusURL string `yaml:"prom_url"`
	ArtfDir       string `yaml:"artf_dir"`
}
