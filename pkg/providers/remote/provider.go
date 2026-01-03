package remote

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/retry"
	ssh_config "github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Provider implements remote system testing via SSH
type Provider struct {
	client     *ssh.Client
	jumpClient *ssh.Client // Jump host client (if using jump host)
	config     *Config
}

// Config holds remote connection configuration
type Config struct {
	Host                   string
	Port                   int
	User                   string
	IdentityFile           string
	Timeout                time.Duration
	StrictHostKeyChecking  bool          // Enable strict host key checking (default: true)
	KnownHostsFile         string        // Path to known_hosts file (default: ~/.ssh/known_hosts)
	InsecureIgnoreHostKey  bool          // Disable host key verification (INSECURE, not recommended)
	JumpHost               string        // Jump host (bastion) hostname or IP
	JumpPort               int           // Jump host SSH port (default: 22)
	JumpUser               string        // Jump host SSH user
	JumpIdentityFile       string        // SSH private key for jump host (optional, defaults to IdentityFile)
	RetryConfig            *retry.Config // Retry configuration (nil = no retries)
	Verbose                bool          // Enable verbose output showing timing and commands
}

// ParseTarget parses a target string like "user@host" or "host"
func ParseTarget(target string, defaultUser string) (user, host string, err error) {
	parts := strings.Split(target, "@")
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else if len(parts) == 1 {
		user = defaultUser
		if user == "" {
			user = "root"
		}
		return user, parts[0], nil
	}
	return "", "", fmt.Errorf("invalid target format: %s", target)
}

// NewProvider creates a new SSH provider
func NewProvider(config *Config) *Provider {
	return &Provider{
		config: config,
	}
}

// Connect establishes the SSH connection with optional retry logic
// If a jump host is configured, it will connect through the jump host
func (p *Provider) Connect(ctx context.Context) error {
	// If retry config is nil, execute directly without retries
	if p.config.RetryConfig == nil {
		return p.connectOnce(ctx)
	}

	// Wrap connection logic with retry
	return retry.Do(ctx, p.config.RetryConfig, retry.IsRetryableSSHError, func() error {
		return p.connectOnce(ctx)
	})
}

// connectOnce performs a single connection attempt without retry logic
func (p *Provider) connectOnce(ctx context.Context) error {
	startTime := time.Now()
	if p.config.Verbose {
		target := fmt.Sprintf("%s@%s:%d", p.config.User, p.config.Host, p.config.Port)
		if p.config.JumpHost != "" {
			fmt.Printf("[%s] Connecting to %s via jump host %s\n", formatElapsed(0), target, p.config.JumpHost)
		} else {
			fmt.Printf("[%s] Connecting to %s\n", formatElapsed(0), target)
		}
	}

	// Configure host key verification
	hostKeyCallback, err := p.getHostKeyCallback()
	if err != nil {
		return fmt.Errorf("failed to configure host key verification: %w", err)
	}

	// If jump host is configured, connect through it with separate auth
	if p.config.JumpHost != "" {
		// Build auth methods for jump host
		jumpAuthMethods, err := p.buildAuthMethods(p.config.JumpIdentityFile, "jump host")
		if err != nil {
			return err
		}

		// Build auth methods for target host
		targetAuthMethods, err := p.buildAuthMethods(p.config.IdentityFile, "target host")
		if err != nil {
			return err
		}

		client, err := p.connectViaJumpHost(jumpAuthMethods, targetAuthMethods, hostKeyCallback, startTime)
		if err != nil {
			return err
		}
		p.client = client
		if p.config.Verbose {
			fmt.Printf("[%s] Connected and authenticated (%.2fs)\n", formatElapsed(time.Since(startTime)), time.Since(startTime).Seconds())
		}
		return nil
	}

	// Direct connection (no jump host) - use target auth methods
	targetAuthMethods, err := p.buildAuthMethods(p.config.IdentityFile, "target host")
	if err != nil {
		return err
	}

	sshConfig := &ssh.ClientConfig{
		User:            p.config.User,
		Auth:            targetAuthMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         p.config.Timeout,
	}

	// Resolve hostname via SSH config before DNS resolution
	resolvedHost := resolveHostFromSSHConfig(p.config.Host)
	addr := fmt.Sprintf("%s:%d", resolvedHost, p.config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	p.client = client
	if p.config.Verbose {
		fmt.Printf("[%s] Connected and authenticated (%.2fs)\n", formatElapsed(time.Since(startTime)), time.Since(startTime).Seconds())
	}
	return nil
}

// buildAuthMethods creates SSH authentication methods for a given identity file
// If identityFile is empty, it falls back to SSH agent only
func (p *Provider) buildAuthMethods(identityFile string, hostType string) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	// Try SSH key file authentication if provided
	if identityFile != "" {
		// #nosec G304 -- Reading user-specified SSH key file is intentional and required functionality.
		// The user controls the path via CLI flag, similar to ssh -i flag behavior.
		key, err := os.ReadFile(identityFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key for %s: %w", hostType, err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key for %s: %w", hostType, err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Try SSH agent authentication as fallback
	if sshAgent, err := getSSHAgent(); err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(sshAgent.Signers))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method available for %s (no key file provided and no SSH agent found)", hostType)
	}

	return authMethods, nil
}

