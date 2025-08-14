package input

import (
	"encoding/json"
	"fmt"
	"os"
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
func ParseGenAIPerfJSON(filename string) (*ProfileExport, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %v", filename, err)
	}

	var profile ProfileExport
	if err := json.Unmarshal(data, &profile); err != nil {
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
