package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/resolver"
	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Show resolved endpoints for all configured services",
	Long: `Prints the endpoint that driftwatch would use for each service,
taking into account config, environment overrides, and programmatic overrides.

Useful for debugging endpoint resolution before running a scan.`,
	RunE: runResolve,
}

var resolveService string

func init() {
	rootCmd.AddCommand(resolveCmd)
	resolveCmd.Flags().StringVar(&resolveService, "service", "", "Resolve a single named service")
}

func runResolve(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	base := make(map[string]string, len(cfg.Services))
	for _, svc := range cfg.Services {
		base[svc.Name] = svc.Endpoint
	}

	r := resolver.New(base)

	if resolveService != "" {
		ep, err := r.Resolve(resolveService)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", ep)
		return nil
	}

	entries := r.ResolveAll()
	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no services configured")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tENDPOINT\tSOURCE")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.Service, e.Endpoint, e.Source)
	}
	return w.Flush()
}
