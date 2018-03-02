package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd is a root of all commands.
var RootCmd = &cobra.Command{
	Use:   "nil [command] [flags]",
	Short: "nil command-line interface",
	Long:  `nil command-line interface`,
	Run:   rootCmdRun,
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {
	// Add commands.
	RootCmd.AddCommand(mdsCmd)
	RootCmd.AddCommand(osdCmd)
	RootCmd.AddCommand(gwCmd)
}
