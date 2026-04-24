package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/driftwatch/internal/config"
	"github.com/user/driftwatch/internal/drift"
	"github.com/user/driftwatch/internal/fetcher"
	"github.com/user/driftwatch/internal/policy"
)

var (
	policyFile   string
	policyStrict bool
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Evaluate drift results against policy rules",
	Long:  "Fetch live config, compare against declared state, then check results against policy rules defined in a JSON file.",
	RunE:  runPolicy,
}

func init() {
	policyCmd.Flags().StringVar(&cfgFile, "config", "driftwatch.yaml", "path to service config file")
	policyCmd.Flags().StringVar(&policyFile, "policy", "policy.json", "path to policy rules JSON file")
	policyCmd.Flags().BoolVar(&policyStrict, "strict", false, "exit with non-zero code when violations are found")
	rootCmd.AddCommand(policyCmd)
}

func runPolicy(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	f := fetcher.New(cfg.Timeout)
	var results []drift.Result
	for _, svc := range cfg.Services {
		actual, err := f.Fetch(svc.Endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: fetch %s: %v\n", svc.Name, err)
			continue
		}
		results = append(results, drift.Scan(svc, actual))
	}

	rules, err := loadPolicyRules(policyFile)
	if err != nil {
		return fmt.Errorf("loading policy: %w", err)
	}

	violations := policy.Evaluate(results, rules)
	if len(violations) == 0 {
		fmt.Println("✓ No policy violations detected.")
		return nil
	}

	fmt.Printf("✗ %d policy violation(s) found:\n", len(violations))
	for _, v := range violations {
		fmt.Printf("  %s\n", v)
	}

	if policyStrict {
		os.Exit(1)
	}
	return nil
}

func loadPolicyRules(path string) ([]policy.Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rules []policy.Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
