// Package pipeline provides a composable, ordered execution pipeline
// for chaining drift scan post-processing steps (filter, rank, label,
// redact, validate, etc.) into a single reusable workflow.
package pipeline

import (
	"fmt"

	"github.com/driftwatch/internal/drift"
)

// StepFunc is a function that transforms a slice of drift results.
// Each step receives the output of the previous step.
type StepFunc func([]drift.Result) ([]drift.Result, error)

// Pipeline holds an ordered list of processing steps.
type Pipeline struct {
	steps []namedStep
}

type namedStep struct {
	name string
	fn   StepFunc
}

// New creates an empty Pipeline.
func New() *Pipeline {
	return &Pipeline{}
}

// Add appends a named step to the pipeline.
// The name is used for error reporting only.
func (p *Pipeline) Add(name string, fn StepFunc) *Pipeline {
	p.steps = append(p.steps, namedStep{name: name, fn: fn})
	return p
}

// Run executes each step in order, threading results through.
// If any step returns an error the pipeline halts and returns that error
// along with the name of the failing step.
func (p *Pipeline) Run(input []drift.Result) ([]drift.Result, error) {
	current := input
	for _, s := range p.steps {
		out, err := s.fn(current)
		if err != nil {
			return nil, fmt.Errorf("pipeline step %q: %w", s.name, err)
		}
		current = out
	}
	return current, nil
}

// Len returns the number of steps registered in the pipeline.
func (p *Pipeline) Len() int {
	return len(p.steps)
}

// StepNames returns the names of all registered steps in order.
func (p *Pipeline) StepNames() []string {
	names := make([]string, len(p.steps))
	for i, s := range p.steps {
		names[i] = s.name
	}
	return names
}
