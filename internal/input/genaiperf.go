package input

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/explorerray/itpe-report/config"
)

// ProfileExport represents the structure of the GenAI-Perf profile-export.json
type ProfileExport struct {
	Experiments []Experiment `json:"experiments"`
	Version     string       `json:"version"`
	ServiceKind string       `json:"service_kind"`
	Endpoint    string       `json:"endpoint"`
}

// Experiment contains details about a single experiment
type Experiment struct {
	Experiment struct {
		Mode  string `json:"mode"`
		Value int    `json:"value"`
	} `json:"experiment"`
	Requests         []Request `json:"requests"`
	WindowBoundaries []int64   `json:"window_boundaries"`
}

// Request represents a single request within an experiment
type Request struct {
	Timestamp     int64 `json:"timestamp"`
	RequestInputs struct {
		Payload string `json:"payload"`
	} `json:"request_inputs"`
	ResponseTimestamps []int64 `json:"response_timestamps"`
	ResponseOutputs    []struct {
		Response string `json:"response"`
	} `json:"response_outputs"`
}

// GenAIPerf config for specific experiment
type GenAIPerfExpConf struct {
	Model       string
	PMSize      int // Parameter size (unit: billion parameters)
	InputMean   int
	OutputMean  int
	Concurrency int
	RunCount    int
}

// GenAIPerfMetrics holds computed metrics from the profile
type GenAIPerfMetrics struct {
	Model                 string
	Concurrency           int
	TotalTimeSec          float64
	NumRequests           int
	RequestThroughput     float64
	AvgTTFTMs             float64
	AvgRequestLatencyMs   float64
	AvgITLMs              float64
	TotalOutputTokens     int
	OutputTokenThroughput float64
}

// ParseGenAIPerfJSON reads and parses the profile-export.json file
func ParseGenAIPerfJSON(filename string, logger *slog.Logger) (*ProfileExport, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		logger.Error("Failed to read GenAI-Perf JSON file", "file", filename, "error", err)
		return nil, fmt.Errorf("reading file %s: %v", filename, err)
	}

	var profile ProfileExport
	if err := json.Unmarshal(data, &profile); err != nil {
		logger.Error("Failed to parse GenAI-Perf JSON file", "file", filename, "error", err)
		return nil, fmt.Errorf("parsing JSON from %s: %v", filename, err)
	}

	return &profile, nil
}

// ComputeMetrics computes performance metrics from an experiment
func ComputeMetrics(exp Experiment) GenAIPerfMetrics {
	reqs := exp.Requests
	lastReq := reqs[len(reqs)-1]
	expBegin := reqs[0].Timestamp
	expEnd := lastReq.ResponseTimestamps[len(lastReq.ResponseTimestamps)-1]
	metrics := GenAIPerfMetrics{
		Concurrency:  exp.Experiment.Value,
		TotalTimeSec: float64(expEnd-expBegin) / 1e9,
		NumRequests:  len(reqs),
	}

	// Extract model from first payload
	if len(reqs) > 0 {
		var payloadMap map[string]interface{}
		if err := json.Unmarshal([]byte(reqs[0].RequestInputs.Payload), &payloadMap); err == nil {
			if model, ok := payloadMap["model"].(string); ok {
				metrics.Model = model
			}
		}
	}

	var sumTTFT, sumRequestLatency, sumITL float64
	var totalOutputTokens, numITLIntervals int

	for _, req := range reqs {
		if len(req.ResponseTimestamps) == 0 {
			continue
		}

		reqBegin := req.Timestamp
		reqEnd := req.ResponseTimestamps[len(req.ResponseTimestamps)-1]

		// Time to First Token (TTFT) in milliseconds
		ttft := float64(req.ResponseTimestamps[0]-reqBegin) / 1e6
		sumTTFT += ttft

		// Request Latency in milliseconds
		requestLatency := float64(reqEnd-reqBegin) / 1e6
		sumRequestLatency += requestLatency

		// Output tokens: count response outputs minus the [DONE] chunk
		outputTokens := len(req.ResponseOutputs) - 1
		totalOutputTokens += outputTokens

		// Inter-Token Latency (ITL): average time between responses after the first
		if len(req.ResponseTimestamps) > 1 {
			var sumInterval float64
			for i := 1; i < len(req.ResponseTimestamps); i++ {
				interval := float64(req.ResponseTimestamps[i]-req.ResponseTimestamps[i-1]) / 1e6
				sumInterval += interval
			}
			itl := sumInterval / float64(len(req.ResponseTimestamps)-1)
			sumITL += itl
			numITLIntervals++
		}
	}

	if metrics.NumRequests > 0 {
		metrics.AvgTTFTMs = sumTTFT / float64(metrics.NumRequests)
		metrics.AvgRequestLatencyMs = sumRequestLatency / float64(metrics.NumRequests)
		metrics.RequestThroughput = float64(metrics.NumRequests) / metrics.TotalTimeSec
		metrics.TotalOutputTokens = totalOutputTokens
		metrics.OutputTokenThroughput = float64(totalOutputTokens) / metrics.TotalTimeSec
	}
	if numITLIntervals > 0 {
		metrics.AvgITLMs = sumITL / float64(numITLIntervals)
	}

	return metrics
}

