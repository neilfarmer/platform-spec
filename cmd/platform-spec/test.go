package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
	k8splugin "github.com/neilfarmer/platform-spec/pkg/core/kubernetes"
	"github.com/neilfarmer/platform-spec/pkg/core/system"
	"github.com/neilfarmer/platform-spec/pkg/inventory"
	"github.com/neilfarmer/platform-spec/pkg/output"
	"github.com/neilfarmer/platform-spec/pkg/providers/kubernetes"
	"github.com/neilfarmer/platform-spec/pkg/providers/local"
	"github.com/neilfarmer/platform-spec/pkg/providers/remote"
	"github.com/spf13/cobra"
)

var (
	// Remote connection flags
	identityFile          string
	inventoryFile         string
	remotePort            int
	timeout               int
	strictHostKeyChecking bool
	knownHostsFile        string
	insecureIgnoreHostKey bool
	jumpHost              string
	jumpPort              int
	jumpUser              string
	jumpIdentityFile      string

	// Kubernetes flags
	kubeconfig    string
	kubeContext   string
	kubeNamespace string

	// Output flags
	outputFormat string
	verbose      bool
	noColor      bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests against infrastructure",
	Long:  `Run tests against various infrastructure providers (remote, local, kubernetes, etc.)`,
}

var remoteCmd = &cobra.Command{
	Use:   "remote [user@]host spec.yaml [spec2.yaml...] OR --inventory hosts.txt spec.yaml [spec2.yaml...]",
	Short: "Test remote systems via SSH",
	Long:  `Connect to remote systems via SSH and run tests defined in YAML spec files. Use --inventory to test multiple hosts.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   runRemoteTest,
}

var localCmd = &cobra.Command{
	Use:   "local spec.yaml [spec2.yaml...]",
	Short: "Test local system",
	Long:  `Run tests against the local system defined in YAML spec files.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   runLocalTest,
}

var awsCmd = &cobra.Command{
	Use:   "aws spec.yaml",
	Short: "Test AWS infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AWS provider will be implemented in a future release")
		os.Exit(0)
	},
}

var openstackCmd = &cobra.Command{
	Use:   "openstack spec.yaml",
	Short: "Test OpenStack infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("OpenStack provider will be implemented in a future release")
		os.Exit(0)
	},
}

var kubernetesCmd = &cobra.Command{
	Use:     "kubernetes spec.yaml [spec2.yaml...]",
	Aliases: []string{"k8s"},
	Short:   "Test Kubernetes resources",
	Long:    `Connect to a Kubernetes cluster and run tests defined in YAML spec files.`,
	Args:    cobra.MinimumNArgs(1),
	Run:     runKubernetesTest,
}

func init() {
	// Remote command flags
	remoteCmd.Flags().StringVarP(&identityFile, "identity", "i", "", "Path to SSH private key")
	remoteCmd.Flags().StringVarP(&inventoryFile, "inventory", "I", "", "Path to inventory file containing hosts to test")
	remoteCmd.Flags().IntVarP(&remotePort, "port", "p", 22, "SSH port")
	remoteCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Connection timeout in seconds")
	remoteCmd.Flags().BoolVar(&strictHostKeyChecking, "strict-host-key-checking", true, "Enable strict host key checking (default: true)")
	remoteCmd.Flags().StringVar(&knownHostsFile, "known-hosts-file", "", "Path to known_hosts file (default: ~/.ssh/known_hosts)")
	remoteCmd.Flags().BoolVar(&insecureIgnoreHostKey, "insecure-ignore-host-key", false, "Disable host key verification (INSECURE, not recommended)")
	remoteCmd.Flags().StringVarP(&jumpHost, "jump-host", "J", "", "Jump host (bastion) for SSH connection (format: [user@]host)")
	remoteCmd.Flags().IntVar(&jumpPort, "jump-port", 22, "Jump host SSH port (default: 22)")
	remoteCmd.Flags().StringVar(&jumpUser, "jump-user", "", "Jump host SSH user (overrides user from --jump-host)")
	remoteCmd.Flags().StringVar(&jumpIdentityFile, "jump-identity", "", "SSH private key for jump host (defaults to --identity if not specified)")

	// Output flags (shared across all test commands)
	remoteCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	remoteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	remoteCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Local command flags
	localCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	localCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	localCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Kubernetes command flags
	kubernetesCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: ~/.kube/config)")
	kubernetesCmd.Flags().StringVar(&kubeContext, "context", "", "Kubernetes context to use")
	kubernetesCmd.Flags().StringVar(&kubeNamespace, "namespace", "", "Default namespace for tests")
	kubernetesCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	kubernetesCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	kubernetesCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Add subcommands to test
	testCmd.AddCommand(remoteCmd)
	testCmd.AddCommand(localCmd)
	testCmd.AddCommand(awsCmd)
	testCmd.AddCommand(openstackCmd)
	testCmd.AddCommand(kubernetesCmd)
}

