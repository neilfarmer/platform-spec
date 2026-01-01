package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
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

// FormatMultiHostHuman formats multi-host test results in human-readable format
func FormatMultiHostHuman(results *core.MultiHostResults) string {
	var sb strings.Builder

	// Print results for each host
	for i, host := range results.Hosts {
		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(strings.Repeat("=", 40) + "\n")
		sb.WriteString(fmt.Sprintf("Testing: %s\n", host.Target))
		sb.WriteString(strings.Repeat("=", 40) + "\n\n")

		if !host.Connected {
			// Connection error
			sb.WriteString(applyColor(colorRed, fmt.Sprintf("Connection failed: %v\n", host.ConnectionError)))
			sb.WriteString(applyColor(colorBold+colorRed, "❌ FAILED"))
			sb.WriteString("\n")
		} else {
			// Show results for each spec
			for _, specResult := range host.SpecResults {
				if specResult.SpecName != "" {
					sb.WriteString(fmt.Sprintf("Spec: %s\n\n", specResult.SpecName))
				}

				// Print individual test results
				for _, result := range specResult.Results {
					symbol := getStatusSymbol(result.Status)
					color := getStatusColor(result.Status)
					sb.WriteString(fmt.Sprintf("%s (%.2fs)\n",
						applyColor(color, symbol+" "+result.Name), result.Duration.Seconds()))

					if result.Message != "" && result.Status != core.StatusPass {
						sb.WriteString(fmt.Sprintf("  %s\n", applyColor(color, result.Message)))
					}
				}

				sb.WriteString("\n")

				// Summary for this spec
				_, passed, failed, skipped, errors := specResult.Summary()
				sb.WriteString(fmt.Sprintf("Tests: %d passed, %d failed, %d skipped, %d errors\n",
					passed, failed, skipped, errors))
				sb.WriteString(fmt.Sprintf("Duration: %.2fs\n", specResult.Duration.Seconds()))
				sb.WriteString("\n")
			}

			// Overall status for this host
			if host.Success() {
				sb.WriteString(applyColor(colorBold+colorGreen, "✅ PASSED"))
			} else {
				sb.WriteString(applyColor(colorBold+colorRed, "❌ FAILED"))
			}
			sb.WriteString("\n")
		}
	}

	// Summary section
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("=", 80) + "\n")
	sb.WriteString("Results Summary\n")
	sb.WriteString(strings.Repeat("=", 80) + "\n\n")

	totalHosts, passedHosts, failedHosts, connectionErrors := results.Summary()
	sb.WriteString(fmt.Sprintf("Total hosts: %d\n", totalHosts))
	sb.WriteString(fmt.Sprintf("Passed: %d %s\n", passedHosts, applyColor(colorGreen, "✅")))
	sb.WriteString(fmt.Sprintf("Failed: %d %s\n", failedHosts, applyColor(colorRed, "❌")))
	if connectionErrors > 0 {
		sb.WriteString(fmt.Sprintf("Connection errors: %d\n", connectionErrors))
	}

	// Results table using go-pretty
	sb.WriteString("\n")

	t := table.NewWriter()
	t.SetOutputMirror(&sb)
	t.AppendHeader(table.Row{"Host", "Status", "Details"})

	// Configure table style with ASCII borders (+ | -)
	style := table.Style{
		Name: "ASCII",
		Box: table.BoxStyle{
			BottomLeft:       "+",
			BottomRight:      "+",
			BottomSeparator:  "+",
			Left:             "|",
			LeftSeparator:    "+",
			MiddleHorizontal: "-",
			MiddleSeparator:  "+",
			MiddleVertical:   "|",
			PaddingLeft:      " ",
			PaddingRight:     " ",
			Right:            "|",
			RightSeparator:   "+",
			TopLeft:          "+",
			TopRight:         "+",
			TopSeparator:     "+",
		},
		Options: table.Options{
			DrawBorder:      true,
			SeparateColumns: true,
			SeparateHeader:  true,
			SeparateRows:    false,
		},
	}
	t.SetStyle(style)

	// Add rows
	for _, host := range results.Hosts {
		status := ""
		details := ""

		if !host.Connected {
			// Connection error - simplified message
			status = applyColor(colorBold+colorRed, "FAILED")
			details = applyColor(colorRed, "• Unable to connect via SSH")
		} else if host.Success() {
			// All tests passed
			status = applyColor(colorBold+colorGreen, "PASSED")
			details = applyColor(colorGreen, "All tests passed")
		} else {
			// Some tests failed - show bullet list of failures
			status = applyColor(colorBold+colorRed, "FAILED")

			// Collect failed and error test names
			var failedTests []string

			for _, specResult := range host.SpecResults {
				for _, result := range specResult.Results {
					if result.Status == core.StatusFail || result.Status == core.StatusError {
						failedTests = append(failedTests, result.Name)
					}
				}
			}

			// Create bullet list
			if len(failedTests) > 0 {
				bullets := make([]string, len(failedTests))
				for i, testName := range failedTests {
					bullets[i] = "• " + testName
				}
				details = applyColor(colorRed, strings.Join(bullets, "\n"))
			}
		}

		t.AppendRow(table.Row{host.Target, status, details})
	}

	t.Render()

	return sb.String()
}

// wrapText wraps text to fit within a specified width, preserving ANSI color codes
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	// Strip ANSI codes to get visible text
	visibleText := stripAnsiCodes(text)

	// If text fits, return as-is
	if len(visibleText) <= width {
		return []string{text}
	}

	// Extract color codes from the beginning
	colorPrefix := ""
	colorSuffix := ""
	if strings.Contains(text, "\033[") {
		// Find the first non-ANSI character position
		for i := 0; i < len(text); i++ {
			if text[i] == '\033' {
				// Find end of ANSI sequence
				end := strings.Index(text[i:], "m")
				if end != -1 {
					colorPrefix = text[:i+end+1]
					break
				}
			}
		}
		// Check for reset code at the end
		if strings.HasSuffix(text, colorReset) {
			colorSuffix = colorReset
		}
	}

	// Wrap the visible text
	var lines []string
	words := strings.Fields(visibleText)
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, colorPrefix+currentLine+colorSuffix)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, colorPrefix+currentLine+colorSuffix)
	}

	if len(lines) == 0 {
		return []string{text}
	}

	return lines
}

// stripAnsiCodes removes ANSI color codes from a string
func stripAnsiCodes(text string) string {
	result := ""
	i := 0
	for i < len(text) {
		if text[i] == '\033' && i+1 < len(text) && text[i+1] == '[' {
			// Find the end of the ANSI sequence
			end := strings.Index(text[i:], "m")
			if end != -1 {
				i += end + 1
				continue
			}
		}
		result += string(text[i])
		i++
	}
	return result
}

// padRight pads text to the right while preserving ANSI codes
func padRight(text string, width int) string {
	visibleLen := len(stripAnsiCodes(text))
	if visibleLen >= width {
		return text
	}
	return text + strings.Repeat(" ", width-visibleLen)
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

// ClearProgressLine clears the current progress line
func ClearProgressLine() {
	fmt.Fprint(os.Stderr, "\r\033[K") // Carriage return + clear line
}
