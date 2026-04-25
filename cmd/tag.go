package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/reporter"
	"github.com/driftwatch/internal/tagger"
	"github.com/spf13/cobra"
)

var (
	tagMapFile  string
	filterTag   string
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Scan services and filter results by tag",
	Long: `Runs a drift scan and applies a tag map to results.
Optionally filter output to only services matching a specific tag.`,
	RunE: runTag,
}

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.Flags().StringVar(&tagMapFile, "tag-map", "", "Path to JSON tag map file (required)")
	tagCmd.Flags().StringVar(&filterTag, "filter-tag", "", "Only show results with this tag")
	tagCmd.Flags().StringVar(&cfgFile, "config", "driftwatch.yaml", "Path to config file")
	tagCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	_ = tagCmd.MarkFlagRequired("tag-map")
}

func runTag(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	tm, err := tagger.LoadTagMap(tagMapFile)
	if err != nil {
		return fmt.Errorf("load tag map: %w", err)
	}

	f := fetcher.New(nil)
	results := drift.Scan(cfg.Services, f)

	tagged := tagger.Apply(results, tm)

	if filterTag != "" {
		tagged = tagger.FilterByTag(tagged, filterTag)
	}

	return reporter.Write(os.Stdout, tagged, outputFormat)
}
