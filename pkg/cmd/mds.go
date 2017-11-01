package cmd

import (
	"log"

	"github.com/chanyoung/nil/pkg/mds"
	"github.com/spf13/cobra"
)

var mdsCmd = &cobra.Command{
	Use:   "mds",
	Short: "mds control commands",
	Long:  "mds control commands",
	Run:   mdsRun,
}

func mdsRun(cmd *cobra.Command, args []string) {
	m, err := mds.New()
	if err != nil {
		log.Fatal(err)
	}

	m.Start()
}
