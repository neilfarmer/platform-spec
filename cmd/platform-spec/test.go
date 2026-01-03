package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
	k8splugin "github.com/neilfarmer/platform-spec/pkg/core/kubernetes"
	"github.com/neilfarmer/platform-spec/pkg/core/system"
	"github.com/neilfarmer/platform-spec/pkg/inventory"
	"github.com/neilfarmer/platform-spec/pkg/output"
	"github.com/neilfarmer/platform-spec/pkg/providers/kubernetes"
	"github.com/neilfarmer/platform-spec/pkg/providers/local"
	"github.com/neilfarmer/platform-spec/pkg/providers/remote"
	"github.com/neilfarmer/platform-spec/pkg/retry"
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

	// Parallel execution flags
	parallel    string
	maxParallel int
	failFast    bool

	// Retry flags
	retries       int
	retryDelay    string
	retryBackoff  string
	retryMaxDelay string

	// Exit code control
	ignoreFailure bool
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

	// Retry flags
	remoteCmd.Flags().IntVar(&retries, "retries", 3, "Number of retry attempts for transient failures (0 = no retries)")
	remoteCmd.Flags().StringVar(&retryDelay, "retry-delay", "1s", "Initial delay between retry attempts (e.g., 1s, 500ms)")
	remoteCmd.Flags().StringVar(&retryBackoff, "retry-backoff", "linear", "Retry backoff strategy: linear, exponential, jittered")
	remoteCmd.Flags().StringVar(&retryMaxDelay, "retry-max-delay", "30s", "Maximum delay between retry attempts")

	// Output flags (shared across all test commands)
	remoteCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	remoteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	remoteCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	remoteCmd.Flags().BoolVar(&ignoreFailure, "ignore-failure", false, "Exit with code 0 even if tests fail")

	// Parallel execution flags
	remoteCmd.Flags().StringVar(&parallel, "parallel", "1", "Number of concurrent workers (integer or 'auto' for auto-detect)")
	remoteCmd.Flags().IntVar(&maxParallel, "max-parallel", 50, "Maximum number of concurrent workers (safety cap)")
	remoteCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop testing remaining hosts on first failure")

	// Local command flags
	localCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	localCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	localCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	localCmd.Flags().BoolVar(&ignoreFailure, "ignore-failure", false, "Exit with code 0 even if tests fail")

	// Kubernetes command flags
	kubernetesCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: ~/.kube/config)")
	kubernetesCmd.Flags().StringVar(&kubeContext, "context", "", "Kubernetes context to use")
	kubernetesCmd.Flags().StringVar(&kubeNamespace, "namespace", "", "Default namespace for tests")
	kubernetesCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	kubernetesCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	kubernetesCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	kubernetesCmd.Flags().BoolVar(&ignoreFailure, "ignore-failure", false, "Exit with code 0 even if tests fail")

	// Add subcommands to test
	testCmd.AddCommand(remoteCmd)
	testCmd.AddCommand(localCmd)
	testCmd.AddCommand(awsCmd)
	testCmd.AddCommand(openstackCmd)
	testCmd.AddCommand(kubernetesCmd)
}

// parseParallelFlag parses the --parallel flag value and returns the number of workers
func parseParallelFlag(parallelStr string, maxParallel int) (int, error) {
	if parallelStr == "auto" {
		// Auto-detect: Use runtime.NumCPU() with cap
		workers := runtime.NumCPU()
		if workers > maxParallel {
			workers = maxParallel
		}
		if workers < 1 {
			workers = 1
		}
		return workers, nil
	}

	// Parse as integer
	workers, err := strconv.Atoi(parallelStr)
	if err != nil {
		return 0, fmt.Errorf("invalid --parallel value: %s (must be integer or 'auto')", parallelStr)
	}

	if workers < 1 {
		return 0, fmt.Errorf("--parallel must be at least 1, got %d", workers)
	}

	// Apply max-parallel cap
	if workers > maxParallel {
		workers = maxParallel
	}

	return workers, nil
}

