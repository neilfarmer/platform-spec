package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Provider implements SSH-based testing
type Provider struct {
	client *ssh.Client
	config *Config
}

// Config holds SSH configuration
type Config struct {
	Host         string
	Port         int
	User         string
	IdentityFile string
	Timeout      time.Duration
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

	sshConfig := &ssh.ClientConfig{
		User:            p.config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Add proper host key verification
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
