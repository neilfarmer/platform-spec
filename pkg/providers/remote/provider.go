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
	client *ssh.Client
	config *Config
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
func (p *Provider) Connect(ctx context.Context) error {
	authMethods := []ssh.AuthMethod{}

	// Try SSH key file authentication if provided
	if p.config.IdentityFile != "" {
		key, err := os.ReadFile(p.config.IdentityFile)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Try SSH agent authentication
	if sshAgent, err := getSSHAgent(); err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(sshAgent.Signers))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication method available (no key file provided and no SSH agent found)")
	}

	// Configure host key verification
	hostKeyCallback, err := p.getHostKeyCallback()
	if err != nil {
		return fmt.Errorf("failed to configure host key verification: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            p.config.User,
		Auth:            authMethods,
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

// Close closes the SSH connection
func (p *Provider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
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
