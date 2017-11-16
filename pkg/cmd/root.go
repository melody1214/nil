package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RootCmd is a root of all commands.
var RootCmd = &cobra.Command{
	Use:   "nil [command] [flags]",
	Short: "nil command-line interface",
	Long:  `nil command-line interface`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cli work!!")
	},
}

func init() {
	// Add commands.
	RootCmd.AddCommand(mdsCmd)
	RootCmd.AddCommand(osdCmd)
	RootCmd.AddCommand(mapCmd)
	RootCmd.AddCommand(gwCmd)
}