// connectViaJumpHost establishes an SSH connection through a jump host
// jumpAuthMethods: authentication for the jump host
// targetAuthMethods: authentication for the target host (can be different)
func (p *Provider) connectViaJumpHost(jumpAuthMethods, targetAuthMethods []ssh.AuthMethod, hostKeyCallback ssh.HostKeyCallback, startTime time.Time) (*ssh.Client, error) {
	// First, connect to the jump host using jump host auth methods
	jumpConfig := &ssh.ClientConfig{
		User:            p.config.JumpUser,
		Auth:            jumpAuthMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         p.config.Timeout,
	}

	// Resolve jump host hostname via SSH config before DNS resolution
	resolvedJumpHost := resolveHostFromSSHConfig(p.config.JumpHost)
	jumpAddr := fmt.Sprintf("%s:%d", resolvedJumpHost, p.config.JumpPort)
	jumpClient, err := ssh.Dial("tcp", jumpAddr, jumpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to jump host %s: %w", jumpAddr, err)
	}

	if p.config.Verbose {
		fmt.Printf("[%s] Connected to jump host (%.2fs)\n", formatElapsed(time.Since(startTime)), time.Since(startTime).Seconds())
	}

	// Store the jump client so it can be closed later
	p.jumpClient = jumpClient

	// Resolve target host hostname via SSH config before DNS resolution
	resolvedTargetHost := resolveHostFromSSHConfig(p.config.Host)
	targetAddr := fmt.Sprintf("%s:%d", resolvedTargetHost, p.config.Port)
	targetConn, err := jumpClient.Dial("tcp", targetAddr)
	if err != nil {
		// #nosec G104 -- Already in error path, ignoring close error is acceptable
		jumpClient.Close()
		return nil, fmt.Errorf("failed to dial target %s through jump host: %w", targetAddr, err)
	}

	// Create SSH connection to target using target host auth methods
	targetConfig := &ssh.ClientConfig{
		User:            p.config.User,
		Auth:            targetAuthMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         p.config.Timeout,
	}

	ncc, chans, reqs, err := ssh.NewClientConn(targetConn, targetAddr, targetConfig)
	if err != nil {
		// #nosec G104 -- Already in error path, cleanup errors can be safely ignored
		targetConn.Close()
		// #nosec G104 -- Already in error path, cleanup errors can be safely ignored
		jumpClient.Close()
		return nil, fmt.Errorf("failed to establish SSH connection to target %s: %w", targetAddr, err)
	}

	if p.config.Verbose {
		fmt.Printf("[%s] Connected to target through jump host (%.2fs)\n", formatElapsed(time.Since(startTime)), time.Since(startTime).Seconds())
	}

	// Create the target client
	targetClient := ssh.NewClient(ncc, chans, reqs)

	return targetClient, nil
}

// Close closes the SSH connection(s)
// If using a jump host, both the target and jump host connections are closed
func (p *Provider) Close() error {
	var err error

	// Close target client first
	if p.client != nil {
		if closeErr := p.client.Close(); closeErr != nil {
			err = closeErr
		}
	}

	// Close jump host client if it exists
	if p.jumpClient != nil {
		if closeErr := p.jumpClient.Close(); closeErr != nil {
			// Return the first error, but try to close both
			if err == nil {
				err = closeErr
			}
		}
	}

	return err
}

// ExecuteCommand executes a command via SSH with optional retry logic
func (p *Provider) ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	// If retry config is nil, execute directly without retries
	if p.config.RetryConfig == nil {
		return p.executeCommandOnce(ctx, command)
	}

	// Wrap execution with retry logic
	var stdoutResult, stderrResult string
	var exitCodeResult int

	retryErr := retry.Do(ctx, p.config.RetryConfig, retry.IsRetryableSSHError, func() error {
		var execErr error
		stdoutResult, stderrResult, exitCodeResult, execErr = p.executeCommandOnce(ctx, command)
		return execErr
	})

	return stdoutResult, stderrResult, exitCodeResult, retryErr
}

