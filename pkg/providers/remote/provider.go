package remote

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	StrictHostKeyChecking  bool   // Enable strict host key checking (default: true)
	KnownHostsFile         string // Path to known_hosts file (default: ~/.ssh/known_hosts)
	InsecureIgnoreHostKey  bool   // Disable host key verification (INSECURE, not recommended)
	JumpHost               string // Jump host (bastion) hostname or IP
	JumpPort               int    // Jump host SSH port (default: 22)
	JumpUser               string // Jump host SSH user
	JumpIdentityFile       string // SSH private key for jump host (optional, defaults to IdentityFile)
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

// Connect establishes the SSH connection
// If a jump host is configured, it will connect through the jump host
func (p *Provider) Connect(ctx context.Context) error {
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

		client, err := p.connectViaJumpHost(jumpAuthMethods, targetAuthMethods, hostKeyCallback)
		if err != nil {
			return err
		}
		p.client = client
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

	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	p.client = client
	return nil
}

// buildAuthMethods creates SSH authentication methods for a given identity file
// If identityFile is empty, it falls back to SSH agent only
func (p *Provider) buildAuthMethods(identityFile string, hostType string) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	// Try SSH key file authentication if provided
	if identityFile != "" {
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
func (p *Provider) connectViaJumpHost(jumpAuthMethods, targetAuthMethods []ssh.AuthMethod, hostKeyCallback ssh.HostKeyCallback) (*ssh.Client, error) {
	// First, connect to the jump host using jump host auth methods
	jumpConfig := &ssh.ClientConfig{
		User:            p.config.JumpUser,
		Auth:            jumpAuthMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         p.config.Timeout,
	}

	jumpAddr := fmt.Sprintf("%s:%d", p.config.JumpHost, p.config.JumpPort)
	jumpClient, err := ssh.Dial("tcp", jumpAddr, jumpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to jump host %s: %w", jumpAddr, err)
	}

	// Store the jump client so it can be closed later
	p.jumpClient = jumpClient

	// Use the jump host connection to dial the target host
	targetAddr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	targetConn, err := jumpClient.Dial("tcp", targetAddr)
	if err != nil {
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
		targetConn.Close()
		jumpClient.Close()
		return nil, fmt.Errorf("failed to establish SSH connection to target %s: %w", targetAddr, err)
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

// ExecuteCommand executes a command via SSH and returns stdout, stderr, and exit code
func (p *Provider) ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	session, err := p.client.NewSession()
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to create session: %w", err)
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

	return stdout, stderr, exitCode, nil
}

// getHostKeyCallback returns the appropriate host key callback based on configuration
func (p *Provider) getHostKeyCallback() (ssh.HostKeyCallback, error) {
	// If explicitly set to insecure mode, use InsecureIgnoreHostKey
	// This is NOT recommended and should only be used in controlled environments
	if p.config.InsecureIgnoreHostKey {
		fmt.Fprintf(os.Stderr, "WARNING: SSH host key verification is disabled (insecure mode)\n")
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
