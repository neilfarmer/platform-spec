package output

import (
	"fmt"
	"strings"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
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
		color := getStatusColor(result.Status)
		sb.WriteString(fmt.Sprintf("%s%s %s%s (%.2fs)\n",
			color, symbol, result.Name, colorReset, result.Duration.Seconds()))

		if result.Message != "" && result.Status != core.StatusPass {
			sb.WriteString(fmt.Sprintf("  %s%s%s\n", color, result.Message, colorReset))
		}
	}

	sb.WriteString("\n")

	// Summary
	_, passed, failed, skipped, errors := results.Summary()
	sb.WriteString(fmt.Sprintf("Tests: %d passed, %d failed, %d skipped, %d errors\n",
		passed, failed, skipped, errors))
	sb.WriteString(fmt.Sprintf("Duration: %.2fs\n", results.Duration.Seconds()))

	if results.Success() {
		sb.WriteString(fmt.Sprintf("Status: %s%sPASSED%s\n", colorBold, colorGreen, colorReset))
	} else {
		sb.WriteString(fmt.Sprintf("Status: %s%sFAILED%s\n", colorBold, colorRed, colorReset))
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

func getStatusColor(status core.Status) string {
	switch status {
	case core.StatusPass:
		return colorGreen
	case core.StatusFail:
		return colorRed
	case core.StatusSkip:
		return colorYellow
	case core.StatusError:
		return colorYellow
	default:
		return colorReset
	}
}