// testSingleHost tests a single host with the given specs
func testSingleHost(ctx context.Context, host, user string, specs []*core.Spec, config *remote.Config, isMultiHost bool) (*core.HostResults, error) {
	startTime := time.Now()
	target := fmt.Sprintf("%s@%s", user, host)
	hostResults := &core.HostResults{
		Target:    target,
		Connected: false,
	}

	// Create remote provider
	remoteProvider := remote.NewProvider(config)

	// Connect to target
	if err := remoteProvider.Connect(ctx); err != nil {
		hostResults.ConnectionError = err
		hostResults.Duration = time.Since(startTime)

		// Show connection error immediately
		if isMultiHost {
			fmt.Printf("%s: %s\n",
				output.ApplyColorExport(output.ColorBold, target),
				output.ApplyColorExport(output.ColorRed, "✗ Connection failed: "+err.Error()))
		}

		return hostResults, err
	}
	defer remoteProvider.Close()

	hostResults.Connected = true

	if verbose {
		fmt.Printf("Connected to %s\n", target)
	}

	// Execute tests for each spec file
	for _, spec := range specs {
		// Execute tests with plugins
		executor := core.NewExecutor(spec, remoteProvider, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())

		// Set callback for real-time streaming output with hostname prefix
		if isMultiHost {
			// For multi-host, prefix each result with hostname
			executor.SetResultCallback(func(result core.Result) {
				output.FormatSingleResultWithHost(result, target)
			})
		} else {
			// For single-host, no prefix needed
			executor.SetResultCallback(output.FormatSingleResult)
		}

		results, err := executor.Execute(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to execute tests: %w", err)
		}

		results.Target = target
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

	// Parse parallel flags
	workers, err := parseParallelFlag(parallel, maxParallel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if verbose && workers > 1 {
		fmt.Printf("Parallel execution: %d workers\n", workers)
		if failFast {
			fmt.Printf("Fail-fast: enabled\n")
		}
		fmt.Printf("\n")
	}

	// Parse jump host if provided
	var parsedJumpHost, parsedJumpUser string
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

	// Parse retry configuration
	var retryConfig *retry.Config
	if retries > 0 {
		initialDelay, err := time.ParseDuration(retryDelay)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --retry-delay: %v\n", err)
			os.Exit(1)
		}

		maxDelay, err := time.ParseDuration(retryMaxDelay)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --retry-max-delay: %v\n", err)
			os.Exit(1)
		}

		var strategy retry.Strategy
		switch retryBackoff {
		case "linear":
			strategy = retry.StrategyLinear
		case "exponential":
			strategy = retry.StrategyExponential
		case "jittered":
			strategy = retry.StrategyJittered
		default:
			fmt.Fprintf(os.Stderr, "Invalid --retry-backoff: %s (must be linear, exponential, or jittered)\n", retryBackoff)
			os.Exit(1)
		}

		retryConfig = &retry.Config{
			MaxRetries:   retries,
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
			Strategy:     strategy,
		}
	}

	// Display security warning once if using insecure mode
	if insecureIgnoreHostKey && len(hosts) > 0 {
		if len(hosts) == 1 {
			fmt.Fprintf(os.Stderr, "WARNING: SSH host key verification is disabled (insecure mode)\n")
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: SSH host key verification is disabled for %d hosts (insecure mode)\n", len(hosts))
		}
	}

	// Build job list for all hosts
	var jobs []core.HostJob
	for _, hostEntry := range hosts {
		// Parse host entry - may be "host" or "user@host"
		parsedUser, parsedHost, err := remote.ParseTarget(hostEntry, defaultUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse host entry '%s': %v\n", hostEntry, err)
			fmt.Print(output.PrintFailed())
			os.Exit(1)
		}

		// Create config for this host
		config := &remote.Config{
			Host:                  parsedHost,
			Port:                  remotePort,
			User:                  parsedUser,
			IdentityFile:          identityFile,
			Timeout:               time.Duration(timeout) * time.Second,
			StrictHostKeyChecking: strictHostKeyChecking,
			KnownHostsFile:        knownHostsFile,
			InsecureIgnoreHostKey: insecureIgnoreHostKey,
			JumpHost:              parsedJumpHost,
			JumpPort:              jumpPort,
			JumpUser:              parsedJumpUser,
			JumpIdentityFile:      jumpIdentityFile,
			RetryConfig:           retryConfig,
		}

		jobs = append(jobs, core.HostJob{
			HostEntry: hostEntry,
			User:      parsedUser,
			Config:    config,
		})
	}

	// Determine if we're in multi-host mode
	isMultiHost := len(hosts) > 1

	// Create test function wrapper
	testFunc := func(ctx context.Context, job core.HostJob) (*core.HostResults, error) {
		config := job.Config.(*remote.Config)
		return testSingleHost(ctx, config.Host, config.User, specs, config, isMultiHost)
	}

	// Execute tests (sequential or parallel based on workers)
	ctx := context.Background()
	var multiResults *core.MultiHostResults
	overallStart := time.Now()

	if workers == 1 {
		// Sequential execution (backward compatible)
		multiResults = &core.MultiHostResults{
			Hosts: make([]*core.HostResults, 0, len(jobs)),
		}

		for _, job := range jobs {
			hostResult, err := testFunc(ctx, job)
			if err != nil && hostResult == nil {
				// Unexpected error (not a connection error)
				fmt.Fprintf(os.Stderr, "Failed to test host: %v\n", err)
				fmt.Print(output.PrintFailed())
				os.Exit(1)
			}

			multiResults.Hosts = append(multiResults.Hosts, hostResult)

			// Check fail-fast in sequential mode
			if failFast && !hostResult.Success() {
				if verbose {
					fmt.Fprintf(os.Stderr, "Fail-fast: stopping due to failure on %s\n", hostResult.Target)
				}
				break
			}
		}
	} else {
		// Parallel execution
		executor := core.NewParallelExecutor(workers, failFast, verbose)
		multiResults, err = executor.Execute(jobs, testFunc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parallel execution failed: %v\n", err)
			os.Exit(1)
		}

		// Clear progress line before showing results
		if !verbose {
			output.ClearProgressLine()
			fmt.Fprintf(os.Stderr, "\n")
		}
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
		// Multi-host mode: just show final summary table (results already streamed)
		switch outputFormat {
		case "json":
			fmt.Println("JSON output not yet implemented for multi-host")
		case "junit":
			fmt.Println("JUnit output not yet implemented for multi-host")
		default:
			fmt.Print(output.FormatMultiHostSummaryTable(multiResults))
		}
	}

	// Exit with error code if any tests failed (unless --ignore-failure is set)
	if !multiResults.Success() && !ignoreFailure {
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

		// Set callback for real-time streaming output
		executor.SetResultCallback(output.FormatSingleResult)

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

	// Exit with error code if any tests failed (unless --ignore-failure is set)
	if !ignoreFailure {
		for _, results := range allResults {
			if !results.Success() {
				os.Exit(1)
			}
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

		// Set callback for real-time streaming output
		executor.SetResultCallback(output.FormatSingleResult)

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

	// Exit with error code if any tests failed (unless --ignore-failure is set)
	if !ignoreFailure {
		for _, results := range allResults {
			if !results.Success() {
				os.Exit(1)
			}
		}
	}
}
