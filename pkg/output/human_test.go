package output

import (
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
				"Platform-Spec Test Results",
				"Spec: Test Suite",
				"Target: ubuntu@host",
				"✓ Test 1",
				"✓ Test 2",
				"2 passed, 0 failed",
				"PASSED",
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
				"✓ Test 1",
				"✗ Test 2",
				"Test failed",
				"○ Test 3",
				"1 passed, 1 failed, 1 skipped",
				"FAILED",
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
				"⚠ Test 1",
				"Error occurred",
				"0 passed, 0 failed, 0 skipped, 1 errors",
				"FAILED",
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
