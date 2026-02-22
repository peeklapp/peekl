package main

import (
	"github.com/peeklapp/peekl/cmd/server/commands"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "peekl-server",
	Short: "peekl-server is the server component of the Peekl suite.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "/etc/peekl/config/server.yml", "Path to the configuration file for the server")
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.VersionCmd)
	rootCmd.AddCommand(commands.CaCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
