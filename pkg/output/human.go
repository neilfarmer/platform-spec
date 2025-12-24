package output

import (
	"fmt"
	"strings"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// FormatHuman formats test results in human-readable format
func FormatHuman(results *core.TestResults) string {
	var sb strings.Builder

	sb.WriteString("Platform-Spec Test Results\n")
	sb.WriteString("==========================\n\n")

	if results.SpecName != "" {
		sb.WriteString(fmt.Sprintf("Spec: %s\n", results.SpecName))
	}
	if results.Target != "" {
		sb.WriteString(fmt.Sprintf("Target: %s\n", results.Target))
	}
	sb.WriteString("\n")

	// Print individual test results
	for _, result := range results.Results {
		symbol := getStatusSymbol(result.Status)
		sb.WriteString(fmt.Sprintf("%s %s (%.2fs)\n", symbol, result.Name, result.Duration.Seconds()))

		if result.Message != "" && result.Status != core.StatusPass {
			sb.WriteString(fmt.Sprintf("  %s\n", result.Message))
		}
	}

	sb.WriteString("\n")

	// Summary
	_, passed, failed, skipped, errors := results.Summary()
	sb.WriteString(fmt.Sprintf("Tests: %d passed, %d failed, %d skipped, %d errors\n",
		passed, failed, skipped, errors))
	sb.WriteString(fmt.Sprintf("Duration: %.2fs\n", results.Duration.Seconds()))

	if results.Success() {
		sb.WriteString("Status: PASSED\n")
	} else {
		sb.WriteString("Status: FAILED\n")
	}

	return sb.String()
}

func getStatusSymbol(status core.Status) string {
	switch status {
	case core.StatusPass:
		return "✓"
	case core.StatusFail:
		return "✗"
	case core.StatusSkip:
		return "○"
	case core.StatusError:
		return "⚠"
	default:
		return "?"
	}
}
