package cli

import (
	"github.com/spf13/cobra"
)

// rootCmd is a root of all commands.
var rootCmd = &cobra.Command{
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
	rootCmd.AddCommand(mdsCmd)
	rootCmd.AddCommand(dsCmd)
	rootCmd.AddCommand(gwCmd)
}
