package plot

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/explorerray/itpe-report/config"
	"github.com/explorerray/itpe-report/internal/input"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func CreatePlotsSubdir(conf config.Config) string {
	plotDir := filepath.Join(conf.GenAIArtfDir, "plots")
	if err := os.MkdirAll(filepath.Join(plotDir, "by_model"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(plotDir, "by_length"), os.ModePerm); err != nil {
		panic(err)
	}
	return plotDir
}

func genPlotterXY(xValues []float64, yValues []float64) plotter.XYs {
	pts := make(plotter.XYs, 0, len(xValues))
	for i := range xValues {
		if yValues[i] != 0 && !math.IsNaN(yValues[i]) {
			pts = append(pts, plotter.XY{X: xValues[i], Y: yValues[i]})
		}
	}
	return pts
}

// lengthKey represents input/output length combination
type lengthKey struct {
	inputMean  int
	outputMean int
}

// modelGroup groups data by model and parameter size
type modelGroup struct {
	model  string
	pmSize int
}

// MetricByLengthData groups metric data by lengthKey then modelGroup
type MetricByLengthData map[string]map[lengthKey]map[modelGroup][]float64

// MetricByModelData groups metric data by modelGroup then lengthKey
type MetricByModelData map[string]map[modelGroup]map[lengthKey][]float64

func getDistinctModels(emp input.ExpMetricPair) []string {
	modelSet := make(map[string]bool)
	for ec := range emp {
		modelSet[ec.Model] = true
	}
	models := make([]string, 0, len(modelSet))
	for model := range modelSet {
		models = append(models, model)
	}
	sort.Strings(models) // Sort for consistent order
	return models
}

// collectMetricData gathers metric data for all models and input/output lengths
func collectMetricData(emp input.ExpMetricPair) (MetricByLengthData, MetricByModelData, []float64, []int, []int, error) {
	// Step 1: Collect data by lengthKey, modelGroup, and concurrency
	type metricValues struct {
		concurrency int
		values      config.YPlot
	}
	dataByLengthAndModel := make(map[lengthKey]map[modelGroup][]metricValues)
	inputMeans := make(map[int]bool)
	outputMeans := make(map[int]bool)
	for ec, mp := range emp {
		lk := lengthKey{inputMean: ec.InputMean, outputMean: ec.OutputMean}
		mg := modelGroup{model: ec.Model, pmSize: ec.PMSize}
		if _, exists := dataByLengthAndModel[lk]; !exists {
			dataByLengthAndModel[lk] = make(map[modelGroup][]metricValues)
		}
		dataByLengthAndModel[lk][mg] = append(dataByLengthAndModel[lk][mg], metricValues{
			concurrency: ec.Concurrency,
			values: config.YPlot{
				RequestThroughput:     mp.PerfM.RequestThroughput,
				OutputTokenThroughput: mp.PerfM.OutputTokenThroughput,
				AvgRequestLatencyMs:   mp.PerfM.AvgRequestLatencyMs,
				AvgTTFTMs:             mp.PerfM.AvgTTFTMs,
				AvgITLMs:              mp.PerfM.AvgITLMs,
				NodePlatformJ:         mp.PowerM.NodePlatformJ,
				NodeGPUJ:              mp.PowerM.NodeGPUJ,
				NodeCPUJ:              mp.PowerM.NodePackageJ,
				EnergyPerToken:        mp.PowerM.NodePlatformJ / float64(mp.PerfM.TotalOutputTokens),
			},
		})
		inputMeans[ec.InputMean] = true
		outputMeans[ec.OutputMean] = true
	}

	// Step 2: Collect all unique Concurrency values
	var allConcurrencies []int
	concurrencySet := make(map[int]bool)
	for _, modelData := range dataByLengthAndModel {
		for _, metrics := range modelData {
			for _, mv := range metrics {
				if !concurrencySet[mv.concurrency] {
					concurrencySet[mv.concurrency] = true
					allConcurrencies = append(allConcurrencies, mv.concurrency)
				}
			}
		}
	}
	if len(allConcurrencies) == 0 {
		return nil, nil, nil, nil, nil, fmt.Errorf("no data found in emp")
	}
	sort.Ints(allConcurrencies)
	xValues := make([]float64, len(allConcurrencies))
	for i, c := range allConcurrencies {
		xValues[i] = float64(c)
	}

	// Step 3: Collect unique InputMean and OutputMean values
	var uniqueInputMeans, uniqueOutputMeans []int
	for im := range inputMeans {
		uniqueInputMeans = append(uniqueInputMeans, im)
	}
	for om := range outputMeans {
		uniqueOutputMeans = append(uniqueOutputMeans, om)
	}
	sort.Ints(uniqueInputMeans)
	sort.Ints(uniqueOutputMeans)

	// Step 4: Organize data by metric, lengthKey, and modelGroup (for createMetricPlotByModel)
	metricsByLength := make(MetricByLengthData)
	// Also organize by metric, modelGroup, and lengthKey (for createMetricPlotByLength)
	metricsByModel := make(MetricByModelData)
	metricsConfig := config.GetMetricsConfig()
	for metricName := range metricsConfig {
		metricsByLength[metricName] = make(map[lengthKey]map[modelGroup][]float64)
		metricsByModel[metricName] = make(map[modelGroup]map[lengthKey][]float64)
	}
	for lk, modelData := range dataByLengthAndModel {
		for mg, metrics := range modelData {
			for metricName := range metricsConfig {
				if _, exists := metricsByLength[metricName][lk]; !exists {
					metricsByLength[metricName][lk] = make(map[modelGroup][]float64)
				}
				if _, exists := metricsByModel[metricName][mg]; !exists {
					metricsByModel[metricName][mg] = make(map[lengthKey][]float64)
				}
				yValues := make([]float64, len(allConcurrencies))
				for i := range yValues {
					yValues[i] = 0
				}
				for _, mv := range metrics {
					idx := -1
					for i, c := range allConcurrencies {
						if c == mv.concurrency {
							idx = i
							break
						}
					}
					if idx == -1 {
						continue
					}
					switch metricName {
					case "Request Throughput":
						yValues[idx] = mv.values.RequestThroughput
					case "Output Token Throughput":
						yValues[idx] = mv.values.OutputTokenThroughput
					case "Avg Request Latency":
						yValues[idx] = mv.values.AvgRequestLatencyMs
					case "Avg TTFT":
						yValues[idx] = mv.values.AvgTTFTMs
					case "Avg ITL":
						yValues[idx] = mv.values.AvgITLMs
					case "Node Platform":
						yValues[idx] = mv.values.NodePlatformJ
					case "Node GPU":
						yValues[idx] = mv.values.NodeGPUJ
					case "Node CPU":
						yValues[idx] = mv.values.NodeCPUJ
					case "Energy Per Token":
						yValues[idx] = mv.values.EnergyPerToken
					}
				}
				metricsByLength[metricName][lk][mg] = yValues
				metricsByModel[metricName][mg][lk] = yValues
			}
		}
	}
	return metricsByLength, metricsByModel, xValues, uniqueInputMeans, uniqueOutputMeans, nil
}

// createMetricPlotByModel generates a plot for a metric, grouped by PMSize and either InputMean or OutputMean
func createMetricPlotByModel(metricName string, pmSize int, groupBy string, groupValue int, dataByLength map[lengthKey]map[modelGroup][]float64, xValues []float64, plotDir string) error {
	config, exists := config.GetMetricsConfig()[metricName]
	if !exists {
		return fmt.Errorf("unknown metric: %s", metricName)
	}

	var title, filename string
	if groupBy == "input" {
		title = fmt.Sprintf("%s (%db Parameters, Input %d)", metricName, pmSize, groupValue)
		filename = fmt.Sprintf("by_model/%s_%db_input%d.png", config.Filename, pmSize, groupValue)
	} else {
		title = fmt.Sprintf("%s (%db Parameters, Output %d)", metricName, pmSize, groupValue)
		filename = fmt.Sprintf("by_model/%s_%db_output%d.png", config.Filename, pmSize, groupValue)
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Concurrency"
	p.Y.Label.Text = config.YLabel

	args := []interface{}{}
	for lk, modelData := range dataByLength {
		if (groupBy == "input" && lk.inputMean != groupValue) || (groupBy == "output" && lk.outputMean != groupValue) {
			continue
		}
		for mg, yValues := range modelData {
			if mg.pmSize != pmSize {
				continue
			}
			var label string
			if groupBy == "input" {
				label = fmt.Sprintf("%s-output%d", mg.model, lk.outputMean)
			} else {
				label = fmt.Sprintf("%s-input%d", mg.model, lk.inputMean)
			}
			args = append(args, label, genPlotterXY(xValues, yValues))
		}
	}
	if len(args) == 0 {
		return fmt.Errorf("no data for metric %s, PMSize %db, %s %d", metricName, pmSize, groupBy, groupValue)
	}
	if err := plotutil.AddLinePoints(p, args...); err != nil {
		return fmt.Errorf("failed to add line points for %s: %v", metricName, err)
	}

	filepath := filepath.Join(plotDir, filename)
	if err := p.Save(4*vg.Inch, 4*vg.Inch, filepath); err != nil {
		return fmt.Errorf("failed to save plot %s: %v", filepath, err)
	}
	return nil
}

// createMetricPlotByLength generates a plot for a metric and model, with lines for input/output lengths
func createMetricPlotByLength(metricName string, mg modelGroup, dataByLength map[lengthKey][]float64, xValues []float64, plotDir string) error {
	config, exists := config.GetMetricsConfig()[metricName]
	if !exists {
		return fmt.Errorf("unknown metric: %s", metricName)
	}

	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s (%s, %db Parameters)", metricName, mg.model, mg.pmSize)
	p.X.Label.Text = "Concurrency"
	p.Y.Label.Text = config.YLabel

	args := []interface{}{}
	for lk, yValues := range dataByLength {
		lengthLabel := fmt.Sprintf("%d/%d", lk.inputMean, lk.outputMean)
		args = append(args, lengthLabel, genPlotterXY(xValues, yValues))
	}
	if len(args) == 0 {
		return fmt.Errorf("no data for metric %s, model %s", metricName, mg.model)
	}
	if err := plotutil.AddLinePoints(p, args...); err != nil {
		return fmt.Errorf("failed to add line points for %s: %v", metricName, err)
	}

	filename := fmt.Sprintf("by_length/%s_%s_%db.png", config.Filename, mg.model, mg.pmSize)
	filepath := filepath.Join(plotDir, filename)
	if err := p.Save(4*vg.Inch, 4*vg.Inch, filepath); err != nil {
		return fmt.Errorf("failed to save plot %s: %v", filepath, err)
	}
	return nil
}

// GeneratePlots coordinates plot generation
func GeneratePlots(emp input.ExpMetricPair, plotDir string) error {
	metricsByLength, metricsByModel, xValues, inputMeans, outputMeans, err := collectMetricData(emp)
	if err != nil {
		return fmt.Errorf("failed to collect metric data: %v", err)
	}

	// Original functionality: Plots by PMSize and either InputMean or OutputMean
	for metricName, dataByLength := range metricsByLength {
		// Group by pmSize
		byPMSize := make(map[int]map[lengthKey]map[modelGroup][]float64)
		for lk, modelData := range dataByLength {
			for mg := range modelData {
				if _, exists := byPMSize[mg.pmSize]; !exists {
					byPMSize[mg.pmSize] = make(map[lengthKey]map[modelGroup][]float64)
				}
				if _, exists := byPMSize[mg.pmSize][lk]; !exists {
					byPMSize[mg.pmSize][lk] = make(map[modelGroup][]float64)
				}
				byPMSize[mg.pmSize][lk][mg] = modelData[mg]
			}
		}
		for pmSize := range byPMSize {
			// Plots grouped by InputMean
			for _, inputMean := range inputMeans {
				if err := createMetricPlotByModel(metricName, pmSize, "input", inputMean, dataByLength, xValues, plotDir); err != nil {
					return err
				}
			}
			// Plots grouped by OutputMean
			for _, outputMean := range outputMeans {
				if err := createMetricPlotByModel(metricName, pmSize, "output", outputMean, dataByLength, xValues, plotDir); err != nil {
					return err
				}
			}
		}
	}

	// New functionality: Plots by model with lines for input/output lengths
	for _, modelName := range getDistinctModels(emp) {
		for metricName, dataByModel := range metricsByModel {
			for mg, dataByLength := range dataByModel {
				if mg.model != modelName {
					continue
				}
				if err := createMetricPlotByLength(metricName, mg, dataByLength, xValues, plotDir); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
