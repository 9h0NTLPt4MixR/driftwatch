// Package scheduler provides periodic drift scan scheduling,
// allowing driftwatch to run scans automatically at a configured interval
// and emit results to a notifier or history recorder.
package scheduler

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/history"
	"github.com/driftwatch/internal/notifier"
	"github.com/driftwatch/internal/config"
	"github.com/driftwatch/internal/fetcher"
)

// Options controls the behaviour of the scheduler.
type Options struct {
	// Interval between scans. Must be > 0.
	Interval time.Duration

	// HistoryFile is the path used to persist scan history.
	// Leave empty to skip history recording.
	HistoryFile string

	// NotifyOnlyDrift suppresses notifications for clean services.
	NotifyOnlyDrift bool

	// Out is the writer used for notification output.
	// Defaults to os.Stdout when nil.
	Out io.Writer
}

// Scheduler runs drift scans on a fixed interval.
type Scheduler struct {
	cfg  *config.Config
	opts Options
}

// New creates a Scheduler for the given config and options.
// Returns an error if the interval is not positive.
func New(cfg *config.Config, opts Options) (*Scheduler, error) {
	if opts.Interval <= 0 {
		return nil, fmt.Errorf("scheduler: interval must be positive, got %s", opts.Interval)
	}
	return &Scheduler{cfg: cfg, opts: opts}, nil
}

// Run starts the scheduling loop and blocks until ctx is cancelled.
// Each tick triggers a full drift scan across all configured services.
func (s *Scheduler) Run(ctx context.Context) error {
	log.Printf("scheduler: starting — interval=%s services=%d", s.opts.Interval, len(s.cfg.Services))

	ticker := time.NewTicker(s.opts.Interval)
	defer ticker.Stop()

	// Run an immediate scan before the first tick so operators get
	// feedback right away when the scheduler starts.
	if err := s.runOnce(ctx); err != nil {
		log.Printf("scheduler: scan error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler: context cancelled, stopping")
			return ctx.Err()
		case t := <-ticker.C:
			log.Printf("scheduler: tick at %s", t.Format(time.RFC3339))
			if err := s.runOnce(ctx); err != nil {
				// Log but do not stop the loop on a transient scan error.
				log.Printf("scheduler: scan error: %v", err)
			}
		}
	}
}

// runOnce executes a single drift scan, records history, and notifies.
func (s *Scheduler) runOnce(ctx context.Context) error {
	f := fetcher.New(nil) // uses default http.Client

	results, err := drift.Scan(ctx, s.cfg, f)
	if err != nil {
		return fmt.Errorf("drift scan: %w", err)
	}

	if s.opts.HistoryFile != "" {
		if herr := history.Record(s.opts.HistoryFile, results); herr != nil {
			log.Printf("scheduler: history record error: %v", herr)
		}
	}

	nopts := notifier.Options{
		OnlyDrift: s.opts.NotifyOnlyDrift,
		Out:       s.opts.Out,
	}
	if nerr := notifier.Notify(results, nopts); nerr != nil {
		log.Printf("scheduler: notify error: %v", nerr)
	}

	return nil
}
