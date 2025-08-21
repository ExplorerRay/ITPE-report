package input

import (
	"log/slog"

	"github.com/explorerray/itpe-report/config"
)

type ExpMetrics struct {
	PerfM  GenAIPerfMetrics
	PowerM KeplerPowerMetrics
}

type ExpMetricPair map[GenAIPerfExpConf]ExpMetrics

func GenExpMetricPair(c config.Config, logger *slog.Logger) (ExpMetricPair, error) {
	// Mapping plot name (model_inputMean_outputMean) to a list of MetricPair
	// One plot would have #concurrency metric
	expMetricsPair := make(ExpMetricPair)

	paths, err := GenJSONPaths(c)
	if err != nil {
		return nil, err
	}

	// logging how many files need to parse
	logger.Info("Start parsing GenAI-Perf experiment results", "count", len(paths))
	for _, path := range paths {
		profile, err := ParseGenAIPerfJSON(path, logger)
		if err != nil {
			return nil, err
		}

		ec, err := GetConfFromPath(path)
		if err != nil {
			return nil, err
		}

		// Only one experiment in Custom GenAIPerf
		pfm := ComputeMetrics(profile.Experiments[0], ec, logger)
		pwm := GetPowerMetrics(profile.Experiments[0])

		expMetricsPair[ec] = ExpMetrics{
			PerfM:  pfm,
			PowerM: pwm,
		}
	}

	return expMetricsPair, nil
}
