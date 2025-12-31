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

// Global flag to control color output
var NoColor = false

// applyColor returns the colored string if colors are enabled, otherwise returns plain string
func applyColor(color, text string) string {
	if NoColor {
		return text
	}
	return color + text + colorReset
}

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
		sb.WriteString(fmt.Sprintf("%s (%.2fs)\n",
			applyColor(color, symbol+" "+result.Name), result.Duration.Seconds()))

		if result.Message != "" && result.Status != core.StatusPass {
			sb.WriteString(fmt.Sprintf("  %s\n", applyColor(color, result.Message)))
		}
	}

	sb.WriteString("\n")

	// Summary
	_, passed, failed, skipped, errors := results.Summary()
	sb.WriteString(fmt.Sprintf("Tests: %d passed, %d failed, %d skipped, %d errors\n",
		passed, failed, skipped, errors))
	sb.WriteString(fmt.Sprintf("Duration: %.2fs\n", results.Duration.Seconds()))

	sb.WriteString("\n")
	if results.Success() {
		sb.WriteString(applyColor(colorBold+colorGreen, "✅ PASSED"))
	} else {
		sb.WriteString(applyColor(colorBold+colorRed, "❌ FAILED"))
	}
	sb.WriteString("\n")

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

// PrintFailed prints a FAILED status message (for connection errors and other failures)
func PrintFailed() string {
	return "\n" + applyColor(colorBold+colorRed, "❌ FAILED") + "\n"
}

// PrintPassed prints a PASSED status message
func PrintPassed() string {
	return "\n" + applyColor(colorBold+colorGreen, "✅ PASSED") + "\n"
}
