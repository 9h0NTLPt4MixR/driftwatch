package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/comparator"
	"github.com/spf13/cobra"
)

var (
	compareExpectedFile string
	compareActualFile   string
	compareRulesFile    string
	compareOutputJSON   bool
)

func init() {
	compareCmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare two config snapshots using per-key strategies",
		RunE:  runCompare,
	}
	compareCmd.Flags().StringVar(&compareExpectedFile, "expected", "", "Path to expected (declared) config JSON (required)")
	compareCmd.Flags().StringVar(&compareActualFile, "actual", "", "Path to actual (live) config JSON (required)")
	compareCmd.Flags().StringVar(&compareRulesFile, "rules", "", "Path to comparator rules JSON (optional)")
	compareCmd.Flags().BoolVar(&compareOutputJSON, "json", false, "Output results as JSON")
	_ = compareCmd.MarkFlagRequired("expected")
	_ = compareCmd.MarkFlagRequired("actual")
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, args []string) error {
	expected, err := loadJSONMap(compareExpectedFile)
	if err != nil {
		return fmt.Errorf("loading expected config: %w", err)
	}
	actual, err := loadJSONMap(compareActualFile)
	if err != nil {
		return fmt.Errorf("loading actual config: %w", err)
	}

	var rules []comparator.Rule
	if compareRulesFile != "" {
		data, err := os.ReadFile(compareRulesFile)
		if err != nil {
			return fmt.Errorf("reading rules file: %w", err)
		}
		if err := json.Unmarshal(data, &rules); err != nil {
			return fmt.Errorf("parsing rules: %w", err)
		}
	}

	c := comparator.New(rules)
	results := c.Compare(expected, actual)

	if compareOutputJSON {
		return json.NewEncoder(os.Stdout).Encode(results)
	}

	drifted := 0
	for _, r := range results {
		if r.Drifted {
			drifted++
			fmt.Printf("  DRIFT  %-30s  %s\n", r.Key, r.Reason)
		} else {
			fmt.Printf("  OK     %-30s\n", r.Key)
		}
	}
	fmt.Printf("\n%d key(s) checked, %d drifted.\n", len(results), drifted)
	return nil
}

func loadJSONMap(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