// executeCommandOnce performs a single command execution attempt with automatic reconnection
func (p *Provider) executeCommandOnce(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	startTime := time.Now()
	if p.config.Verbose {
		fmt.Printf("[%s] Executing: %s\n", formatElapsed(0), command)
	}

	session, err := p.client.NewSession()
	if err != nil {
		// Connection might be dead - try to reconnect once
		if reconnectErr := p.connectOnce(ctx); reconnectErr != nil {
			return "", "", -1, fmt.Errorf("failed to reconnect after session error: %w", reconnectErr)
		}
		// Retry session creation after reconnect
		session, err = p.client.NewSession()
		if err != nil {
			return "", "", -1, fmt.Errorf("failed to create session after reconnect: %w", err)
		}
	}
	defer session.Close()

	var stdoutBuf, stderrBuf strings.Builder
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(command)
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
			err = nil
		} else {
			return stdout, stderr, -1, fmt.Errorf("command execution failed: %w", err)
		}
	} else {
		exitCode = 0
	}

	if p.config.Verbose {
		elapsed := time.Since(startTime)
		fmt.Printf("[%s] Command completed (%.2fs, exit code: %d)\n", formatElapsed(elapsed), elapsed.Seconds(), exitCode)
	}

	return stdout, stderr, exitCode, nil
}

// getHostKeyCallback returns the appropriate host key callback based on configuration
func (p *Provider) getHostKeyCallback() (ssh.HostKeyCallback, error) {
	// If explicitly set to insecure mode, use InsecureIgnoreHostKey
	// This is NOT recommended and should only be used in controlled environments
	// Note: Warning is displayed once at the CLI level, not per-host
	if p.config.InsecureIgnoreHostKey {
		// #nosec G106 -- Insecure mode is explicitly opt-in via CLI flag with warning
		return ssh.InsecureIgnoreHostKey(), nil
	}

	// Determine known_hosts file path
	knownHostsPath := p.config.KnownHostsFile
	if knownHostsPath == "" {
		// Default to ~/.ssh/known_hosts
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		knownHostsPath = filepath.Join(home, ".ssh", "known_hosts")
	}

	// Check if known_hosts file exists
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		// If StrictHostKeyChecking is enabled (default), return error
		// This matches OpenSSH behavior
		if p.config.StrictHostKeyChecking {
			return nil, fmt.Errorf("known_hosts file not found at %s (strict host key checking enabled). "+
				"Either create the file, disable strict checking, or use insecure mode (not recommended)", knownHostsPath)
		}
		// If not strict, warn and use insecure mode
		fmt.Fprintf(os.Stderr, "WARNING: known_hosts file not found at %s, disabling host key verification\n", knownHostsPath)
		// #nosec G106 -- Fallback when known_hosts missing and strict checking disabled
		return ssh.InsecureIgnoreHostKey(), nil
	}

	// Use known_hosts for host key verification
	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create host key callback from %s: %w", knownHostsPath, err)
	}

	return hostKeyCallback, nil
}

// getSSHAgent connects to the SSH agent
func getSSHAgent() (agent.Agent, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	return agent.NewClient(conn), nil
}

// resolveHostFromSSHConfig resolves a hostname using SSH config files
// It checks ~/.ssh/config and /etc/ssh/ssh_config for Host patterns
// and returns the HostName directive value if found, or the original hostname if not
func resolveHostFromSSHConfig(host string) string {
	// Try to read SSH config files in order of precedence:
	// 1. User config: ~/.ssh/config
	// 2. System config: /etc/ssh/ssh_config

	// Try user config first
	home, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(home, ".ssh", "config")
		if hostname := getHostnameFromConfig(userConfigPath, host); hostname != "" {
			return hostname
		}
	}

	// Try system config
	systemConfigPath := "/etc/ssh/ssh_config"
	if hostname := getHostnameFromConfig(systemConfigPath, host); hostname != "" {
		return hostname
	}

	// No HostName found in config, return original host
	return host
}

// getHostnameFromConfig reads an SSH config file and returns the HostName for the given host
func getHostnameFromConfig(configPath string, host string) string {
	// #nosec G304 -- Reading SSH config files is intentional and required functionality
	f, err := os.Open(configPath)
	if err != nil {
		// Config file doesn't exist or can't be read, which is fine
		return ""
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		// Can't parse config, return empty
		return ""
	}

	// Get the HostName directive for this host
	// ssh_config.Config.Get returns empty string if not found
	hostname, _ := cfg.Get(host, "HostName")
	return hostname
}

// formatElapsed formats a duration as MM:SS.SS for verbose output
func formatElapsed(d time.Duration) string {
	seconds := d.Seconds()
	minutes := int(seconds / 60)
	seconds = seconds - float64(minutes*60)
	return fmt.Sprintf("%02d:%05.2f", minutes, seconds)
}
