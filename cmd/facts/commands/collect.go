package commands

import (
	"encoding/json"
	"fmt"

	"github.com/peeklapp/peekl/pkg/facts"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(collectCmd)
}

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect facts about your system",
	Run: func(cmd *cobra.Command, args []string) {
		facter := facts.NewFacter()
		facts, err := facter.Collect()
		if err != nil {
			logrus.Fatal(
				fmt.Sprintf("An error happened while gathering facts : %s", err.Error()),
			)
		}
		factsJson, err := json.Marshal(facts)
		if err != nil {
			logrus.Fatal(
				fmt.Sprintf("An error happened while gathering facts : %s", err.Error()),
			)
		}
		fmt.Println(string(factsJson))
	},
}