// testSingleHost tests a single host with the given specs
func testSingleHost(ctx context.Context, host, user string, specs []*core.Spec, config *remote.Config) (*core.HostResults, error) {
	startTime := time.Now()
	hostResults := &core.HostResults{
		Target:    fmt.Sprintf("%s@%s", user, host),
		Connected: false,
	}

	// Create remote provider
	remoteProvider := remote.NewProvider(config)

	// Connect to target
	if err := remoteProvider.Connect(ctx); err != nil {
		hostResults.ConnectionError = err
		hostResults.Duration = time.Since(startTime)
		return hostResults, err
	}
	defer remoteProvider.Close()

	hostResults.Connected = true

	if verbose {
		fmt.Printf("Connected to %s@%s\n\n", user, host)
	}

	// Execute tests for each spec file
	for _, spec := range specs {
		// Execute tests with plugins
		executor := core.NewExecutor(spec, remoteProvider, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())
		results, err := executor.Execute(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to execute tests: %w", err)
		}

		results.Target = fmt.Sprintf("%s@%s", user, host)
		hostResults.SpecResults = append(hostResults.SpecResults, results)
	}

	hostResults.Duration = time.Since(startTime)
	return hostResults, nil
}

func runRemoteTest(cmd *cobra.Command, args []string) {
	// Set color output preference
	output.NoColor = noColor

	// Determine mode and parse arguments
	var hosts []string
	var specFiles []string
	var defaultUser string

	if inventoryFile != "" {
		// Inventory mode: all args are spec files
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: at least one spec file required\n")
			os.Exit(1)
		}
		specFiles = args

		// Parse inventory file
		inv, err := inventory.ParseInventoryFile(inventoryFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse inventory file: %v\n", err)
			os.Exit(1)
		}
		hosts = inv.Hosts

		// In inventory mode, user comes from flags or defaults to root
		// Note: Inventory entries can optionally include user@ prefix
		defaultUser = "root"

		if verbose {
			fmt.Printf("Inventory mode: %d hosts from %s\n", len(hosts), inventoryFile)
			fmt.Printf("Spec files: %v\n", specFiles)
			fmt.Printf("\n")
		}
	} else {
		// Single-host mode: first arg is target, rest are spec files
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Error: target and at least one spec file required\n")
			fmt.Fprintf(os.Stderr, "Usage: %s\n", cmd.Use)
			os.Exit(1)
		}

		target := args[0]
		specFiles = args[1:]

		// Parse target to extract user and host
		user, host, err := remote.ParseTarget(target, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		hosts = []string{host}
		defaultUser = user

		if verbose {
			fmt.Printf("Target: %s\n", target)
			fmt.Printf("Spec files: %v\n", specFiles)
			fmt.Printf("\n")
		}
	}

	// Parse and validate spec files FIRST (fail fast)
	var specs []*core.Spec
	for _, specFile := range specFiles {
		spec, err := core.ParseSpec(specFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse spec %s: %v\n", specFile, err)
			os.Exit(1)
		}
		specs = append(specs, spec)
	}

	// Parse jump host if provided
	var parsedJumpHost, parsedJumpUser string
	var err error
	if jumpHost != "" {
		parsedJumpUser, parsedJumpHost, err = remote.ParseTarget(jumpHost, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing jump host: %v\n", err)
			os.Exit(1)
		}
		// Allow explicit jump user override
		if jumpUser != "" {
			parsedJumpUser = jumpUser
		}
		// Default jump identity to main identity if not specified (backwards compatibility)
		if jumpIdentityFile == "" {
			jumpIdentityFile = identityFile
		}
	}

	if verbose && (remotePort != 22 || identityFile != "" || jumpHost != "") {
		fmt.Printf("Port: %d\n", remotePort)
		if identityFile != "" {
			fmt.Printf("Identity: %s\n", identityFile)
		}
		if jumpHost != "" {
			fmt.Printf("Jump Host: %s\n", jumpHost)
			fmt.Printf("Jump Port: %d\n", jumpPort)
		}
		fmt.Printf("\n")
	}

	// Display security warning once if using insecure mode
	if insecureIgnoreHostKey && len(hosts) > 0 {
		if len(hosts) == 1 {
			fmt.Fprintf(os.Stderr, "WARNING: SSH host key verification is disabled (insecure mode)\n")
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: SSH host key verification is disabled for %d hosts (insecure mode)\n", len(hosts))
		}
	}

	// Test each host
	ctx := context.Background()
	multiResults := &core.MultiHostResults{
		Hosts: make([]*core.HostResults, 0, len(hosts)),
	}
	overallStart := time.Now()

	for _, hostEntry := range hosts {
		// Parse host entry - may be "host" or "user@host"
		var hostUser, hostname string
		parsedUser, parsedHost, err := remote.ParseTarget(hostEntry, defaultUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse host entry '%s': %v\n", hostEntry, err)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}
		hostUser = parsedUser
		hostname = parsedHost

		// Create config for this host
		config := &remote.Config{
			Host:                  hostname,
			Port:                  remotePort,
			User:                  hostUser,
			IdentityFile:          identityFile,
			Timeout:               time.Duration(timeout) * time.Second,
			StrictHostKeyChecking: strictHostKeyChecking,
			KnownHostsFile:        knownHostsFile,
			InsecureIgnoreHostKey: insecureIgnoreHostKey,
			JumpHost:              parsedJumpHost,
			JumpPort:              jumpPort,
			JumpUser:              parsedJumpUser,
			JumpIdentityFile:      jumpIdentityFile,
		}

		// Test this host
		hostResult, err := testSingleHost(ctx, hostname, hostUser, specs, config)
		if err != nil && hostResult == nil {
			// Unexpected error (not a connection error)
			fmt.Fprintf(os.Stderr, "Failed to test host %s: %v\n", hostname, err)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}

		multiResults.Hosts = append(multiResults.Hosts, hostResult)
	}

	multiResults.TotalDuration = time.Since(overallStart)

	// Output results
	if len(hosts) == 1 {
		// Single-host mode: use existing output format for backward compatibility
		hostResult := multiResults.Hosts[0]

		if !hostResult.Connected {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", hostResult.ConnectionError)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}

		for _, results := range hostResult.SpecResults {
			switch outputFormat {
			case "json":
				fmt.Println("JSON output not yet implemented")
			case "junit":
				fmt.Println("JUnit output not yet implemented")
			default:
				fmt.Print(output.FormatHuman(results))
			}
		}
	} else {
		// Multi-host mode: use multi-host output format
		switch outputFormat {
		case "json":
			fmt.Println("JSON output not yet implemented for multi-host")
		case "junit":
			fmt.Println("JUnit output not yet implemented for multi-host")
		default:
			fmt.Print(output.FormatMultiHostHuman(multiResults))
		}
	}

	// Exit with error code if any tests failed
	if !multiResults.Success() {
		os.Exit(1)
	}
}

func runLocalTest(cmd *cobra.Command, args []string) {
	specFiles := args

	// Set color output preference
	output.NoColor = noColor

	if verbose {
		fmt.Printf("Target: localhost\n")
		fmt.Printf("Spec files: %v\n", specFiles)
		fmt.Printf("\n")
	}

	// Create local provider
	localProvider := local.NewProvider()

	// Execute tests for each spec file
	ctx := context.Background()
	var allResults []*core.TestResults
	for _, specFile := range specFiles {
		// Parse spec
		spec, err := core.ParseSpec(specFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse spec %s: %v\n", specFile, err)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}

		// Execute tests with plugins
		executor := core.NewExecutor(spec, localProvider, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())
		results, err := executor.Execute(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute tests: %v\n", err)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}

		results.Target = "localhost"
		allResults = append(allResults, results)
	}

	// Output results
	for _, results := range allResults {
		switch outputFormat {
		case "json":
			fmt.Println("JSON output not yet implemented")
		case "junit":
			fmt.Println("JUnit output not yet implemented")
		default:
			fmt.Print(output.FormatHuman(results))
		}
	}

	// Exit with error code if any tests failed
	for _, results := range allResults {
		if !results.Success() {
			os.Exit(1)
		}
	}
}

func runKubernetesTest(cmd *cobra.Command, args []string) {
	specFiles := args

	// Set color output preference
	output.NoColor = noColor

	// Set default kubeconfig if not specified
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = filepath.Join(homeDir, ".kube", "config")
		}
	}

	if verbose {
		fmt.Printf("Target: Kubernetes cluster\n")
		if kubeconfig != "" {
			fmt.Printf("Kubeconfig: %s\n", kubeconfig)
		}
		if kubeContext != "" {
			fmt.Printf("Context: %s\n", kubeContext)
		}
		if kubeNamespace != "" {
			fmt.Printf("Default Namespace: %s\n", kubeNamespace)
		}
		fmt.Printf("Spec files: %v\n", specFiles)
		fmt.Printf("\n")
	}

	// Create Kubernetes provider
	k8sProvider := kubernetes.NewProvider(&kubernetes.Config{
		Kubeconfig: kubeconfig,
		Context:    kubeContext,
		Namespace:  kubeNamespace,
	})

	// Execute tests for each spec file
	ctx := context.Background()
	var allResults []*core.TestResults
	for _, specFile := range specFiles {
		// Parse spec
		spec, err := core.ParseSpec(specFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse spec %s: %v\n", specFile, err)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}

		// Override config namespace if flag provided
		if kubeNamespace != "" && spec.Config.KubernetesNamespace == "" {
			spec.Config.KubernetesNamespace = kubeNamespace
		}
		if kubeContext != "" && spec.Config.KubernetesContext == "" {
			spec.Config.KubernetesContext = kubeContext
		}

		// Execute tests with plugins
		executor := core.NewExecutor(spec, k8sProvider, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())
		results, err := executor.Execute(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute tests: %v\n", err)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}

		targetStr := "kubernetes"
		if kubeContext != "" {
			targetStr = fmt.Sprintf("kubernetes:%s", kubeContext)
		}
		results.Target = targetStr
		allResults = append(allResults, results)
	}

	// Output results
	for _, results := range allResults {
		switch outputFormat {
		case "json":
			fmt.Println("JSON output not yet implemented")
		case "junit":
			fmt.Println("JUnit output not yet implemented")
		default:
			fmt.Print(output.FormatHuman(results))
		}
	}

	// Exit with error code if any tests failed
	for _, results := range allResults {
		if !results.Success() {
			os.Exit(1)
		}
	}
}
