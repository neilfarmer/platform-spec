package core

import "time"

// Status represents the result status of a test
type Status string

const (
	StatusPass  Status = "passed"
	StatusFail  Status = "failed"
	StatusSkip  Status = "skipped"
	StatusError Status = "error"
)

// Result represents the result of a single test
type Result struct {
	Name     string
	Status   Status
	Message  string
	Duration time.Duration
	Details  map[string]interface{}
}

// TestResults represents the aggregated results of all tests
type TestResults struct {
	SpecName  string
	Target    string
	StartTime time.Time
	Duration  time.Duration
	Results   []Result
}

// Summary returns a summary of the test results
func (tr *TestResults) Summary() (total, passed, failed, skipped, errors int) {
	total = len(tr.Results)
	for _, r := range tr.Results {
		switch r.Status {
		case StatusPass:
			passed++
		case StatusFail:
			failed++
		case StatusSkip:
			skipped++
		case StatusError:
			errors++
		}
	}
	return
}

// Success returns true if all tests passed
func (tr *TestResults) Success() bool {
	_, _, failed, _, errors := tr.Summary()
	return failed == 0 && errors == 0
}
