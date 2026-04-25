package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/pruner"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	pruneSnapshotFile  string
	pruneMaxAge        string
	pruneRemoveClean   bool
	pruneExclude       []string
	pruneOutputFile    string
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove stale or clean results from a snapshot",
	RunE:  runPrune,
}

func init() {
	rootCmd.AddCommand(pruneCmd)
	pruneCmd.Flags().StringVar(&pruneSnapshotFile, "snapshot", "snapshot.json", "Path to snapshot file")
	pruneCmd.Flags().StringVar(&pruneMaxAge, "max-age", "", "Remove results older than this duration (e.g. 24h)")
	pruneCmd.Flags().BoolVar(&pruneRemoveClean, "remove-clean", false, "Drop results with no drift")
	pruneCmd.Flags().StringSliceVar(&pruneExclude, "exclude", nil, "Service names to exclude")
	pruneCmd.Flags().StringVar(&pruneOutputFile, "output", "", "Write pruned snapshot to this file (default: overwrite input)")
}

func runPrune(cmd *cobra.Command, args []string) error {
	results, err := snapshot.Load(pruneSnapshotFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	opts := pruner.Options{
		RemoveClean:     pruneRemoveClean,
		ExcludeServices: pruneExclude,
	}

	if pruneMaxAge != "" {
		d, err := time.ParseDuration(pruneMaxAge)
		if err != nil {
			return fmt.Errorf("invalid --max-age %q: %w", pruneMaxAge, err)
		}
		opts.MaxAge = d
	}

	kept, stats := pruner.PruneWithStats(results, opts)

	dest := pruneOutputFile
	if dest == "" {
		dest = pruneSnapshotFile
	}

	if err := snapshot.Save(kept, dest); err != nil {
		return fmt.Errorf("saving pruned snapshot: %w", err)
	}

	fmt.Fprintf(os.Stdout, "pruned: total=%d retained=%d removed=%d → %s\n",
		stats.Total, stats.Retained, stats.Removed, dest)
	return nil
}
