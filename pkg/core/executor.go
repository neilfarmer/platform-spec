package core

import (
	"context"
	"strings"
	"time"
)

// Executor executes tests against a provider
type Executor struct {
	spec     *Spec
	provider Provider
	plugins  []Plugin
}

// Provider interface that all providers must implement
type Provider interface {
	ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error)
}

// Plugin interface that all test plugins must implement
type Plugin interface {
	// Execute runs all tests handled by this plugin
	// Returns results and a boolean indicating whether to stop (for fail-fast)
	Execute(ctx context.Context, spec *Spec, provider Provider, failFast bool) ([]Result, bool)
}

// NewExecutor creates a new executor with the given plugins
func NewExecutor(spec *Spec, provider Provider, plugins ...Plugin) *Executor {
	return &Executor{
		spec:     spec,
		provider: provider,
		plugins:  plugins,
	}
}

// Execute runs all tests in the spec using registered plugins
func (e *Executor) Execute(ctx context.Context) (*TestResults, error) {
	startTime := time.Now()

	results := &TestResults{
		SpecName:  e.spec.Metadata.Name,
		StartTime: startTime,
		Results:   []Result{},
	}

	// Execute each plugin in order
	for _, plugin := range e.plugins {
		pluginResults, shouldStop := plugin.Execute(ctx, e.spec, e.provider, e.spec.Config.FailFast)
		results.Results = append(results.Results, pluginResults...)

		// If plugin indicates we should stop (fail-fast), break
		if shouldStop {
			break
		}
	}

	results.Duration = time.Since(startTime)
	return results, nil
}






// ShellQuote quotes a string for safe use in shell commands
func ShellQuote(s string) string {
	// Simple shell quoting - escape single quotes and wrap in single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}






