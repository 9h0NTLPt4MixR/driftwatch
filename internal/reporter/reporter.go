package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/driftwatch/internal/drift"
)

// Format represents the output format for drift reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Write renders drift results to w in the given format.
func Write(w io.Writer, results []drift.Result, format Format) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, results)
	default:
		return writeText(w, results)
	}
}

func writeText(w io.Writer, results []drift.Result) error {
	for _, r := range results {
		if len(r.Diffs) == 0 {
			_, err := fmt.Fprintf(w, "[OK]   %s — no drift detected\n", r.Service)
			if err != nil {
				return err
			}
			continue
		}
		_, err := fmt.Fprintf(w, "[DRIFT] %s\n", r.Service)
		if err != nil {
			return err
		}
		for _, d := range r.Diffs {
			_, err := fmt.Fprintf(w, "  %-20s expected=%q actual=%q\n", d.Key, d.Expected, d.Actual)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeJSON(w io.Writer, results []drift.Result) error {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("  {\"service\": %q, \"drift\": %v, \"diffs\": [", r.Service, len(r.Diffs) > 0))
		for j, d := range r.Diffs {
			sb.WriteString(fmt.Sprintf("{\"key\": %q, \"expected\": %q, \"actual\": %q}", d.Key, d.Expected, d.Actual))
			if j < len(r.Diffs)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]}")
		if i < len(results)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("]\n")
	_, err := fmt.Fprint(w, sb.String())
	return err
}
