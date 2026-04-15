package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "driftwatch",
	Short: "Detect configuration drift between deployed services and version control",
	Long: `driftwatch compares the declared state of your services in version control
against their actual deployed configuration and reports any drift.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .driftwatch.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}
