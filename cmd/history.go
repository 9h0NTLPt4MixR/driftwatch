package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/driftwatch/internal/history"
	"github.com/spf13/cobra"
)

var historyFile string

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show drift trend from recorded scan history",
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().StringVar(&historyFile, "file", "drift-history.json", "Path to history file")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	entries, err := history.LoadAll(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No history file found. Run 'scan --record' to start tracking.")
			return nil
		}
		return fmt.Errorf("failed to load history: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No history entries found.")
		return nil
	}
	points := history.DriftTrend(entries)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tDRIFTED SERVICES")
	for _, p := range points {
		fmt.Fprintf(w, "%s\t%d\n", p.Timestamp.Format("2006-01-02 15:04:05"), p.DriftedCount)
	}
	w.Flush()
	return nil
}
