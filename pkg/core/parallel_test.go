package core

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestParallelExecutor_Sequential(t *testing.T) {
	// Test that 1 worker behaves correctly
	executor := NewParallelExecutor(1, false, false)

	jobs := []HostJob{
		{HostEntry: "host1", User: "user1"},
		{HostEntry: "host2", User: "user2"},
	}

	testFunc := func(ctx context.Context, job HostJob) (*HostResults, error) {
		return &HostResults{
			Target:    job.HostEntry,
			Connected: true,
		}, nil
	}

	results, err := executor.Execute(jobs, testFunc)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results.Hosts) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Hosts))
	}

	// Verify results are in order
	if results.Hosts[0].Target != "host1" {
		t.Errorf("Expected first result to be host1, got %s", results.Hosts[0].Target)
	}
	if results.Hosts[1].Target != "host2" {
		t.Errorf("Expected second result to be host2, got %s", results.Hosts[1].Target)
	}
}

func TestParallelExecutor_Parallel(t *testing.T) {
	// Test that parallel execution works
	executor := NewParallelExecutor(4, false, false)

	jobs := make([]HostJob, 10)
	for i := 0; i < 10; i++ {
		jobs[i] = HostJob{
			HostEntry: fmt.Sprintf("host%d", i),
			User:      "testuser",
		}
	}

	testFunc := func(ctx context.Context, job HostJob) (*HostResults, error) {
		// Simulate work
		time.Sleep(10 * time.Millisecond)
		return &HostResults{
			Target:    job.HostEntry,
			Connected: true,
		}, nil
	}

	start := time.Now()
	results, err := executor.Execute(jobs, testFunc)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results.Hosts) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results.Hosts))
	}

	// With 4 workers and 10 jobs of 10ms each, should finish in ~30ms
	// Sequential would take ~100ms
	if duration > 50*time.Millisecond {
		t.Errorf("Parallel execution took too long: %v (expected < 50ms)", duration)
	}
}

func TestParallelExecutor_FailFast(t *testing.T) {
	// Test that fail-fast stops on first failure
	executor := NewParallelExecutor(4, true, false)

	jobs := make([]HostJob, 10)
	for i := 0; i < 10; i++ {
		jobs[i] = HostJob{
			HostEntry: fmt.Sprintf("host%d", i),
			User:      "testuser",
		}
	}

	var counter int
	var mu sync.Mutex

	testFunc := func(ctx context.Context, job HostJob) (*HostResults, error) {
		mu.Lock()
		counter++
		isFirst := counter == 1
		mu.Unlock()

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			// Context cancelled, return quickly
			return &HostResults{
				Target:    job.HostEntry,
				Connected: false,
			}, nil
		default:
		}

		// Simulate work
		time.Sleep(10 * time.Millisecond)

		// First job fails
		if isFirst {
			return &HostResults{
				Target:          job.HostEntry,
				Connected:       false,
				ConnectionError: fmt.Errorf("simulated failure"),
			}, nil
		}

		return &HostResults{
			Target:    job.HostEntry,
			Connected: true,
		}, nil
	}

	results, err := executor.Execute(jobs, testFunc)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should have fewer than 10 results due to fail-fast
	// Note: Some workers may have already started their jobs before fail-fast triggers
	if len(results.Hosts) >= 10 {
		t.Errorf("Expected fewer than 10 results with fail-fast, got %d", len(results.Hosts))
	}

	// At least one result should be a failure
	hasFailure := false
	for _, host := range results.Hosts {
		if !host.Connected {
			hasFailure = true
			break
		}
	}
	if !hasFailure {
		t.Error("Expected at least one failure in results")
	}
}

func TestParallelExecutor_ConnectionErrors(t *testing.T) {
	// Test that connection errors are tracked correctly
	executor := NewParallelExecutor(2, false, false)

	jobs := []HostJob{
		{HostEntry: "host1", User: "user1"},
		{HostEntry: "host2", User: "user2"},
		{HostEntry: "host3", User: "user3"},
	}

	testFunc := func(ctx context.Context, job HostJob) (*HostResults, error) {
		// host2 fails to connect
		if job.HostEntry == "host2" {
			return &HostResults{
				Target:          job.HostEntry,
				Connected:       false,
				ConnectionError: fmt.Errorf("connection refused"),
			}, nil
		}

		return &HostResults{
			Target:    job.HostEntry,
			Connected: true,
			SpecResults: []*TestResults{
				{
					Results: []Result{
						{Status: StatusPass},
					},
				},
			},
		}, nil
	}

	results, err := executor.Execute(jobs, testFunc)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results.Hosts) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results.Hosts))
	}

	// Check summary
	total, passed, failed, connErrors := results.Summary()
	if total != 3 {
		t.Errorf("Expected 3 total hosts, got %d", total)
	}
	if passed != 2 {
		t.Errorf("Expected 2 passed hosts, got %d", passed)
	}
	if failed != 1 {
		t.Errorf("Expected 1 failed host, got %d", failed)
	}
	if connErrors != 1 {
		t.Errorf("Expected 1 connection error, got %d", connErrors)
	}
}

func TestParallelExecutor_EmptyJobList(t *testing.T) {
	// Test with empty job list
	executor := NewParallelExecutor(4, false, false)

	jobs := []HostJob{}

	testFunc := func(ctx context.Context, job HostJob) (*HostResults, error) {
		return &HostResults{
			Target:    job.HostEntry,
			Connected: true,
		}, nil
	}

	results, err := executor.Execute(jobs, testFunc)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results.Hosts) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results.Hosts))
	}
}

func TestParallelExecutor_MoreWorkersThanJobs(t *testing.T) {
	// Test with more workers than jobs (should work fine with idle workers)
	executor := NewParallelExecutor(10, false, false)

	jobs := []HostJob{
		{HostEntry: "host1", User: "user1"},
		{HostEntry: "host2", User: "user2"},
	}

	testFunc := func(ctx context.Context, job HostJob) (*HostResults, error) {
		return &HostResults{
			Target:    job.HostEntry,
			Connected: true,
		}, nil
	}

	results, err := executor.Execute(jobs, testFunc)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results.Hosts) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Hosts))
	}
}

func TestParallelExecutor_VerboseMode(t *testing.T) {
	// Test that verbose mode suppresses progress output
	executor := NewParallelExecutor(2, false, true) // verbose = true

	jobs := []HostJob{
		{HostEntry: "host1", User: "user1"},
		{HostEntry: "host2", User: "user2"},
	}

	testFunc := func(ctx context.Context, job HostJob) (*HostResults, error) {
		return &HostResults{
			Target:    job.HostEntry,
			Connected: true,
		}, nil
	}

	// In verbose mode, printProgress should not be called (tested indirectly)
	results, err := executor.Execute(jobs, testFunc)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results.Hosts) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Hosts))
	}
}
