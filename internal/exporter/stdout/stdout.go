package stdout

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/explorerray/itpe-report/internal/promclient"
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
