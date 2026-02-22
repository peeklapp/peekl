package main

import (
	"fmt"
	"os"

	"github.com/peeklapp/peekl/cmd/facts/commands"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(commands.CollectCmd)
}

var rootCmd = &cobra.Command{
	Use:   "peekl-facts",
	Short: "peekl-facts is an utility that allows you to collect facts about a node.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
