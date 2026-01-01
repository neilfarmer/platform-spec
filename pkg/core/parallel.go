package core

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// HostJob represents a single host to be tested
type HostJob struct {
	HostEntry string
	User      string
	Config    interface{} // Provider-specific config (e.g., *remote.Config)
}

// ProgressTracker tracks test progress
type ProgressTracker struct {
	mu             sync.Mutex
	totalHosts     int
	completedHosts int
	failedHosts    int
	connErrors     int
}

// ParallelExecutor manages concurrent host testing
type ParallelExecutor struct {
	workers    int
	failFast   bool
	verbose    bool
	jobChan    chan HostJob
	resultChan chan *HostResults
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	results    []*HostResults
	progress   ProgressTracker
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(workers int, failFast bool, verbose bool) *ParallelExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &ParallelExecutor{
		workers:    workers,
		failFast:   failFast,
		verbose:    verbose,
		jobChan:    make(chan HostJob, workers*2), // Buffer for smoother flow
		resultChan: make(chan *HostResults, workers),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Execute runs tests across multiple hosts in parallel
func (pe *ParallelExecutor) Execute(jobs []HostJob, testFunc func(context.Context, HostJob) (*HostResults, error)) (*MultiHostResults, error) {
	startTime := time.Now()
	pe.progress.totalHosts = len(jobs)

	// Start result collector goroutine
	collectorDone := make(chan struct{})
	go pe.collectResults(collectorDone)

	// Start worker pool
	pe.wg.Add(pe.workers)
	for i := 0; i < pe.workers; i++ {
		go pe.worker(i, testFunc)
	}

	// Feed jobs to workers
	for _, job := range jobs {
		select {
		case pe.jobChan <- job:
		case <-pe.ctx.Done():
			// Fail-fast triggered, stop sending jobs
			break
		}
	}
	close(pe.jobChan)

	// Wait for all workers to finish
	pe.wg.Wait()
	close(pe.resultChan)

	// Wait for result collector to finish
	<-collectorDone

	return &MultiHostResults{
		Hosts:         pe.results,
		TotalDuration: time.Since(startTime),
	}, nil
}

// worker processes jobs from the job channel
func (pe *ParallelExecutor) worker(id int, testFunc func(context.Context, HostJob) (*HostResults, error)) {
	defer pe.wg.Done()

	for job := range pe.jobChan {
		select {
		case <-pe.ctx.Done():
			// Fail-fast triggered, stop processing
			return
		default:
			// Execute test for this host
			result, _ := testFunc(pe.ctx, job)

			// Send result to collector
			select {
			case pe.resultChan <- result:
			case <-pe.ctx.Done():
				return
			}

			// Check fail-fast condition
			if pe.failFast && !result.Success() {
				pe.cancel() // Signal all workers to stop
				return
			}
		}
	}
}

// collectResults aggregates results from workers
func (pe *ParallelExecutor) collectResults(done chan struct{}) {
	defer close(done)

	for result := range pe.resultChan {
		pe.mu.Lock()
		pe.results = append(pe.results, result)
		pe.progress.completedHosts++

		if !result.Connected {
			pe.progress.connErrors++
			pe.progress.failedHosts++
		} else if !result.Success() {
			pe.progress.failedHosts++
		}

		// Print progress (if not in verbose mode)
		if !pe.verbose {
			pe.printProgress()
		}
		pe.mu.Unlock()
	}
}

// printProgress displays current progress using carriage return
func (pe *ParallelExecutor) printProgress() {
	fmt.Fprintf(os.Stderr, "\rTesting hosts: %d/%d completed (%d passed, %d failed, %d conn errors)",
		pe.progress.completedHosts,
		pe.progress.totalHosts,
		pe.progress.completedHosts-pe.progress.failedHosts,
		pe.progress.failedHosts,
		pe.progress.connErrors,
	)
}
