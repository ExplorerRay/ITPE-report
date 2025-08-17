package config

// metricConfig defines plot settings for a metric
type metricConfig struct {
	YLabel   string
	Filename string
}

type YPlot struct {
	// perf
	RequestThroughput     float64
	OutputTokenThroughput float64
	AvgRequestLatencyMs   float64
	AvgTTFTMs             float64
	AvgITLMs              float64
	// power
	NodePlatformJ  float64
	NodeGPUJ       float64
	NodeCPUJ       float64
	EnergyPerToken float64
}

// GetMetricsConfig returns the configuration for all metrics
func GetMetricsConfig() map[string]metricConfig {
	return map[string]metricConfig{
		"Request Throughput": {
			YLabel:   "Requests per Second",
			Filename: "req_throughput",
		},
		"Output Token Throughput": {
			YLabel:   "Tokens per Second",
			Filename: "out_token_throughput",
		},
		"Avg Request Latency": {
			YLabel:   "Milliseconds",
			Filename: "avg_req_latency",
		},
		"Avg TTFT": {
			YLabel:   "Milliseconds",
			Filename: "avg_ttft",
		},
		"Avg ITL": {
			YLabel:   "Milliseconds",
			Filename: "avg_itl",
		},
		"Node Platform": {
			YLabel:   "Joules",
			Filename: "node_pltf_energy",
		},
		"Node GPU": {
			YLabel:   "Joules",
			Filename: "node_gpu_energy",
		},
		"Node CPU": {
			YLabel:   "Joules",
			Filename: "node_cpu_energy",
		},
		"Energy Per Token": {
			YLabel:   "Joules per Token",
			Filename: "energy_per_token",
		},
	}
}
