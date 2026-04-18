package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/snapshot"
	"github.com/driftwatch/internal/summary"
	"github.com/spf13/cobra"
)

var summaryFile string
var summaryJSON bool

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show aggregated drift statistics from a snapshot",
	RunE:  runSummary,
}

func init() {
	summaryCmd.Flags().StringVarP(&summaryFile, "snapshot", "s", "drift-snapshot.json", "Snapshot file to summarise")
	summaryCmd.Flags().BoolVar(&summaryJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
	results, err := snapshot.Load(summaryFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	stats := summary.Compute(results)

	if summaryJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(stats)
	}

	fmt.Printf("Services scanned : %d\n", stats.TotalServices)
	fmt.Printf("Drifted          : %d\n", stats.DriftedCount)
	fmt.Printf("Clean            : %d\n", stats.CleanCount)
	fmt.Printf("Total mismatches : %d\n", stats.TotalMismatches)
	fmt.Printf("Drift rate       : %.1f%%\n", stats.DriftRate())
	if len(stats.ByService) > 0 {
		fmt.Println("\nPer-service mismatches:")
		for svc, count := range stats.ByService {
			fmt.Printf("  %-30s %d\n", svc, count)
		}
	}
	return nil
}
