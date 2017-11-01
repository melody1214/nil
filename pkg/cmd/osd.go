package cmd

import (
	"log"

	"github.com/chanyoung/nil/pkg/osd"
	"github.com/spf13/cobra"
)

var osdCmd = &cobra.Command{
	Use:   "osd",
	Short: "osd control commands",
	Long:  "osd control commands",
	Run:   osdRun,
}

func osdRun(cmd *cobra.Command, args []string) {
	o, err := osd.New()
	if err != nil {
		log.Fatal(err)
	}

	o.Start()
}
