package retry

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"testing"
)

func TestIsRetryableSSHError_NetworkErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		want  bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "network timeout",
			err:  &testTimeoutError{},
			want: true,
		},
		{
			name: "connection refused",
			err:  &net.OpError{Err: syscall.ECONNREFUSED},
			want: true,
		},
		{
			name: "host unreachable",
			err:  &net.OpError{Err: syscall.EHOSTUNREACH},
			want: true,
		},
		{
			name: "network unreachable",
			err:  &net.OpError{Err: syscall.ENETUNREACH},
			want: true,
		},
		{
			name: "connection reset",
			err:  &net.OpError{Err: syscall.ECONNRESET},
			want: true,
		},
		{
			name: "connection timed out",
			err:  &net.OpError{Err: syscall.ETIMEDOUT},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableSSHError(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryableSSHError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRetryableSSHError_ConnectionErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		want     bool
	}{
		{
			name:   "connection reset by peer",
			errMsg: "connection reset by peer",
			want:   true,
		},
		{
			name:   "broken pipe",
			errMsg: "broken pipe",
			want:   true,
		},
		{
			name:   "i/o timeout",
			errMsg: "i/o timeout",
			want:   true,
		},
		{
			name:   "connection refused",
			errMsg: "dial tcp: connection refused",
			want:   true,
		},
		{
			name:   "no route to host",
			errMsg: "no route to host",
			want:   true,
		},
		{
			name:   "network is unreachable",
			errMsg: "network is unreachable",
			want:   true,
		},
		{
			name:   "connection timed out",
			errMsg: "dial tcp: connection timed out",
			want:   true,
		},
		{
			name:   "failed to create session",
			errMsg: "failed to create session",
			want:   true,
		},
		{
			name:   "administratively prohibited",
			errMsg: "ssh: rejected: administratively prohibited (open failed)",
			want:   true,
		},
		{
			name:   "EOF",
			errMsg: "EOF",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			got := IsRetryableSSHError(err)
			if got != tt.want {
				t.Errorf("IsRetryableSSHError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestIsNonRetryableSSHError_AuthErrors(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		want   bool
	}{
		{
			name:   "nil error",
			errMsg: "",
			want:   false,
		},
		{
			name:   "unable to authenticate",
			errMsg: "ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
			want:   true,
		},
		{
			name:   "permission denied",
			errMsg: "ssh: permission denied",
			want:   true,
		},
		{
			name:   "handshake failed",
			errMsg: "ssh: handshake failed: ssh: unable to authenticate",
			want:   true,
		},
		{
			name:   "authentication failed",
			errMsg: "authentication failed",
			want:   true,
		},
		{
			name:   "no supported methods remain",
			errMsg: "no supported methods remain",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = errors.New(tt.errMsg)
			}
			got := IsNonRetryableSSHError(err)
			if got != tt.want {
				t.Errorf("IsNonRetryableSSHError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestIsNonRetryableSSHError_ConfigErrors(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		want   bool
	}{
		{
			name:   "failed to read private key",
			errMsg: "failed to read private key: open /path/to/key: no such file or directory",
			want:   true,
		},
		{
			name:   "failed to parse private key",
			errMsg: "failed to parse private key",
			want:   true,
		},
		{
			name:   "no such file or directory",
			errMsg: "open ~/.ssh/id_rsa: no such file or directory",
			want:   true,
		},
		{
			name:   "known_hosts file not found",
			errMsg: "known_hosts file not found",
			want:   true,
		},
		{
			name:   "cannot decode encrypted private key",
			errMsg: "cannot decode encrypted private key without passphrase",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			got := IsNonRetryableSSHError(err)
			if got != tt.want {
				t.Errorf("IsNonRetryableSSHError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestIsNonRetryableSSHError_DNSErrors(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		want   bool
	}{
		{
			name:   "no such host",
			errMsg: "dial tcp: lookup invalid-host.example.com: no such host",
			want:   true,
		},
		{
			name:   "name or service not known",
			errMsg: "dial tcp: lookup invalid-host: name or service not known",
			want:   true,
		},
		{
			name:   "nodename nor servname provided",
			errMsg: "dial tcp: lookup invalid-host: nodename nor servname provided, or not known",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			got := IsNonRetryableSSHError(err)
			if got != tt.want {
				t.Errorf("IsNonRetryableSSHError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestIsNonRetryableSSHError_HostKeyErrors(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		want   bool
	}{
		{
			name:   "host key verification failed",
			errMsg: "ssh: host key verification failed",
			want:   true,
		},
		{
			name:   "knownhosts error",
			errMsg: "knownhosts: key mismatch",
			want:   true,
		},
		{
			name:   "key mismatch",
			errMsg: "remote host identification has changed: key mismatch",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			got := IsNonRetryableSSHError(err)
			if got != tt.want {
				t.Errorf("IsNonRetryableSSHError(%q) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

func TestErrorClassification_Precedence(t *testing.T) {
	// Test that certain errors are correctly identified as retryable
	// and don't accidentally match non-retryable patterns

	tests := []struct {
		name          string
		errMsg        string
		wantRetryable bool
		wantNonRetry  bool
	}{
		{
			name:          "connection refused is retryable",
			errMsg:        "dial tcp 192.168.1.1:22: connection refused",
			wantRetryable: true,
			wantNonRetry:  false,
		},
		{
			name:          "auth failure is non-retryable",
			errMsg:        "ssh: unable to authenticate",
			wantRetryable: false,
			wantNonRetry:  true,
		},
		{
			name:          "timeout is retryable",
			errMsg:        "i/o timeout",
			wantRetryable: true,
			wantNonRetry:  false,
		},
		{
			name:          "DNS failure is non-retryable",
			errMsg:        "no such host",
			wantRetryable: false,
			wantNonRetry:  true,
		},
		{
			name:          "unknown error is neither",
			errMsg:        "some unexpected error",
			wantRetryable: false,
			wantNonRetry:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)

			gotRetryable := IsRetryableSSHError(err)
			if gotRetryable != tt.wantRetryable {
				t.Errorf("IsRetryableSSHError(%q) = %v, want %v", tt.errMsg, gotRetryable, tt.wantRetryable)
			}

			gotNonRetry := IsNonRetryableSSHError(err)
			if gotNonRetry != tt.wantNonRetry {
				t.Errorf("IsNonRetryableSSHError(%q) = %v, want %v", tt.errMsg, gotNonRetry, tt.wantNonRetry)
			}
		})
	}
}

func TestErrorClassification_WrappedErrors(t *testing.T) {
	// Test that wrapped errors are correctly classified
	baseErr := &net.OpError{Err: syscall.ECONNREFUSED}
	wrappedErr := fmt.Errorf("failed to connect: %w", baseErr)

	if !IsRetryableSSHError(wrappedErr) {
		t.Error("Expected wrapped ECONNREFUSED to be retryable")
	}

	if IsNonRetryableSSHError(wrappedErr) {
		t.Error("Expected wrapped ECONNREFUSED to not be non-retryable")
	}
}

// testTimeoutError is a mock error that implements net.Error with Timeout() == true
type testTimeoutError struct{}

func (e *testTimeoutError) Error() string   { return "timeout error" }
func (e *testTimeoutError) Timeout() bool   { return true }
func (e *testTimeoutError) Temporary() bool { return true }
