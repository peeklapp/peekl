package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Obtain information about version and compilation",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("peekl-agent version 0.1.0")
	},
}
