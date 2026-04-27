package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/enricher"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	enrichSnapshotFile  string
	enrichMetaFile      string
	enrichEnv           string
	enrichOwner         string
	enrichDefaultEnv    string
)

func init() {
	enrichCmd := &cobra.Command{
		Use:   "enrich",
		Short: "Attach metadata (owner, environment, tier) to drift results",
		RunE:  runEnrich,
	}
	enrichCmd.Flags().StringVar(&enrichSnapshotFile, "snapshot", "", "snapshot file to enrich (required)")
	enrichCmd.Flags().StringVar(&enrichMetaFile, "meta", "", "JSON file mapping service names to Meta objects")
	enrichCmd.Flags().StringVar(&enrichEnv, "filter-env", "", "only output results for this environment")
	enrichCmd.Flags().StringVar(&enrichOwner, "filter-owner", "", "only output results for this owner")
	enrichCmd.Flags().StringVar(&enrichDefaultEnv, "default-env", "", "default environment label when not specified per service")
	_ = enrichCmd.MarkFlagRequired("snapshot")
	rootCmd.AddCommand(enrichCmd)
}

func runEnrich(cmd *cobra.Command, args []string) error {
	results, err := snapshot.Load(enrichSnapshotFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	opts := enricher.Options{DefaultEnvironment: enrichDefaultEnv}

	if enrichMetaFile != "" {
		f, err := os.Open(enrichMetaFile)
		if err != nil {
			return fmt.Errorf("opening meta file: %w", err)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&opts.ServiceMeta); err != nil {
			return fmt.Errorf("parsing meta file: %w", err)
		}
	}

	enriched := enricher.Apply(results, opts)

	if enrichEnv != "" {
		enriched = enricher.FilterByEnvironment(enriched, enrichEnv)
	}
	if enrichOwner != "" {
		enriched = enricher.FilterByOwner(enriched, enrichOwner)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(enriched); err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}
	return nil
}
