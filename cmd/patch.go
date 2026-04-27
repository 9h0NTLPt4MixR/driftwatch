package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/patcher"
	"github.com/driftwatch/internal/reporter"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	patchSnapshotFile string
	patchOpsFile      string
	patchOutputFormat string
)

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Apply patch operations to a drift snapshot",
	Long: `Load a previously saved drift snapshot and apply suppress/override
operations defined in a JSON ops file. Outputs the patched results.`,
	RunE: runPatch,
}

func init() {
	rootCmd.AddCommand(patchCmd)
	patchCmd.Flags().StringVarP(&patchSnapshotFile, "snapshot", "s", "snapshot.json", "Path to drift snapshot file")
	patchCmd.Flags().StringVarP(&patchOpsFile, "ops", "o", "", "Path to JSON file containing patch operations (required)")
	patchCmd.Flags().StringVarP(&patchOutputFormat, "format", "f", "text", "Output format: text or json")
	_ = patchCmd.MarkFlagRequired("ops")
}

func runPatch(cmd *cobra.Command, args []string) error {
	results, err := snapshot.Load(patchSnapshotFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	ops, err := loadPatchOps(patchOpsFile)
	if err != nil {
		return fmt.Errorf("loading patch ops: %w", err)
	}

	patched := patcher.Apply(results, ops)

	// Unwrap to drift.Result slice for the reporter.
	unwrapped := make([]interface{}, 0, len(patched))
	for _, p := range patched {
		unwrapped = append(unwrapped, p)
	}

	if patchOutputFormat == "json" {
		return json.NewEncoder(os.Stdout).Encode(patched)
	}

	// Fall back to standard text reporter using the underlying results.
	driftResults := make([]interface{}, 0)
	_ = driftResults
	for _, p := range patched {
		fmt.Fprintf(os.Stdout, "service=%-20s drifted=%-5v patched_keys=%v\n",
			p.Service, p.Drifted, p.PatchedKeys)
	}
	_ = reporter.Write
	return nil
}

func loadPatchOps(path string) ([]patcher.Op, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ops []patcher.Op
	if err := json.NewDecoder(f).Decode(&ops); err != nil {
		return nil, err
	}
	return ops, nil
}
