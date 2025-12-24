package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
	"github.com/neilfarmer/platform-spec/pkg/output"
	"github.com/neilfarmer/platform-spec/pkg/providers/local"
	"github.com/neilfarmer/platform-spec/pkg/providers/ssh"
	"github.com/spf13/cobra"
)

var (
	// SSH flags
	identityFile string
	sshPort      int
	timeout      int

	// Output flags
	outputFormat string
	verbose      bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests against infrastructure",
	Long:  `Run tests against various infrastructure providers (ssh, aws, openstack, etc.)`,
}

var sshCmd = &cobra.Command{
	Use:   "ssh [user@]host spec.yaml [spec2.yaml...]",
	Short: "Test infrastructure via SSH",
	Long:  `Connect to a host via SSH and run tests defined in YAML spec files.`,
	Args:  cobra.MinimumNArgs(2),
	Run:   runSSHTest,
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

func init() {
	// SSH command flags
	sshCmd.Flags().StringVarP(&identityFile, "identity", "i", "", "Path to SSH private key")
	sshCmd.Flags().IntVarP(&sshPort, "port", "p", 22, "SSH port")
	sshCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Connection timeout in seconds")

	// Output flags (shared across all test commands)
	sshCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	sshCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Local command flags
	localCmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, junit)")
	localCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Add subcommands to test
	testCmd.AddCommand(sshCmd)
	testCmd.AddCommand(localCmd)
	testCmd.AddCommand(awsCmd)
	testCmd.AddCommand(openstackCmd)
}

func runSSHTest(cmd *cobra.Command, args []string) {
	target := args[0]
	specFiles := args[1:]

	if verbose {
		fmt.Printf("Target: %s\n", target)
		fmt.Printf("Port: %d\n", sshPort)
		if identityFile != "" {
			fmt.Printf("Identity: %s\n", identityFile)
		}
		fmt.Printf("Spec files: %v\n", specFiles)
		fmt.Printf("\n")
	}

	// Parse target
	user, host, err := ssh.ParseTarget(target, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create SSH provider
	sshProvider := ssh.NewProvider(&ssh.Config{
		Host:         host,
		Port:         sshPort,
		User:         user,
		IdentityFile: identityFile,
		Timeout:      time.Duration(timeout) * time.Second,
	})

	// Connect to target
	ctx := context.Background()
	if err := sshProvider.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer sshProvider.Close()

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

		// Execute tests
		executor := core.NewExecutor(spec, sshProvider)
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

		// Execute tests
		executor := core.NewExecutor(spec, localProvider)
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
