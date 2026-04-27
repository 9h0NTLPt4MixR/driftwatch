package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/retention"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	retentionMaxAge      string
	retentionMaxEntries  int
	retentionKeepDrifted bool
	retentionSnapshotIn  string
	retentionSnapshotOut string
)

var retentionCmd = &cobra.Command{
	Use:   "retention",
	Short: "Apply a retention policy to a snapshot of drift results",
	RunE:  runRetention,
}

func init() {
	retentionCmd.Flags().StringVar(&retentionMaxAge, "max-age", "", "Maximum age of results to retain (e.g. 72h)")
	retentionCmd.Flags().IntVar(&retentionMaxEntries, "max-entries", 0, "Maximum number of results to retain (0 = unlimited)")
	retentionCmd.Flags().BoolVar(&retentionKeepDrifted, "keep-drifted", false, "Never remove drifted results due to age")
	retentionCmd.Flags().StringVar(&retentionSnapshotIn, "in", "snapshot.json", "Input snapshot file")
	retentionCmd.Flags().StringVar(&retentionSnapshotOut, "out", "", "Output snapshot file (defaults to overwriting --in)")
	rootCmd.AddCommand(retentionCmd)
}

func runRetention(cmd *cobra.Command, _ []string) error {
	results, err := snapshot.Load(retentionSnapshotIn)
	if err != nil {
		return fmt.Errorf("retention: loading snapshot: %w", err)
	}

	p := retention.Policy{
		MaxEntries:  retentionMaxEntries,
		KeepDrifted: retentionKeepDrifted,
	}
	if retentionMaxAge != "" {
		d, err := time.ParseDuration(retentionMaxAge)
		if err != nil {
			return fmt.Errorf("retention: invalid --max-age %q: %w", retentionMaxAge, err)
		}
		p.MaxAge = d
	}

	out, err := retention.Apply(results, p)
	if err != nil {
		return fmt.Errorf("retention: applying policy: %w", err)
	}

	dest := retentionSnapshotOut
	if dest == "" {
		dest = retentionSnapshotIn
	}
	if err := snapshot.Save(out.Retained, dest); err != nil {
		return fmt.Errorf("retention: saving snapshot: %w", err)
	}

	stats := retention.Stats(out)
	fmt.Fprintf(os.Stdout, "retention: retained=%d removed=%d -> %s\n",
		stats["retained"], stats["removed"], dest)
	return nil
}
