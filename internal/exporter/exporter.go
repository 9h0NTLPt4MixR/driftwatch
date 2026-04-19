// Package exporter writes drift scan results to external formats (CSV, Markdown).
package exporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/user/driftwatch/internal/drift"
)

// Format represents the export format.
type Format string

const (
	FormatCSV      Format = "csv"
	FormatMarkdown Format = "markdown"
)

// Write exports results in the given format to w.
func Write(w io.Writer, results []drift.Result, format Format) error {
	switch format {
	case FormatCSV:
		return writeCSV(w, results)
	case FormatMarkdown:
		return writeMarkdown(w, results)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

func writeCSV(w io.Writer, results []drift.Result) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"service", "key", "expected", "actual"}); err != nil {
		return err
	}
	for _, r := range results {
		for _, d := range r.Diffs {
			if err := cw.Write([]string{r.Service, d.Key, d.Expected, d.Actual}); err != nil {
				return err
			}
		}
		if !r.HasDrift {
			if err := cw.Write([]string{r.Service, "", "", ""}); err != nil {
				return err
			}
		}
	}
	cw.Flush()
	return cw.Error()
}

func writeMarkdown(w io.Writer, results []drift.Result) error {
	fmt.Fprintln(w, "| Service | Key | Expected | Actual |")
	fmt.Fprintln(w, "|---------|-----|----------|--------|")
	for _, r := range results {
		if !r.HasDrift {
			fmt.Fprintf(w, "| %s | — | — | — |\n", r.Service)
			continue
		}
		for _, d := range r.Diffs {
			expected := strings.ReplaceAll(d.Expected, "|", "\\|")
			actual := strings.ReplaceAll(d.Actual, "|", "\\|")
			fmt.Fprintf(w, "| %s | %s | %s | %s |\n", r.Service, d.Key, expected, actual)
		}
	}
	return nil
}
