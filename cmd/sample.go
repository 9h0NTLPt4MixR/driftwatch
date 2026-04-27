package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/sampler"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	sampleRate     float64
	sampleStrategy string
	sampleSeed     int64
	sampleInput    string
	sampleHash     bool
)

var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "Sample a subset of drift results from a snapshot",
	RunE:  runSample,
}

func init() {
	rootCmd.AddCommand(sampleCmd)
	sampleCmd.Flags().Float64VarP(&sampleRate, "rate", "r", 0.5, "Fraction of results to keep (0.0–1.0)")
	sampleCmd.Flags().StringVarP(&sampleStrategy, "strategy", "s", "random", "Sampling strategy: random, modulo, head")
	sampleCmd.Flags().Int64Var(&sampleSeed, "seed", 0, "Random seed for reproducible sampling")
	sampleCmd.Flags().StringVarP(&sampleInput, "input", "i", "", "Path to snapshot file (required)")
	sampleCmd.Flags().BoolVar(&sampleHash, "hash", false, "Use hash-based deterministic sampling by service name")
	_ = sampleCmd.MarkFlagRequired("input")
}

func runSample(cmd *cobra.Command, args []string) error {
	results, err := snapshot.Load(sampleInput)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	var sampled []drift.Result

	if sampleHash {
		sampled, err = sampler.HashSample(results, sampleRate)
	} else {
		sampled, err = sampler.Sample(results, sampler.Options{
			Rate:     sampleRate,
			Strategy: sampler.Strategy(sampleStrategy),
			Seed:     sampleSeed,
		})
	}
	if err != nil {
		return fmt.Errorf("sampling: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Sampled %d / %d results (rate=%.2f strategy=%s)\n",
		len(sampled), len(results), sampleRate, sampleStrategy)

	for _, r := range sampled {
		status := "clean"
		if r.Drifted {
			status = "drifted"
		}
		fmt.Fprintf(os.Stdout, "  %-30s %s\n", r.Service, status)
	}
	return nil
}
