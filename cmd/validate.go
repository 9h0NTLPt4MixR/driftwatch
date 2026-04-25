package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/validator"
	"github.com/spf13/cobra"
)

var (
	validateMaxServices int
	validateMaxKeys     int
	validateClean       []string
	validatePolicyFile  string
)

func init() {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate drift results against threshold rules",
		RunE:  runValidate,
	}
	validateCmd.Flags().IntVar(&validateMaxServices, "max-drifted-services", -1, "maximum number of drifted services allowed (-1 = unlimited)")
	validateCmd.Flags().IntVar(&validateMaxKeys, "max-drifted-keys", -1, "maximum drifted keys per service (-1 = unlimited)")
	validateCmd.Flags().StringSliceVar(&validateClean, "require-clean", nil, "comma-separated list of services that must have zero drift")
	validateCmd.Flags().StringVar(&validatePolicyFile, "rules", "", "JSON file containing a validator.Rule object (overrides flags)")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	f := fetcher.New(nil)
	results, err := drift.Scan(cfg.Services, f)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	rule := validator.Rule{
		MaxDriftedServices:   validateMaxServices,
		MaxDriftedKeys:       validateMaxKeys,
		RequireCleanServices: validateClean,
	}

	if validatePolicyFile != "" {
		data, readErr := os.ReadFile(validatePolicyFile)
		if readErr != nil {
			return fmt.Errorf("read rules file: %w", readErr)
		}
		if jsonErr := json.Unmarshal(data, &rule); jsonErr != nil {
			return fmt.Errorf("parse rules file: %w", jsonErr)
		}
	}

	out := validator.Validate(results, rule)
	if out.Passed {
		fmt.Println("✔  All validation rules passed.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "✘  Validation failed — %d violation(s):\n", len(out.Violations))
	for _, v := range out.Violations {
		fmt.Fprintf(os.Stderr, "   [%s] %s\n", v.Service, v.Message)
	}
	return fmt.Errorf("validation failed with %d violation(s)", len(out.Violations))
}

// ensure drift.Result is used (avoids import cycle in test file helper)
var _ drift.Result
