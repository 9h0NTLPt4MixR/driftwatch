package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/labeler"
	"github.com/driftwatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	labelSnapshotFile string
	labelRulesFile    string
	labelFilterLabel  string
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Assign and filter labels on drift results",
	RunE:  runLabel,
}

func init() {
	rootCmd.AddCommand(labelCmd)
	labelCmd.Flags().StringVar(&labelSnapshotFile, "snapshot", "drift-snapshot.json", "Snapshot file to load results from")
	labelCmd.Flags().StringVar(&labelRulesFile, "rules", "", "JSON file containing labeling rules (required)")
	labelCmd.Flags().StringVar(&labelFilterLabel, "filter", "", "Only output results carrying this label")
	_ = labelCmd.MarkFlagRequired("rules")
}

func runLabel(cmd *cobra.Command, args []string) error {
	results, err := snapshot.Load(labelSnapshotFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	rulesData, err := os.ReadFile(labelRulesFile)
	if err != nil {
		return fmt.Errorf("reading rules file: %w", err)
	}
	var rules []labeler.Rule
	if err := json.Unmarshal(rulesData, &rules); err != nil {
		return fmt.Errorf("parsing rules: %w", err)
	}

	labeled := labeler.Apply(results, rules)

	if labelFilterLabel != "" {
		labeled = labeler.FilterByLabel(labeled, labelFilterLabel)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(labeled)
}
