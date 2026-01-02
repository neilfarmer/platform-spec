package retry

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// Strategy represents the backoff strategy for retries
type Strategy string

const (
	// StrategyLinear increases delay linearly with each retry
	StrategyLinear Strategy = "linear"
	// StrategyExponential doubles delay with each retry
	StrategyExponential Strategy = "exponential"
	// StrategyJittered uses exponential backoff with random jitter
	StrategyJittered Strategy = "jittered"
)

// Config holds retry configuration
type Config struct {
	MaxRetries   int           // Maximum number of retry attempts (default: 3)
	InitialDelay time.Duration // Initial delay between retries (default: 1s)
	MaxDelay     time.Duration // Maximum delay between retries (default: 30s)
	Strategy     Strategy      // Backoff strategy (default: linear)
}

// DefaultConfig returns default retry configuration
func DefaultConfig() *Config {
	return &Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Strategy:     StrategyLinear,
	}
}

// CalculateDelay calculates the delay for a given retry attempt
func (c *Config) CalculateDelay(attempt int) time.Duration {
	var delay time.Duration

	// Cap attempt at 30 to prevent integer overflow in bitshift (2^30 is ~1 billion)
	// In practice, retry attempts should never exceed ~10, so this is safe
	if attempt > 30 {
		attempt = 30
	}

	switch c.Strategy {
	case StrategyLinear:
		// Linear: delay increases by initialDelay each time
		delay = c.InitialDelay * time.Duration(attempt+1)

	case StrategyExponential:
		// Exponential: delay doubles each time (2^attempt)
		// #nosec G115 -- attempt is explicitly capped at 30 above, uint conversion is safe
		delay = c.InitialDelay * time.Duration(1<<uint(attempt))

	case StrategyJittered:
		// Jittered exponential: exponential + random jitter (0-50% of delay)
		// #nosec G115 -- attempt is explicitly capped at 30 above, uint conversion is safe
		exponentialDelay := c.InitialDelay * time.Duration(1<<uint(attempt))

		// Use crypto/rand for jitter to satisfy security scanner
		maxJitter := int64(exponentialDelay / 2)
		if maxJitter > 0 {
			jitterBig, err := rand.Int(rand.Reader, big.NewInt(maxJitter))
			if err != nil {
				// Fallback to no jitter if random generation fails
				delay = exponentialDelay
			} else {
				jitter := time.Duration(jitterBig.Int64())
				delay = exponentialDelay + jitter
			}
		} else {
			delay = exponentialDelay
		}

	default:
		// Fallback to initial delay
		delay = c.InitialDelay
	}

	// Cap at max delay
	if delay > c.MaxDelay {
		delay = c.MaxDelay
	}

	return delay
}

// ErrorClassifier determines if an error is retryable
type ErrorClassifier func(error) bool

// Do executes a function with retry logic
func Do(ctx context.Context, config *Config, classifier ErrorClassifier, fn func() error) error {
	if config == nil {
		return fn()
	}

	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !classifier(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't wait after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := config.CalculateDelay(attempt)

		// Wait with context support
		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}
