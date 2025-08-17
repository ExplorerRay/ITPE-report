package input

import (
	"github.com/explorerray/itpe-report/config"
)

type ExpMetrics struct {
	PerfM  GenAIPerfMetrics
	PowerM KeplerPowerMetrics
}

type ExpMetricPair map[GenAIPerfExpConf]ExpMetrics

func GenExpMetricPair(c config.Config) ExpMetricPair {
	// Mapping plot name (model_inputMean_outputMean) to a list of MetricPair
	// One plot would have #concurrency metric
	expMetricsPair := make(ExpMetricPair)

	paths, err := GenJSONPaths(c)
	if err != nil {
		panic(err)
	}

	for _, path := range paths {
		profile, err := ParseGenAIPerfJSON(path)
		if err != nil {
			panic(err)
		}

		ec, err := GetConfFromPath(path)
		if err != nil {
			panic(err)
		}

		// Only one experiment in Custom GenAIPerf
		pfm := ComputeMetrics(profile.Experiments[0])
		pwm := GetPowerMetrics(profile.Experiments[0])

		expMetricsPair[ec] = ExpMetrics{
			PerfM:  pfm,
			PowerM: pwm,
		}
	}

	return expMetricsPair
}
