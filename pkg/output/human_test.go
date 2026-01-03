package output

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestFormatHuman(t *testing.T) {
	tests := []struct {
		name     string
		results  *core.TestResults
		contains []string
	}{
		{
			name: "all passed",
			results: &core.TestResults{
				SpecName: "Test Suite",
				Target:   "ubuntu@host",
				Duration: 1 * time.Second,
				Results: []core.Result{
					{
						Name:     "Test 1",
						Status:   core.StatusPass,
						Duration: 500 * time.Millisecond,
					},
					{
						Name:     "Test 2",
						Status:   core.StatusPass,
						Duration: 500 * time.Millisecond,
					},
				},
			},
			contains: []string{
				"Summary:",
				"2 passed",
				"1.00s",
			},
		},
		{
			name: "mixed results",
			results: &core.TestResults{
				SpecName: "Test Suite",
				Duration: 1 * time.Second,
				Results: []core.Result{
					{
						Name:     "Test 1",
						Status:   core.StatusPass,
						Duration: 250 * time.Millisecond,
					},
					{
						Name:     "Test 2",
						Status:   core.StatusFail,
						Message:  "Test failed",
						Duration: 250 * time.Millisecond,
					},
					{
						Name:     "Test 3",
						Status:   core.StatusSkip,
						Duration: 0,
					},
				},
			},
			contains: []string{
				"Summary:",
				"1 passed, 1 failed",
				"1 skipped",
				"1.00s",
			},
		},
		{
			name: "errors",
			results: &core.TestResults{
				Duration: 500 * time.Millisecond,
				Results: []core.Result{
					{
						Name:     "Test 1",
						Status:   core.StatusError,
						Message:  "Error occurred",
						Duration: 100 * time.Millisecond,
					},
				},
			},
			contains: []string{
				"Summary:",
				"0 passed, 1 failed",
				"0.50s",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatHuman(tt.results)

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatHuman() output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestGetStatusSymbol(t *testing.T) {
	tests := []struct {
		status core.Status
		want   string
	}{
		{core.StatusPass, "✓"},
		{core.StatusFail, "✗"},
		{core.StatusSkip, "○"},
		{core.StatusError, "⚠"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := getStatusSymbol(tt.status); got != tt.want {
				t.Errorf("getStatusSymbol(%v) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestGetStatusColor(t *testing.T) {
	tests := []struct {
		status core.Status
		want   string
	}{
		{core.StatusPass, colorGreen},
		{core.StatusFail, colorRed},
		{core.StatusSkip, colorYellow},
		{core.StatusError, colorYellow},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := getStatusColor(tt.status); got != tt.want {
				t.Errorf("getStatusColor(%v) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestApplyColor(t *testing.T) {
	tests := []struct {
		name     string
		noColor  bool
		color    string
		text     string
		wantText string
	}{
		{
			name:     "colors enabled - green text",
			noColor:  false,
			color:    colorGreen,
			text:     "PASSED",
			wantText: colorGreen + "PASSED" + colorReset,
		},
		{
			name:     "colors enabled - red text",
			noColor:  false,
			color:    colorRed,
			text:     "FAILED",
			wantText: colorRed + "FAILED" + colorReset,
		},
		{
			name:     "colors enabled - bold green",
			noColor:  false,
			color:    colorBold + colorGreen,
			text:     "✅ PASSED",
			wantText: colorBold + colorGreen + "✅ PASSED" + colorReset,
		},
		{
			name:     "colors disabled - should return plain text",
			noColor:  true,
			color:    colorGreen,
			text:     "PASSED",
			wantText: "PASSED",
		},
		{
			name:     "colors disabled - emojis preserved",
			noColor:  true,
			color:    colorBold + colorGreen,
			text:     "✅ PASSED",
			wantText: "✅ PASSED",
		},
		{
			name:     "colors disabled - symbols preserved",
			noColor:  true,
			color:    colorRed,
			text:     "✗ Test failed",
			wantText: "✗ Test failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original NoColor value
			originalNoColor := NoColor
			defer func() { NoColor = originalNoColor }()

			NoColor = tt.noColor
			got := applyColor(tt.color, tt.text)
			if got != tt.wantText {
				t.Errorf("applyColor() = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestFormatHumanWithColors(t *testing.T) {
	tests := []struct {
		name            string
		noColor         bool
		results         *core.TestResults
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:    "colors enabled - passed test shows green ANSI codes",
			noColor: false,
			results: &core.TestResults{
				Duration: 1 * time.Second,
				Results: []core.Result{
					{Name: "Test 1", Status: core.StatusPass, Duration: 500 * time.Millisecond},
				},
			},
			wantContains: []string{
				colorGreen,  // Green color code
				colorReset,  // Reset code
				colorBold,   // Bold for "Summary:"
				"1 passed",  // Summary text
			},
			wantNotContains: []string{},
		},
		{
			name:    "colors enabled - failed test shows red ANSI codes",
			noColor: false,
			results: &core.TestResults{
				Duration: 1 * time.Second,
				Results: []core.Result{
					{Name: "Test 1", Status: core.StatusFail, Message: "failed", Duration: 500 * time.Millisecond},
				},
			},
			wantContains: []string{
				colorRed,         // Red color code
				colorReset,       // Reset code
				colorBold,        // Bold for "Summary:"
				"0 passed, 1 failed", // Summary text
			},
			wantNotContains: []string{},
		},
		{
			name:    "colors disabled - no ANSI codes",
			noColor: true,
			results: &core.TestResults{
				Duration: 1 * time.Second,
				Results: []core.Result{
					{Name: "Test 1", Status: core.StatusPass, Duration: 500 * time.Millisecond},
				},
			},
			wantContains: []string{
				"Summary:",  // Summary text remains
				"1 passed",  // Summary text
			},
			wantNotContains: []string{
				"\033[",     // No ANSI escape codes
				colorGreen,  // No color codes
				colorRed,    // No color codes
				colorReset,  // No reset codes
			},
		},
		{
			name:    "colors disabled - failed test no ANSI codes",
			noColor: true,
			results: &core.TestResults{
				Duration: 1 * time.Second,
				Results: []core.Result{
					{Name: "Test 1", Status: core.StatusFail, Message: "error", Duration: 500 * time.Millisecond},
				},
			},
			wantContains: []string{
				"Summary:",           // Summary text remains
				"0 passed, 1 failed", // Summary text
			},
			wantNotContains: []string{
				"\033[",    // No ANSI escape codes
				colorRed,   // No color codes
				colorReset, // No reset codes
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original NoColor value
			originalNoColor := NoColor
			defer func() { NoColor = originalNoColor }()

			NoColor = tt.noColor
			output := FormatHuman(tt.results)

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatHuman() output missing %q\nGot:\n%s", want, output)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("FormatHuman() output should not contain %q\nGot:\n%s", notWant, output)
				}
			}
		})
	}
}

func TestPrintFailed(t *testing.T) {
	tests := []struct {
		name            string
		noColor         bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:    "colors enabled",
			noColor: false,
			wantContains: []string{
				"❌ FAILED",
				colorBold,
				colorRed,
				colorReset,
			},
			wantNotContains: []string{},
		},
		{
			name:    "colors disabled",
			noColor: true,
			wantContains: []string{
				"❌ FAILED",
			},
			wantNotContains: []string{
				colorBold,
				colorRed,
				colorReset,
				"\033[",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original NoColor value
			originalNoColor := NoColor
			defer func() { NoColor = originalNoColor }()

			NoColor = tt.noColor
			output := PrintFailed()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("PrintFailed() output missing %q\nGot:\n%s", want, output)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("PrintFailed() output should not contain %q\nGot:\n%s", notWant, output)
				}
			}

			// Verify it starts and ends with newline
			if !strings.HasPrefix(output, "\n") {
				t.Errorf("PrintFailed() should start with newline")
			}
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("PrintFailed() should end with newline")
			}
		})
	}
}

func TestPrintPassed(t *testing.T) {
	tests := []struct {
		name            string
		noColor         bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:    "colors enabled",
			noColor: false,
			wantContains: []string{
				"✅ PASSED",
				colorBold,
				colorGreen,
				colorReset,
			},
			wantNotContains: []string{},
		},
		{
			name:    "colors disabled",
			noColor: true,
			wantContains: []string{
				"✅ PASSED",
			},
			wantNotContains: []string{
				colorBold,
				colorGreen,
				colorReset,
				"\033[",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original NoColor value
			originalNoColor := NoColor
			defer func() { NoColor = originalNoColor }()

			NoColor = tt.noColor
			output := PrintPassed()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("PrintPassed() output missing %q\nGot:\n%s", want, output)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("PrintPassed() output should not contain %q\nGot:\n%s", notWant, output)
				}
			}

			// Verify it starts and ends with newline
			if !strings.HasPrefix(output, "\n") {
				t.Errorf("PrintPassed() should start with newline")
			}
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("PrintPassed() should end with newline")
			}
		})
	}
}

func TestFormatMultiHostHuman(t *testing.T) {
	tests := []struct {
		name     string
		results  *core.MultiHostResults
		contains []string
	}{
		{
			name: "all hosts passed",
			results: &core.MultiHostResults{
				TotalDuration: 2 * time.Second,
				Hosts: []*core.HostResults{
					{
						Target:    "ubuntu@web-01.example.com",
						Connected: true,
						Duration:  1 * time.Second,
						SpecResults: []*core.TestResults{
							{
								SpecName: "web-server.yaml",
								Duration: 1 * time.Second,
								Results: []core.Result{
									{Name: "nginx installed", Status: core.StatusPass, Duration: 500 * time.Millisecond},
								},
							},
						},
					},
					{
						Target:    "ubuntu@web-02.example.com",
						Connected: true,
						Duration:  1 * time.Second,
						SpecResults: []*core.TestResults{
							{
								SpecName: "web-server.yaml",
								Duration: 1 * time.Second,
								Results: []core.Result{
									{Name: "nginx installed", Status: core.StatusPass, Duration: 500 * time.Millisecond},
								},
							},
						},
					},
				},
			},
			contains: []string{
				"Testing: ubuntu@web-01.example.com",
				"Testing: ubuntu@web-02.example.com",
				"Spec: web-server.yaml",
				"✓ nginx installed",
				"PASSED",
				"Summary",
				"Total hosts: 2",
				"Passed: 2",
				"Failed: 0",
			},
		},
		{
			name: "mixed host results",
			results: &core.MultiHostResults{
				TotalDuration: 2 * time.Second,
				Hosts: []*core.HostResults{
					{
						Target:    "ubuntu@web-01.example.com",
						Connected: true,
						Duration:  1 * time.Second,
						SpecResults: []*core.TestResults{
							{
								SpecName: "web-server.yaml",
								Duration: 1 * time.Second,
								Results: []core.Result{
									{Name: "nginx installed", Status: core.StatusPass, Duration: 500 * time.Millisecond},
								},
							},
						},
					},
					{
						Target:    "ubuntu@web-02.example.com",
						Connected: true,
						Duration:  1 * time.Second,
						SpecResults: []*core.TestResults{
							{
								SpecName: "web-server.yaml",
								Duration: 1 * time.Second,
								Results: []core.Result{
									{Name: "nginx installed", Status: core.StatusFail, Message: "not installed", Duration: 500 * time.Millisecond},
								},
							},
						},
					},
				},
			},
			contains: []string{
				"Testing: ubuntu@web-01.example.com",
				"Testing: ubuntu@web-02.example.com",
				"✓ nginx installed",
				"✗ nginx installed",
				"not installed",
				"Summary",
				"Total hosts: 2",
				"Passed: 1",
				"Failed: 1",
			},
		},
		{
			name: "connection error",
			results: &core.MultiHostResults{
				TotalDuration: 1 * time.Second,
				Hosts: []*core.HostResults{
					{
						Target:          "ubuntu@invalid-host.example.com",
						Connected:       false,
						ConnectionError: fmt.Errorf("connection refused"),
						Duration:        500 * time.Millisecond,
					},
				},
			},
			contains: []string{
				"Testing: ubuntu@invalid-host.example.com",
				"Connection failed: connection refused",
				"FAILED",
				"Summary",
				"Total hosts: 1",
				"Passed: 0",
				"Failed: 1",
				"Connection errors: 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatMultiHostHuman(tt.results)

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatMultiHostHuman() output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestMultiHostResults_Summary(t *testing.T) {
	tests := []struct {
		name              string
		results           *core.MultiHostResults
		wantTotal         int
		wantPassed        int
		wantFailed        int
		wantConnErrors    int
	}{
		{
			name: "all passed",
			results: &core.MultiHostResults{
				Hosts: []*core.HostResults{
					{Target: "host1", Connected: true, SpecResults: []*core.TestResults{{Results: []core.Result{{Status: core.StatusPass}}}}},
					{Target: "host2", Connected: true, SpecResults: []*core.TestResults{{Results: []core.Result{{Status: core.StatusPass}}}}},
				},
			},
			wantTotal:      2,
			wantPassed:     2,
			wantFailed:     0,
			wantConnErrors: 0,
		},
		{
			name: "one failed connection",
			results: &core.MultiHostResults{
				Hosts: []*core.HostResults{
					{Target: "host1", Connected: true, SpecResults: []*core.TestResults{{Results: []core.Result{{Status: core.StatusPass}}}}},
					{Target: "host2", Connected: false, ConnectionError: fmt.Errorf("connection refused")},
				},
			},
			wantTotal:      2,
			wantPassed:     1,
			wantFailed:     1,
			wantConnErrors: 1,
		},
		{
			name: "one test failed",
			results: &core.MultiHostResults{
				Hosts: []*core.HostResults{
					{Target: "host1", Connected: true, SpecResults: []*core.TestResults{{Results: []core.Result{{Status: core.StatusPass}}}}},
					{Target: "host2", Connected: true, SpecResults: []*core.TestResults{{Results: []core.Result{{Status: core.StatusFail}}}}},
				},
			},
			wantTotal:      2,
			wantPassed:     1,
			wantFailed:     1,
			wantConnErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, passed, failed, connErrors := tt.results.Summary()

			if total != tt.wantTotal {
				t.Errorf("Summary() total = %d, want %d", total, tt.wantTotal)
			}
			if passed != tt.wantPassed {
				t.Errorf("Summary() passed = %d, want %d", passed, tt.wantPassed)
			}
			if failed != tt.wantFailed {
				t.Errorf("Summary() failed = %d, want %d", failed, tt.wantFailed)
			}
			if connErrors != tt.wantConnErrors {
				t.Errorf("Summary() connErrors = %d, want %d", connErrors, tt.wantConnErrors)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  []string
	}{
		{
			name:  "short text fits in one line",
			text:  "All tests passed",
			width: 50,
			want:  []string{"All tests passed"},
		},
		{
			name:  "long text wraps at word boundaries",
			text:  "Connection error: ssh: connect to host db-01.example.com port 22: connection refused",
			width: 50,
			want: []string{
				"Connection error: ssh: connect to host",
				"db-01.example.com port 22: connection refused",
			},
		},
		{
			name:  "empty text",
			text:  "",
			width: 50,
			want:  []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)
			if len(got) != len(tt.want) {
				t.Errorf("wrapText() returned %d lines, want %d\nGot: %v\nWant: %v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrapText() line %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "plain text no codes",
			text: "Hello World",
			want: "Hello World",
		},
		{
			name: "text with red color",
			text: "\033[31mFAILED\033[0m",
			want: "FAILED",
		},
		{
			name: "text with bold and green",
			text: "\033[1m\033[32mPASSED\033[0m",
			want: "PASSED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnsiCodes(tt.text)
			if got != tt.want {
				t.Errorf("stripAnsiCodes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatMultiHostHuman_TableOutput(t *testing.T) {
	// Save and restore original NoColor value
	originalNoColor := NoColor
	defer func() { NoColor = originalNoColor }()
	NoColor = true // Disable colors for easier testing

	tests := []struct {
		name         string
		results      *core.MultiHostResults
		wantContains []string
	}{
		{
			name: "table shows all hosts with passed status",
			results: &core.MultiHostResults{
				Hosts: []*core.HostResults{
					{
						Target:    "web-01.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{{Name: "Test 1", Status: core.StatusPass}}},
						},
					},
					{
						Target:    "web-02.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{{Name: "Test 1", Status: core.StatusPass}}},
						},
					},
				},
			},
			wantContains: []string{
				"Results Summary",
				"Host",
				"Status",
				"Details",
				"web-01.example.com",
				"web-02.example.com",
				"PASSED",
				"All tests passed",
				"+---", // Top border
				"| ",   // Side borders
			},
		},
		{
			name: "table shows failed tests as bullets",
			results: &core.MultiHostResults{
				Hosts: []*core.HostResults{
					{
						Target:    "web-01.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Test 1", Status: core.StatusFail},
								{Name: "Test 2", Status: core.StatusFail},
							}},
						},
					},
				},
			},
			wantContains: []string{
				"web-01.example.com",
				"FAILED",
				"• Test 1",
				"• Test 2",
			},
		},
		{
			name: "table shows connection errors",
			results: &core.MultiHostResults{
				Hosts: []*core.HostResults{
					{
						Target:          "web-01.example.com",
						Connected:       false,
						ConnectionError: fmt.Errorf("connection refused"),
					},
				},
			},
			wantContains: []string{
				"web-01.example.com",
				"FAILED",
				"• Unable to connect via SSH",
			},
		},
		{
			name: "table shows mixed failed and error as bullets",
			results: &core.MultiHostResults{
				Hosts: []*core.HostResults{
					{
						Target:    "web-01.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Test 1", Status: core.StatusFail},
								{Name: "Test 2", Status: core.StatusError},
							}},
						},
					},
				},
			},
			wantContains: []string{
				"web-01.example.com",
				"FAILED",
				"• Test 1",
				"• Test 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatMultiHostHuman(tt.results)

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatMultiHostHuman() output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestFormatMultiHostSummaryTable(t *testing.T) {
	tests := []struct {
		name         string
		results      *core.MultiHostResults
		wantContains []string
	}{
		{
			name: "all hosts passed",
			results: &core.MultiHostResults{
				TotalDuration: 5 * time.Second,
				Hosts: []*core.HostResults{
					{
						Target:    "web-01.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Test 1", Status: core.StatusPass},
								{Name: "Test 2", Status: core.StatusPass},
								{Name: "Test 3", Status: core.StatusPass},
							}},
						},
					},
					{
						Target:    "web-02.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Test 1", Status: core.StatusPass},
								{Name: "Test 2", Status: core.StatusPass},
							}},
						},
					},
				},
			},
			wantContains: []string{
				"Summary:",
				"2/2 hosts passed",
				"5.00s",
				"web-01.example.com",
				"web-02.example.com",
				"PASSED",
				"All 3 tests passed",
				"All 2 tests passed",
				"+",
				"|",
				"-",
			},
		},
		{
			name: "mixed results with failures",
			results: &core.MultiHostResults{
				TotalDuration: 10 * time.Second,
				Hosts: []*core.HostResults{
					{
						Target:    "web-01.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Test 1", Status: core.StatusPass},
								{Name: "Test 2", Status: core.StatusFail},
								{Name: "Test 3", Status: core.StatusPass},
							}},
						},
					},
					{
						Target:    "web-02.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Test 1", Status: core.StatusPass},
								{Name: "Test 2", Status: core.StatusPass},
							}},
						},
					},
				},
			},
			wantContains: []string{
				"Summary:",
				"1 passed, 1 failed",
				"10.00s",
				"web-01.example.com",
				"FAILED",
				"2 passed, 1 failed",
				"• Test 2",
				"web-02.example.com",
				"PASSED",
			},
		},
		{
			name: "connection errors",
			results: &core.MultiHostResults{
				TotalDuration: 2 * time.Second,
				Hosts: []*core.HostResults{
					{
						Target:          "web-01.example.com",
						Connected:       false,
						ConnectionError: fmt.Errorf("connection refused"),
					},
					{
						Target:    "web-02.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Test 1", Status: core.StatusPass},
							}},
						},
					},
				},
			},
			wantContains: []string{
				"Summary:",
				"1 passed, 1 failed",
				"2.00s",
				"web-01.example.com",
				"FAILED",
				"• Unable to connect via SSH",
				"web-02.example.com",
				"PASSED",
			},
		},
		{
			name: "multiple failed tests shown as bullets",
			results: &core.MultiHostResults{
				TotalDuration: 3 * time.Second,
				Hosts: []*core.HostResults{
					{
						Target:    "web-01.example.com",
						Connected: true,
						SpecResults: []*core.TestResults{
							{Results: []core.Result{
								{Name: "Package test", Status: core.StatusFail},
								{Name: "File test", Status: core.StatusFail},
								{Name: "Service test", Status: core.StatusError},
								{Name: "User test", Status: core.StatusPass},
							}},
						},
					},
				},
			},
			wantContains: []string{
				"Summary:",
				"0 passed, 1 failed",
				"web-01.example.com",
				"FAILED",
				"1 passed, 3 failed",
				"• Package test",
				"• File test",
				"• Service test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatMultiHostSummaryTable(tt.results)

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatMultiHostSummaryTable() output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}
