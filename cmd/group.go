package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/fetcher"
	"github.com/driftwatch/internal/grouper"
	"github.com/spf13/cobra"
)

var (
	groupByFlag  string
	groupTagKey  string
)

func init() {
	groupCmd := &cobra.Command{
		Use:   "group",
		Short: "Group drift results by service, status, or tag",
		RunE:  runGroup,
	}
	groupCmd.Flags().StringVar(&groupByFlag, "by", "status", "Grouping dimension: service | status | tag")
	groupCmd.Flags().StringVar(&groupTagKey, "tag-key", "env", "Tag key to group by when --by=tag")
	groupCmd.Flags().StringVar(&cfgFile, "config", "driftwatch.yaml", "Path to config file")
	rootCmd.AddCommand(groupCmd)
}

func runGroup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	f := fetcher.New(nil)
	var results []drift.Result
	for _, svc := range cfg.Services {
		live, err := f.Fetch(svc.Endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: fetch %s: %v\n", svc.Name, err)
			continue
		}
		results = append(results, drift.Scan(svc, live))
	}

	by := grouper.GroupByStatus
	switch strings.ToLower(groupByFlag) {
	case "service":
		by = grouper.GroupByService
	case "tag":
		by = grouper.GroupByTag
	}

	groups := grouper.Compute(results, grouper.Options{By: by, TagKey: groupTagKey})
	for _, line := range grouper.Summary(groups) {
		fmt.Println(line)
	}
	return nil
}
