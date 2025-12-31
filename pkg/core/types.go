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

// HostResults represents the results of testing a single host
type HostResults struct {
	Target          string         // Resolved target (user@host)
	Connected       bool           // Did SSH connection succeed?
	ConnectionError error          // Connection error if any
	SpecResults     []*TestResults // Results for each spec file
	Duration        time.Duration  // Total time for this host
}

// Success returns true if the host connected and all tests passed
func (hr *HostResults) Success() bool {
	if !hr.Connected {
		return false
	}
	for _, spec := range hr.SpecResults {
		if !spec.Success() {
			return false
		}
	}
	return true
}

// MultiHostResults represents the results of testing multiple hosts
type MultiHostResults struct {
	Hosts         []*HostResults
	TotalDuration time.Duration
}

// Success returns true if all hosts connected and all tests passed
func (mhr *MultiHostResults) Success() bool {
	for _, host := range mhr.Hosts {
		if !host.Success() {
			return false
		}
	}
	return true
}

// Summary returns counts: totalHosts, passedHosts, failedHosts, connectionErrors
func (mhr *MultiHostResults) Summary() (totalHosts, passedHosts, failedHosts, connectionErrors int) {
	totalHosts = len(mhr.Hosts)
	for _, host := range mhr.Hosts {
		if !host.Connected {
			connectionErrors++
			failedHosts++
		} else if host.Success() {
			passedHosts++
		} else {
			failedHosts++
		}
	}
	return
}
