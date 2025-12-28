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
	"github.com/neilfarmer/platform-spec/pkg/output"
	"github.com/neilfarmer/platform-spec/pkg/providers/kubernetes"
	"github.com/neilfarmer/platform-spec/pkg/providers/local"
	"github.com/neilfarmer/platform-spec/pkg/providers/remote"
	"github.com/spf13/cobra"
)

var (
	// Remote connection flags
	identityFile          string
	remotePort            int
	timeout               int
	strictHostKeyChecking bool
	knownHostsFile        string
	insecureIgnoreHostKey bool

	// Kubernetes flags
	kubeconfig    string
	kubeContext   string
	kubeNamespace string

	// Output flags
	outputFormat string
	verbose      bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests against infrastructure",
	Long:  `Run tests against various infrastructure providers (remote, local, kubernetes, etc.)`,
}

var remoteCmd = &cobra.Command{
	Use:   "remote [user@]host spec.yaml [spec2.yaml...]",
	Short: "Test remote systems via SSH",
	Long:  `Connect to a remote system via SSH and run tests defined in YAML spec files.`,
	Args:  cobra.MinimumNArgs(2),
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
	remoteCmd.Flags().IntVarP(&remotePort, "port", "p", 22, "SSH port")
	remoteCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Connection timeout in seconds")
	remoteCmd.Flags().BoolVar(&strictHostKeyChecking, "strict-host-key-checking", true, "Enable strict host key checking (default: true)")
	remoteCmd.Flags().StringVar(&knownHostsFile, "known-hosts-file", "", "Path to known_hosts file (default: ~/.ssh/known_hosts)")
	remoteCmd.Flags().BoolVar(&insecureIgnoreHostKey, "insecure-ignore-host-key", false, "Disable host key verification (INSECURE, not recommended)")

	// Output flags (shared across all test commands)
	remoteCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	remoteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Local command flags
	localCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	localCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Kubernetes command flags
	kubernetesCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: ~/.kube/config)")
	kubernetesCmd.Flags().StringVar(&kubeContext, "context", "", "Kubernetes context to use")
	kubernetesCmd.Flags().StringVar(&kubeNamespace, "namespace", "", "Default namespace for tests")
	kubernetesCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	kubernetesCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Add subcommands to test
	testCmd.AddCommand(remoteCmd)
	testCmd.AddCommand(localCmd)
	testCmd.AddCommand(awsCmd)
	testCmd.AddCommand(openstackCmd)
	testCmd.AddCommand(kubernetesCmd)
}

func runRemoteTest(cmd *cobra.Command, args []string) {
	target := args[0]
	specFiles := args[1:]

	if verbose {
		fmt.Printf("Target: %s\n", target)
		fmt.Printf("Port: %d\n", remotePort)
		if identityFile != "" {
			fmt.Printf("Identity: %s\n", identityFile)
		}
		fmt.Printf("Spec files: %v\n", specFiles)
		fmt.Printf("\n")
	}

	// Parse target
	user, host, err := remote.ParseTarget(target, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create remote provider
	remoteProvider := remote.NewProvider(&remote.Config{
		Host:                  host,
		Port:                  remotePort,
		User:                  user,
		IdentityFile:          identityFile,
		Timeout:               time.Duration(timeout) * time.Second,
		StrictHostKeyChecking: strictHostKeyChecking,
		KnownHostsFile:        knownHostsFile,
		InsecureIgnoreHostKey: insecureIgnoreHostKey,
	})

	// Connect to target
	ctx := context.Background()
	if err := remoteProvider.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer remoteProvider.Close()

	if verbose {
		fmt.Printf("Connected to %s@%s\n\n", user, host)
	}

	// Execute tests for each spec file
	var allResults []*core.TestResults
	for _, specFile := range specFiles {
		// Parse spec
		spec, err := core.ParseSpec(specFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse spec %s: %v\n", specFile, err)
			os.Exit(1)
		}

		// Execute tests with plugins
		executor := core.NewExecutor(spec, remoteProvider, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())
		results, err := executor.Execute(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute tests: %v\n", err)
			os.Exit(1)
		}

		results.Target = fmt.Sprintf("%s@%s", user, host)
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

func runLocalTest(cmd *cobra.Command, args []string) {
	specFiles := args

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
			os.Exit(1)
		}

		// Execute tests with plugins
		executor := core.NewExecutor(spec, localProvider, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())
		results, err := executor.Execute(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute tests: %v\n", err)
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
