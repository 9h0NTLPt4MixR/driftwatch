package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/digester"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	digestSnapshotFile string
	digestPrevFile     string
	digestJSON         bool
)

func init() {
	digestCmd := &cobra.Command{
		Use:   "digest",
		Short: "Compute fingerprints for a drift snapshot and detect changes",
		RunE:  runDigest,
	}
	digestCmd.Flags().StringVar(&digestSnapshotFile, "snapshot", "drift-snapshot.json", "snapshot file to fingerprint")
	digestCmd.Flags().StringVar(&digestPrevFile, "prev", "", "previous digest file to compare against")
	digestCmd.Flags().BoolVar(&digestJSON, "json", false, "output results as JSON")
	rootCmd.AddCommand(digestCmd)
}

func runDigest(cmd *cobra.Command, args []string) error {
	results, err := snapshot.Load(digestSnapshotFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	current, err := digester.Compute(results)
	if err != nil {
		return fmt.Errorf("computing digests: %w", err)
	}

	if digestPrevFile != "" {
		raw, err := os.ReadFile(digestPrevFile)
		if err != nil {
			return fmt.Errorf("reading previous digest file: %w", err)
		}
		var prev []digester.Result
		if err := json.Unmarshal(raw, &prev); err != nil {
			return fmt.Errorf("parsing previous digest file: %w", err)
		}
		changed := digester.Changed(prev, current)
		if len(changed) == 0 {
			fmt.Println("No changes detected since previous digest.")
		} else {
			fmt.Printf("Changed services (%d):\n", len(changed))
			for _, s := range changed {
				fmt.Printf("  - %s\n", s)
			}
		}
	}

	if digestJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(current)
	}

	for _, r := range current {
		status := "clean"
		if r.Drifted {
			status = "drifted"
		}
		fmt.Printf("%-30s %s  [%s]\n", r.Service, r.Digest, status)
	}
	return nil
}
