package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/baseline"
	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/fetcher"
	"github.com/spf13/cobra"
)

var (
	baselinePath   string
	baselineCompare bool
)

var baselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Save or compare a configuration baseline",
	RunE:  runBaseline,
}

func init() {
	baselineCmd.Flags().StringVarP(&baselinePath, "output", "o", "baseline.json", "path to baseline file")
	baselineCmd.Flags().StringVarP(&cfgFile, "config", "c", "driftwatch.yaml", "config file")
	baselineCmd.Flags().BoolVar(&baselineCompare, "compare", false, "compare current state against saved baseline")
	rootCmd.AddCommand(baselineCmd)
}

func runBaseline(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	f := fetcher.New(nil)
	results, err := drift.Scan(cfg.Services, f)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	if baselineCompare {
		base, err := baseline.Load(baselinePath)
		if err != nil {
			return fmt.Errorf("load baseline: %w", err)
		}
		compared := baseline.Compare(base, results)
		for _, r := range compared {
			if len(r.Diffs) > 0 {
				fmt.Fprintf(os.Stdout, "[DRIFT] %s: %d key(s) differ from baseline\n", r.ServiceName, len(r.Diffs))
			} else {
				fmt.Fprintf(os.Stdout, "[OK]    %s: matches baseline\n", r.ServiceName)
			}
		}
		return nil
	}

	if err := baseline.Save(baselinePath, results); err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}
	fmt.Fprintf(os.Stdout, "Baseline saved to %s (%d service(s))\n", baselinePath, len(results))
	return nil
}
