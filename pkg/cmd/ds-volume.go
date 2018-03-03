package cmd

import (
	"github.com/spf13/cobra"
)

var dsVolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "control ds volumes",
	Long:  "control ds volumes",
	Run:   dsVolumeRun,
}

func dsVolumeRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {
	dsVolumeCmd.AddCommand(dsVolumeAddCmd)
}
