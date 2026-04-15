package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/driftwatch/internal/config"
	"github.com/user/driftwatch/internal/drift"
)

var (
	outputFormat string
	serviceName  string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for configuration drift",
	Long:  `Compare declared service configuration against the live deployed state.`,
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json")
	scanCmd.Flags().StringVarP(&serviceName, "service", "s", "", "limit scan to a specific service")
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	services := cfg.Services
	if serviceName != "" {
		svc, ok := cfg.Services[serviceName]
		if !ok {
			return fmt.Errorf("service %q not found in config", serviceName)
		}
		services = map[string]config.Service{serviceName: svc}
	}

	results, err := drift.Scan(services, verbose)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if err := drift.Report(results, outputFormat, os.Stdout); err != nil {
		return fmt.Errorf("report failed: %w", err)
	}

	if drift.HasDrift(results) {
		os.Exit(2)
	}
	return nil
}
