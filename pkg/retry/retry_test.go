package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCalculateDelay_Linear(t *testing.T) {
	config := &Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Strategy:     StrategyLinear,
	}

	tests := []struct {
		attempt      int
		expectedDelay time.Duration
	}{
		{0, 1 * time.Second},  // 1s * (0+1) = 1s
		{1, 2 * time.Second},  // 1s * (1+1) = 2s
		{2, 3 * time.Second},  // 1s * (2+1) = 3s
		{3, 4 * time.Second},  // 1s * (3+1) = 4s
	}

	for _, tt := range tests {
		delay := config.CalculateDelay(tt.attempt)
		if delay != tt.expectedDelay {
			t.Errorf("Linear backoff attempt %d: expected %v, got %v", tt.attempt, tt.expectedDelay, delay)
		}
	}
}

func TestCalculateDelay_Exponential(t *testing.T) {
	config := &Config{
		MaxRetries:   4,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Strategy:     StrategyExponential,
	}

	tests := []struct {
		attempt      int
		expectedDelay time.Duration
	}{
		{0, 1 * time.Second},   // 1s * 2^0 = 1s
		{1, 2 * time.Second},   // 1s * 2^1 = 2s
		{2, 4 * time.Second},   // 1s * 2^2 = 4s
		{3, 8 * time.Second},   // 1s * 2^3 = 8s
		{4, 16 * time.Second},  // 1s * 2^4 = 16s
	}

	for _, tt := range tests {
		delay := config.CalculateDelay(tt.attempt)
		if delay != tt.expectedDelay {
			t.Errorf("Exponential backoff attempt %d: expected %v, got %v", tt.attempt, tt.expectedDelay, delay)
		}
	}
}

func TestCalculateDelay_Jittered(t *testing.T) {
	config := &Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Strategy:     StrategyJittered,
	}

	// For jittered backoff, we can't test exact values due to randomness
	// But we can verify the range
	for attempt := 0; attempt < 4; attempt++ {
		delay := config.CalculateDelay(attempt)

		// Calculate expected range
		exponentialDelay := config.InitialDelay * time.Duration(1<<uint(attempt))
		minDelay := exponentialDelay
		maxDelay := exponentialDelay + (exponentialDelay / 2)

		if delay < minDelay || delay > maxDelay {
			t.Errorf("Jittered backoff attempt %d: delay %v outside expected range [%v, %v]",
				attempt, delay, minDelay, maxDelay)
		}
	}
}

func TestCalculateDelay_MaxCap(t *testing.T) {
	config := &Config{
		MaxRetries:   10,
		InitialDelay: 1 * time.Second,
		MaxDelay:     5 * time.Second,
		Strategy:     StrategyExponential,
	}

	// With exponential backoff, later attempts should be capped at maxDelay
	delay := config.CalculateDelay(10) // 1s * 2^10 = 1024s, but capped at 5s
	if delay != 5*time.Second {
		t.Errorf("MaxDelay cap failed: expected 5s, got %v", delay)
	}
}

func TestDo_SuccessFirstAttempt(t *testing.T) {
	config := DefaultConfig()
	ctx := context.Background()

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := Do(ctx, config, func(e error) bool { return true }, fn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	config := &Config{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Strategy:     StrategyLinear,
	}
	ctx := context.Background()

	callCount := 0
	retryableErr := errors.New("temporary error")

	fn := func() error {
		callCount++
		if callCount < 3 {
			return retryableErr
		}
		return nil
	}

	start := time.Now()
	err := Do(ctx, config, func(e error) bool { return true }, fn)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}

	// Should have waited at least 2 delays (10ms + 20ms = 30ms)
	if duration < 30*time.Millisecond {
		t.Errorf("Expected at least 30ms delay, got %v", duration)
	}
}

func TestDo_MaxRetriesExceeded(t *testing.T) {
	config := &Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Strategy:     StrategyLinear,
	}
	ctx := context.Background()

	callCount := 0
	retryableErr := errors.New("persistent error")

	fn := func() error {
		callCount++
		return retryableErr
	}

	err := Do(ctx, config, func(e error) bool { return true }, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, retryableErr) {
		t.Errorf("Expected wrapped retryableErr, got %v", err)
	}

	// Should call fn MaxRetries+1 times (initial + 3 retries = 4 total)
	if callCount != 4 {
		t.Errorf("Expected 4 calls (1 initial + 3 retries), got %d", callCount)
	}
}

func TestDo_NonRetryableError(t *testing.T) {
	config := DefaultConfig()
	ctx := context.Background()

	callCount := 0
	nonRetryableErr := errors.New("auth error")

	fn := func() error {
		callCount++
		return nonRetryableErr
	}

	// Classifier returns false for non-retryable errors
	classifier := func(e error) bool {
		return false
	}

	err := Do(ctx, config, classifier, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Should only be called once (no retries)
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	config := &Config{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Strategy:     StrategyLinear,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	callCount := 0
	retryableErr := errors.New("temporary error")

	fn := func() error {
		callCount++
		return retryableErr
	}

	err := Do(ctx, config, func(e error) bool { return true }, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded wrapped error, got %v", err)
	}

	// Should be called at least once but not all retries due to timeout
	if callCount == 0 {
		t.Error("Expected at least 1 call")
	}
	if callCount > 3 {
		t.Errorf("Expected fewer calls due to timeout, got %d", callCount)
	}
}

func TestDo_NilConfig(t *testing.T) {
	ctx := context.Background()

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	// Nil config should execute function directly without retries
	err := Do(ctx, nil, func(e error) bool { return true }, fn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.InitialDelay != 1*time.Second {
		t.Errorf("Expected InitialDelay=1s, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay=30s, got %v", config.MaxDelay)
	}

	if config.Strategy != StrategyLinear {
		t.Errorf("Expected Strategy=linear, got %v", config.Strategy)
	}
}
