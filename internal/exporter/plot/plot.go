package plot

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/explorerray/itpe-report/config"
	"github.com/explorerray/itpe-report/internal/input"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// CreatePlotsSubdir creates the necessary subdirectories for saving plot images.
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

// genPlotterXY converts slices of float64 into a plotter.XYs structure, filtering out zero/NaN values.
func genPlotterXY(xValues []float64, yValues []float64) plotter.XYs {
	pts := make(plotter.XYs, 0, len(xValues))
	for i := range xValues {
		if yValues[i] != 0 && !math.IsNaN(yValues[i]) {
			pts = append(pts, plotter.XY{X: xValues[i], Y: yValues[i]})
		}
	}
	return pts
}

// lengthKey represents a unique input/output length combination.
type lengthKey struct {
	inputMean  int
	outputMean int
}

// modelGroup represents a unique model and parameter size combination.
type modelGroup struct {
	model  string
	pmSize int
}

// MetricByLengthData groups metric data first by metric name, then by lengthKey, then by modelGroup.
type MetricByLengthData map[string]map[lengthKey]map[modelGroup][]float64

// MetricByModelData groups metric data first by metric name, then by modelGroup, then by lengthKey.
type MetricByModelData map[string]map[modelGroup]map[lengthKey][]float64

// collectMetricData processes the raw experiment data and organizes it into structures suitable for plotting.
func collectMetricData(emp input.ExpMetricPair) (MetricByLengthData, MetricByModelData, []float64, []int, []int, error) {
	// This function remains largely the same, but is a critical part of the data pipeline.
	// It extracts, sorts, and organizes all data points before they are plotted.
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

	var uniqueInputMeans, uniqueOutputMeans []int
	for im := range inputMeans {
		uniqueInputMeans = append(uniqueInputMeans, im)
	}
	for om := range outputMeans {
		uniqueOutputMeans = append(uniqueOutputMeans, om)
	}
	sort.Ints(uniqueInputMeans)
	sort.Ints(uniqueOutputMeans)

	metricsByLength := make(MetricByLengthData)
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
				concurrencyIndexMap := make(map[int]int)
				for i, c := range allConcurrencies {
					concurrencyIndexMap[c] = i
				}
				for _, mv := range metrics {
					idx, ok := concurrencyIndexMap[mv.concurrency]
					if !ok {
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

// createMetricPlotByModel generates a plot, using the styleManager for consistent line styles.
func createMetricPlotByModel(metricName string, pmSize int, groupBy string, groupValue int, dataByLength map[lengthKey]map[modelGroup][]float64, xValues []float64, plotDir string, styleMgr *styleManager, logger *slog.Logger) error {
	metricConfig, exists := config.GetMetricsConfig()[metricName]
	if !exists {
		return fmt.Errorf("unknown metric: %s", metricName)
	}

	var title, filename string
	if groupBy == "input" {
		title = fmt.Sprintf("%s (%db Parameters, Input %d)", metricName, pmSize, groupValue)
		filename = fmt.Sprintf("by_model/%s_%db_input%d.png", metricConfig.Filename, pmSize, groupValue)
	} else {
		title = fmt.Sprintf("%s (%db Parameters, Output %d)", metricName, pmSize, groupValue)
		filename = fmt.Sprintf("by_model/%s_%db_output%d.png", metricConfig.Filename, pmSize, groupValue)
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Concurrency"
	p.Y.Label.Text = metricConfig.YLabel
	p.Legend.Top = true
	p.Legend.XOffs = -vg.Points(10)

	hasData := false
	for lk, modelData := range dataByLength {
		if (groupBy == "input" && lk.inputMean != groupValue) || (groupBy == "output" && lk.outputMean != groupValue) {
			continue
		}
		for mg, yValues := range modelData {
			if mg.pmSize != pmSize {
				continue
			}
			pts := genPlotterXY(xValues, yValues)
			if len(pts) == 0 {
				continue
			}
			hasData = true

			var label string
			if groupBy == "input" {
				label = fmt.Sprintf("%s-output%d", mg.model, lk.outputMean)
			} else {
				label = fmt.Sprintf("%s-input%d", mg.model, lk.inputMean)
			}

			lineStyle, glyphStyle := styleMgr.getStyle(label) // Use label as the unique key

			line, err := plotter.NewLine(pts)
			if err != nil {
				return err
			}
			line.LineStyle = lineStyle

			scatter, err := plotter.NewScatter(pts)
			if err != nil {
				return err
			}
			scatter.GlyphStyle = glyphStyle

			p.Add(line, scatter)
			p.Legend.Add(label, line, scatter)
		}
	}

	if !hasData {
		logger.Info("Skipping plot due to no data", "title", title)
		return nil // Not a fatal error, just no data to plot.
	}

	filepath := filepath.Join(plotDir, filename)
	if err := p.Save(5*vg.Inch, 5*vg.Inch, filepath); err != nil {
		return fmt.Errorf("failed to save plot %s: %v", filepath, err)
	}
	return nil
}

// createMetricPlotByLength generates a plot, using the styleManager for consistent line styles.
func createMetricPlotByLength(metricName string, mg modelGroup, dataByLength map[lengthKey][]float64, xValues []float64, plotDir string, styleMgr *styleManager, logger *slog.Logger) error {
	metricConfig, exists := config.GetMetricsConfig()[metricName]
	if !exists {
		return fmt.Errorf("unknown metric: %s", metricName)
	}

	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s (%s, %db Parameters)", metricName, mg.model, mg.pmSize)
	p.X.Label.Text = "Concurrency"
	p.Y.Label.Text = metricConfig.YLabel
	p.Legend.Top = true
	p.Legend.XOffs = -vg.Points(10)

	var sortedKeys []lengthKey
	for lk := range dataByLength {
		sortedKeys = append(sortedKeys, lk)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		if sortedKeys[i].inputMean != sortedKeys[j].inputMean {
			return sortedKeys[i].inputMean < sortedKeys[j].inputMean
		}
		return sortedKeys[i].outputMean < sortedKeys[j].outputMean
	})

	hasData := false
	for _, lk := range sortedKeys {
		yValues := dataByLength[lk]
		pts := genPlotterXY(xValues, yValues)
		if len(pts) == 0 {
			continue
		}
		hasData = true

		label := fmt.Sprintf("in%d/out%d", lk.inputMean, lk.outputMean)
		lineStyle, glyphStyle := styleMgr.getStyle(label) // Use label as the unique key

		line, err := plotter.NewLine(pts)
		if err != nil {
			return err
		}
		line.LineStyle = lineStyle

		scatter, err := plotter.NewScatter(pts)
		if err != nil {
			return err
		}
		scatter.GlyphStyle = glyphStyle

		p.Add(line, scatter)
		p.Legend.Add(label, line, scatter)
	}

	if !hasData {
		logger.Info("Skipping plot due to no data", "title", p.Title.Text)
		return nil
	}

	filename := fmt.Sprintf("by_length/%s_%s_%db.png", metricConfig.Filename, mg.model, mg.pmSize)
	filepath := filepath.Join(plotDir, filename)
	if err := p.Save(5*vg.Inch, 5*vg.Inch, filepath); err != nil {
		return fmt.Errorf("failed to save plot %s: %v", filepath, err)
	}
	return nil
}

// GeneratePlots coordinates the entire plot generation process.
func GeneratePlots(emp input.ExpMetricPair, plotDir string, logger *slog.Logger) error {
	metricsByLength, metricsByModel, xValues, inputMeans, outputMeans, err := collectMetricData(emp)
	if err != nil {
		return fmt.Errorf("failed to collect metric data: %v", err)
	}

	styleMgrForModelPlots := newStyleManager()
	styleMgrForLengthPlots := newStyleManager()

	// Generate plots grouped by model parameters
	for metricName, dataByLength := range metricsByLength {
		byPMSize := make(map[int]bool)
		for _, modelData := range dataByLength {
			for mg := range modelData {
				byPMSize[mg.pmSize] = true
			}
		}

		for pmSize := range byPMSize {
			for _, inputMean := range inputMeans {
				if err := createMetricPlotByModel(metricName, pmSize, "input", inputMean, dataByLength, xValues, plotDir, styleMgrForModelPlots, logger); err != nil {
					logger.Error("Failed to create plot by model (input)", "error", err, "metricName", metricName, "pmSize", pmSize, "inputMean", inputMean)
				}
			}
			for _, outputMean := range outputMeans {
				if err := createMetricPlotByModel(metricName, pmSize, "output", outputMean, dataByLength, xValues, plotDir, styleMgrForModelPlots, logger); err != nil {
					logger.Error("Failed to create plot by model (output)", "error", err, "metricName", metricName, "pmSize", pmSize, "outputMean", outputMean)
				}
			}
		}
	}

	// Generate plots grouped by input/output length
	for metricName, dataByModel := range metricsByModel {
		for mg, dataByLength := range dataByModel {
			if err := createMetricPlotByLength(metricName, mg, dataByLength, xValues, plotDir, styleMgrForLengthPlots, logger); err != nil {
				logger.Error("Failed to create plot by length", "error", err, "metricName", metricName, "modelGroup", mg)
			}
		}
	}

	logger.Info("Successfully generated all plots.")
	return nil
}
