package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	buildTime = "unknown" // set by Makefile
	goversion = runtime.Version()
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print build information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Build: %s\nGo: %s\n", buildTime, goversion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
