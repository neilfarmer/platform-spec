package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "platform-spec",
	Short: "Infrastructure testing and verification tool",
	Long:  `A pluggable infrastructure testing framework that validates system state across multiple platforms.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(testCmd)
}
