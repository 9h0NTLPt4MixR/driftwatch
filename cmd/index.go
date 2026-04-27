package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/indexer"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	indexSnapshotFile string
	indexLookupService string
	indexLookupKey     string
)

func init() {
	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "Build and query an in-memory index of drift results",
		RunE:  runIndex,
	}
	indexCmd.Flags().StringVar(&indexSnapshotFile, "snapshot", "", "snapshot file to index (required)")
	indexCmd.Flags().StringVar(&indexLookupService, "service", "", "look up a specific service")
	indexCmd.Flags().StringVar(&indexLookupKey, "key", "", "look up services drifted on a specific key")
	_ = indexCmd.MarkFlagRequired("snapshot")
	rootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, _ []string) error {
	results, err := snapshot.Load(indexSnapshotFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	idx, err := indexer.Build(results)
	if err != nil {
		return fmt.Errorf("building index: %w", err)
	}

	if indexLookupService != "" {
		r, ok := idx.LookupService(indexLookupService)
		if !ok {
			fmt.Fprintf(os.Stderr, "service %q not found in snapshot\n", indexLookupService)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	}

	if indexLookupKey != "" {
		hits := idx.LookupKey(indexLookupKey)
		for _, r := range hits {
			fmt.Println(r.Service)
		}
		return nil
	}

	fmt.Println(idx.Stats())
	return nil
}