func GenJSONPaths(config config.Config) ([]string, error) {
	// Use config.GenAIPerf and concate config.GenAIArtfPath
	// example: $(model)-$(inputMean)-$(outputMean)-concurrency$(concurrency)/$(RunCount)_$(concurrency)_profile.json

	paths := []string{}
	gp := config.GenAIPerf
	tc := gp.TokenConfs
	tci := tc.Input
	tco := tc.Output

	if config.GenAIArtfDir == "" {
		return nil, fmt.Errorf("artifacts directory is empty")
	}
	if len(gp.Models) == 0 {
		return nil, fmt.Errorf("no models specified in GenAIPerf")
	}
	if len(gp.Concurrency) == 0 {
		return nil, fmt.Errorf("no concurrency values specified in GenAIPerf")
	}
	if len(gp.Requests.RunCount) == 0 {
		return nil, fmt.Errorf("no run counts specified in GenAIPerf")
	}
	if len(tci) == 0 || len(tco) == 0 {
		return nil, fmt.Errorf("no input or output token configurations specified in GenAIPerf")
	}

	for _, model := range gp.Models {
		for _, i := range tci {
			for _, o := range tco {
				for _, concurrency := range gp.Concurrency {
					for _, runCount := range gp.Requests.RunCount {
						path := fmt.Sprintf("%s/%s-%d-%d-concurrency%d/%d_%d_profile.json",
							config.GenAIArtfDir, model, i.Mean, o.Mean, concurrency, runCount, concurrency)
						paths = append(paths, path)
					}
				}
			}
		}
	}
	return paths, nil
}

func GetConfFromPath(path string) (GenAIPerfExpConf, error) {
	var ec GenAIPerfExpConf
	// Example path: $(model)-$(inputMean)-$(outputMean)-concurrency$(concurrency)/$(RunCount)_$(concurrency)_profile.json
	// Split by '/' and then by '-'
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ec, fmt.Errorf("invalid path format: %s", path)
	}

	dirName := strings.Split(parts[len(parts)-2], "-")
	if len(dirName) < 4 {
		return ec, fmt.Errorf("invalid directory name in path: %s", parts[len(parts)-2])
	}

	modelParts := strings.Split(dirName[0], ":")
	ec.Model = modelParts[0]
	ec.PMSize, _ = strconv.Atoi(strings.TrimSuffix(modelParts[1], "b"))
	ec.InputMean, _ = strconv.Atoi(dirName[1])
	ec.OutputMean, _ = strconv.Atoi(dirName[2])
	ec.Concurrency, _ = strconv.Atoi(strings.TrimPrefix(dirName[3], "concurrency"))
	ec.RunCount, _ = strconv.Atoi(strings.Split(parts[len(parts)-1], "_")[0])

	return ec, nil
}
