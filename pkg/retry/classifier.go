package retry

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

// IsRetryableSSHError determines if an SSH error should be retried
// Retryable errors are typically transient network or connection issues
func IsRetryableSSHError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check for network-level errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Network timeout errors are retryable
		if netErr.Timeout() {
			return true
		}
	}

	// Check for syscall errors wrapped in net.OpError
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Err == syscall.ECONNREFUSED ||
			opErr.Err == syscall.EHOSTUNREACH ||
			opErr.Err == syscall.ENETUNREACH ||
			opErr.Err == syscall.ECONNRESET ||
			opErr.Err == syscall.ETIMEDOUT {
			return true
		}
	}

	// SSH-specific retryable errors based on error messages
	retryablePatterns := []string{
		"connection reset by peer",
		"broken pipe",
		"i/o timeout",
		"connection refused",
		"no route to host",
		"network is unreachable",
		"connection timed out",
		"failed to create session",
		"ssh: rejected: administratively prohibited",
		"EOF", // Can indicate connection drop
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// IsNonRetryableSSHError determines if an SSH error should NOT be retried
// Non-retryable errors are permanent configuration or authentication issues
func IsNonRetryableSSHError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Authentication failures - permanent errors
	authPatterns := []string{
		"unable to authenticate",
		"no supported methods remain",
		"permission denied",
		"ssh: handshake failed",
		"authentication failed",
	}

	for _, pattern := range authPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Host key verification failures
	hostKeyPatterns := []string{
		"host key verification failed",
		"knownhosts:",
		"key mismatch",
	}

	for _, pattern := range hostKeyPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// File and configuration errors
	configPatterns := []string{
		"failed to read private key",
		"failed to parse private key",
		"no such file or directory",
		"known_hosts file not found",
		"cannot decode encrypted private key",
	}

	for _, pattern := range configPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// DNS resolution failures (permanent for this run)
	dnsPatterns := []string{
		"no such host",
		"name or service not known",
		"nodename nor servname provided",
	}

	for _, pattern := range dnsPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
