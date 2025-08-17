package stdout

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/explorerray/itpe-report/internal/client/promclient"
	"github.com/explorerray/itpe-report/internal/input"
	"github.com/jedib0t/go-pretty/v6/table"
)

// cleanQueryName removes kepler_ prefix and _joules_total or _total suffix
func cleanQueryName(name string) string {
	// Remove labels (e.g., {container_name="ollama"}) if present
	if idx := strings.Index(name, "{"); idx != -1 {
		name = name[:idx]
	}
	// Remove kepler_ prefix
	name = strings.TrimPrefix(name, "kepler_")
	// Remove _joules_total* suffix
	name = strings.Split(name, "_joules_total")[0]
	return name
}

func QueryToTableOut(qs []promclient.QueryInfo, qrs []promclient.QueryResponse) {
	// Create a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Query", "Value", "Timestamp"})

	// Process responses
	for i, resp := range qrs {
		queryName := cleanQueryName(qs[i].Name)
		if resp.Error != nil {
			fmt.Printf("Error: %v\n", resp.Error)
			t.AppendRow(table.Row{queryName, "N/A", "N/A"})
			continue
		}
		if len(resp.Warnings) > 0 {
			fmt.Printf("Warnings for %s: %v\n", queryName, resp.Warnings)
		}
		if len(resp.Results) == 0 {
			t.AppendRow(table.Row{queryName, "N/A", "N/A"})
			continue
		}
		for _, result := range resp.Results {
			valueStr := fmt.Sprintf("%v", result.Value)
			timeStr := result.Timestamp.Time().Format(time.RFC3339)
			t.AppendRow(table.Row{queryName, valueStr, timeStr})
		}
	}

	// Render the table
	fmt.Println("Prometheus Query Results:")
	t.Render()
	fmt.Println()
}

func PerfMetricsToTableOut(metrics input.GenAIPerfMetrics) {
	// Create a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", "Value"})

	// Append metrics
	t.AppendRow(table.Row{"Model", metrics.Model})
	t.AppendRow(table.Row{"Concurrency", metrics.Concurrency})
	t.AppendRow(table.Row{"Total Time (s)", fmt.Sprintf("%.2f", metrics.TotalTimeSec)})
	t.AppendRow(table.Row{"Number of Requests", metrics.NumRequests})
	t.AppendRow(table.Row{"Request Throughput (req/s)", fmt.Sprintf("%.2f", metrics.RequestThroughput)})
	t.AppendRow(table.Row{"Avg TTFT (ms)", fmt.Sprintf("%.2f", metrics.AvgTTFTMs)})
	t.AppendRow(table.Row{"Avg Inter Token Latency (ms)", fmt.Sprintf("%.2f", metrics.AvgITLMs)})
	t.AppendRow(table.Row{"Avg Request Latency (ms)", fmt.Sprintf("%.2f", metrics.AvgRequestLatencyMs)})
	t.AppendRow(table.Row{"Total Output Tokens", metrics.TotalOutputTokens})
	t.AppendRow(table.Row{"Output Token Throughput (tokens/s)", fmt.Sprintf("%.2f", metrics.OutputTokenThroughput)})

	// Render the table
	fmt.Println("GenAI Perf Metrics:")
	t.Render()
	fmt.Println()
}

func PowerMetricsToTableOut(metrics input.KeplerPowerMetrics) {
	// Create a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", "Value"})

	// Append power metrics
	t.AppendRow(table.Row{"Node Platform (J)", fmt.Sprintf("%.2f", metrics.NodePlatformJ)})
	t.AppendRow(table.Row{"Node GPU (J)", fmt.Sprintf("%.2f", metrics.NodeGPUJ)})
	t.AppendRow(table.Row{"Node Package (J)", fmt.Sprintf("%.2f", metrics.NodePackageJ)})
	t.AppendRow(table.Row{"Node DRAM (J)", fmt.Sprintf("%.2f", metrics.NodeDRAMJ)})
	t.AppendRow(table.Row{"Node Other (J)", fmt.Sprintf("%.2f", metrics.NodeOtherJ)})
	t.AppendRow(table.Row{"Pod Platform (J)", fmt.Sprintf("%.2f", metrics.PodPlatformJ)})
	t.AppendRow(table.Row{"Pod GPU (J)", fmt.Sprintf("%.2f", metrics.PodGPUJ)})
	t.AppendRow(table.Row{"Pod Package (J)", fmt.Sprintf("%.2f", metrics.PodPackageJ)})
	t.AppendRow(table.Row{"Pod DRAM (J)", fmt.Sprintf("%.2f", metrics.PodDRAMJ)})
	t.AppendRow(table.Row{"Pod Other (J)", fmt.Sprintf("%.2f", metrics.PodOtherJ)})

	// Render the table
	fmt.Println("Power Metrics:")
	t.Render()
	fmt.Println()
}
