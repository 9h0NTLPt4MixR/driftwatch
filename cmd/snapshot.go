package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var snapshotDir string

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Scan services and save results as a snapshot",
	RunE:  runSnapshot,
}

func init() {
	snapshotCmd.Flags().StringVar(&snapshotDir, "dir", "", "Directory to store snapshots (default: .driftwatch/snapshots)")
	snapshotCmd.Flags().StringVar(&cfgFile, "config", "driftwatch.yaml", "Path to config file")
	rootCmd.AddCommand(snapshotCmd)
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	f := fetcher.New(nil)
	var results []drift.Result

	for _, svc := range cfg.Services {
		actual, err := f.Fetch(svc.Endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: fetch %s: %v\n", svc.Name, err)
			continue
		}
		result := drift.Scan(svc, actual)
		results = append(results, result)
	}

	path, err := snapshot.Save(results, snapshotDir)
	if err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	fmt.Printf("Snapshot saved: %s\n", path)
	return nil
}
