package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/ranker"
	"github.com/spf13/cobra"
)

var (
	rankProtectedKeys   []string
	rankProtectedWeight int
)

var rankCmd = &cobra.Command{
	Use:   "rank",
	Short: "Rank services by drift severity",
	Long:  "Scan all services and rank results from most to least severe drift.",
	RunE:  runRank,
}

func init() {
	rankCmd.Flags().StringSliceVar(&rankProtectedKeys, "protected", nil, "Comma-separated list of protected keys (weighted higher)")
	rankCmd.Flags().IntVar(&rankProtectedWeight, "weight", 3, "Score multiplier for protected key drifts")
	rootCmd.AddCommand(rankCmd)
}

func runRank(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	f := fetcher.New(nil)
	var results []drift.Result
	for _, svc := range cfg.Services {
		actual, err := f.Fetch(svc.Endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: fetch %s: %v\n", svc.Name, err)
			continue
		}
		res := drift.Scan(svc, actual)
		results = append(results, res)
	}

	opts := ranker.Options{
		ProtectedKeys:   rankProtectedKeys,
		ProtectedWeight: rankProtectedWeight,
	}
	ranked := ranker.Rank(results, opts)

	fmt.Printf("%-30s %-10s %s\n", "SERVICE", "SEVERITY", "SCORE")
	fmt.Println(strings.Repeat("-", 50))
	for _, r := range ranked {
		fmt.Printf("%-30s %-10s %d\n", r.Service, r.Severity, r.Score)
	}
	return nil
}
