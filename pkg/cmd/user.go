package cmd

import (
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "control user information",
	Long:  "control user information",
	Run:   userRun,
}

func userRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {
	userCmd.AddCommand(userAddCmd)
}
