package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Obtain information about version and compilation",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Name: peekl-agent\nVersion: %s\nCommit: %s\nDate: %s\n", version, commit, date)
	},
}
