package cli

import (
	"github.com/spf13/cobra"
)

var mdsUserCmd = &cobra.Command{
	Use:   "user",
	Short: "control user information",
	Long:  "control user information",
	Run:   mdsUserRun,
}

func mdsUserRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {
	mdsUserCmd.AddCommand(mdsUserAddCmd)
}
