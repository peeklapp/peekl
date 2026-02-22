package main

import (
	"github.com/peeklapp/peekl/cmd/agent/commands"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "peekl-agent",
	Short: "The configuration management agent of the Peekl suite.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "/etc/peekl/config/agent.yml", "Path to the configuration file for the agent")
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.VersionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
