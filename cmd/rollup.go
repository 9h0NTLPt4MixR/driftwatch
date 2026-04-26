package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/driftwatch/internal/history"
	"github.com/driftwatch/internal/rollup"
	"github.com/spf13/cobra"
)

var (
	rollupGranularity string
	rollupHistoryFile string
	rollupFormat      string
)

var rollupCmd = &cobra.Command{
	Use:   "rollup",
	Short: "Aggregate drift history into time-bucketed summaries",
	RunE:  runRollup,
}

func init() {
	rollupCmd.Flags().StringVar(&rollupGranularity, "granularity", "daily", "Time bucket size: hourly, daily, weekly")
	rollupCmd.Flags().StringVar(&rollupHistoryFile, "history", ".driftwatch/history.jsonl", "Path to history file")
	rollupCmd.Flags().StringVar(&rollupFormat, "format", "table", "Output format: table, json")
	rootCmd.AddCommand(rollupCmd)
}

func runRollup(cmd *cobra.Command, args []string) error {
	entries, err := history.LoadAll(rollupHistoryFile)
	if err != nil {
		return fmt.Errorf("loading history: %w", err)
	}

	var gran rollup.Granularity
	switch rollupGranularity {
	case "hourly":
		gran = rollup.Hourly
	case "weekly":
		gran = rollup.Weekly
	default:
		gran = rollup.Daily
	}

	buckets := rollup.Compute(entries, gran)
	if len(buckets) == 0 {
		fmt.Println("No history data to roll up.")
		return nil
	}

	if rollupFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(buckets)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "WINDOW_START\tWINDOW_END\tTOTAL\tDRIFTED\tCLEAN\tDIFFS\tDRIFT_RATE")
	for _, b := range buckets {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%.2f%%\n",
			b.WindowStart.Format(time.DateOnly),
			b.WindowEnd.Format(time.DateOnly),
			b.TotalServices,
			b.DriftedCount,
			b.CleanCount,
			b.TotalDiffs,
			b.DriftRate*100,
		)
	}
	return w.Flush()
}
