package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/fetcher"
	"github.com/example/driftwatch/internal/scorecard"
	"github.com/spf13/cobra"
)

var scorecardJSON bool

var scorecardCmd = &cobra.Command{
	Use:   "scorecard",
	Short: "Display a drift health scorecard for all services",
	RunE:  runScorecard,
}

func init() {
	scorecardCmd.Flags().BoolVar(&scorecardJSON, "json", false, "Output scorecard as JSON")
	rootCmd.AddCommand(scorecardCmd)
}

func runScorecard(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	f := fetcher.New(nil)
	results := make([]drift.Result, 0, len(cfg.Services))
	for _, svc := range cfg.Services {
		res, err := drift.Scan(svc, f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: skipping %s: %v\n", svc.Name, err)
			continue
		}
		results = append(results, res)
	}

	if len(results) == 0 {
		return fmt.Errorf("no services could be scanned; check configuration and connectivity")
	}

	score := scorecard.Compute(results)

	if scorecardJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(score)
	}

	fmt.Println(score.String())
	return nil
}
