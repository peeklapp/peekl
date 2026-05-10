package main

import (
	"os"

	"github.com/peeklapp/peekl/cmd/code/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "peekl-code",
	Short: "peekl-code is used to sync the code base of Peekl with distant.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose (debug) output")
	rootCmd.PersistentFlags().StringP("config", "c", "/etc/peekl/config/code.yml", "Configuration file to use for peekl-code")
	rootCmd.AddCommand(commands.CleanCmd)
	rootCmd.AddCommand(commands.SyncCmd)
	rootCmd.AddCommand(commands.DeleteCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
