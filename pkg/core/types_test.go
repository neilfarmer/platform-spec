package core

import (
	"testing"
	"time"
)

func TestTestResults_Summary(t *testing.T) {
	tests := []struct {
		name    string
		results *TestResults
		want    struct {
			total   int
			passed  int
			failed  int
			skipped int
			errors  int
		}
	}{
		{
			name: "all passed",
			results: &TestResults{
				Results: []Result{
					{Status: StatusPass},
					{Status: StatusPass},
					{Status: StatusPass},
				},
			},
			want: struct {
				total   int
				passed  int
				failed  int
				skipped int
				errors  int
			}{3, 3, 0, 0, 0},
		},
		{
			name: "mixed results",
			results: &TestResults{
				Results: []Result{
					{Status: StatusPass},
					{Status: StatusFail},
					{Status: StatusSkip},
					{Status: StatusError},
				},
			},
			want: struct {
				total   int
				passed  int
				failed  int
				skipped int
				errors  int
			}{4, 1, 1, 1, 1},
		},
		{
			name:    "empty results",
			results: &TestResults{Results: []Result{}},
			want: struct {
				total   int
				passed  int
				failed  int
				skipped int
				errors  int
			}{0, 0, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, passed, failed, skipped, errors := tt.results.Summary()
			if total != tt.want.total {
				t.Errorf("total = %v, want %v", total, tt.want.total)
			}
			if passed != tt.want.passed {
				t.Errorf("passed = %v, want %v", passed, tt.want.passed)
			}
			if failed != tt.want.failed {
				t.Errorf("failed = %v, want %v", failed, tt.want.failed)
			}
			if skipped != tt.want.skipped {
				t.Errorf("skipped = %v, want %v", skipped, tt.want.skipped)
			}
			if errors != tt.want.errors {
				t.Errorf("errors = %v, want %v", errors, tt.want.errors)
			}
		})
	}
}

func TestTestResults_Success(t *testing.T) {
	tests := []struct {
		name    string
		results *TestResults
		want    bool
	}{
		{
			name: "all passed",
			results: &TestResults{
				Results: []Result{
					{Status: StatusPass},
					{Status: StatusPass},
				},
			},
			want: true,
		},
		{
			name: "has failures",
			results: &TestResults{
				Results: []Result{
					{Status: StatusPass},
					{Status: StatusFail},
				},
			},
			want: false,
		},
		{
			name: "has errors",
			results: &TestResults{
				Results: []Result{
					{Status: StatusPass},
					{Status: StatusError},
				},
			},
			want: false,
		},
		{
			name: "skipped is ok",
			results: &TestResults{
				Results: []Result{
					{Status: StatusPass},
					{Status: StatusSkip},
				},
			},
			want: true,
		},
		{
			name:    "empty is success",
			results: &TestResults{Results: []Result{}},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.results.Success(); got != tt.want {
				t.Errorf("Success() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_Duration(t *testing.T) {
	result := Result{
		Name:     "test",
		Status:   StatusPass,
		Duration: 150 * time.Millisecond,
	}

	if result.Duration != 150*time.Millisecond {
		t.Errorf("Duration = %v, want 150ms", result.Duration)
	}
}
