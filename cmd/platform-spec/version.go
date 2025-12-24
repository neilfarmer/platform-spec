package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("platform-spec version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
