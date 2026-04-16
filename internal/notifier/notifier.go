// Package notifier provides alerting capabilities when drift is detected.
package notifier

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/driftwatch/internal/drift"
)

// Channel represents a notification output channel.
type Channel string

const (
	ChannelStdout Channel = "stdout"
	ChannelStderr Channel = "stderr"
)

// Options configures notification behaviour.
type Options struct {
	Channel  Channel
	OnlyDrift bool
}

// Notify writes a summary notification for the given scan results.
func Notify(results []drift.Result, opts Options) error {
	w := writerFor(opts.Channel)

	drifted := 0
	for _, r := range results {
		if r.HasDrift {
			drifted++
		}
	}

	if opts.OnlyDrift && drifted == 0 {
		return nil
	}

	lines := []string{
		fmt.Sprintf("[driftwatch] scan complete: %d service(s) checked, %d with drift", len(results), drifted),
	}

	for _, r := range results {
		if !r.HasDrift && opts.OnlyDrift {
			continue
		}
		status := "OK"
		if r.HasDrift {
			status = fmt.Sprintf("DRIFT (%d key(s))", len(r.Diffs))
		}
		lines = append(lines, fmt.Sprintf("  %-30s %s", r.Service, status))
	}

	_, err := fmt.Fprintln(w, strings.Join(lines, "\n"))
	return err
}

func writerFor(ch Channel) io.Writer {
	if ch == ChannelStderr {
		return os.Stderr
	}
	return os.Stdout
}
