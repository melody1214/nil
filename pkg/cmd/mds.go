package cmd

import (
	"log"

	"github.com/chanyoung/nil/pkg/mds"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var (
	mdscfg config.Mds
)

var mdsCmd = &cobra.Command{
	Use:   "mds",
	Short: "mds control commands",
	Long:  "mds control commands",
	Run:   mdsRun,
}

func mdsRun(cmd *cobra.Command, args []string) {
	m, err := mds.New(&mdscfg)
	if err != nil {
		log.Fatal(err)
	}

	m.Start()
}

func init() {
	mdsCmd.Flags().StringVarP(&mdscfg.ServerAddr, "bind", "b", config.Get("mds.addr"), "address to which the mds will bind")
	mdsCmd.Flags().StringVarP(&mdscfg.ServerPort, "port", "p", config.Get("mds.port"), "port on which the mds will listen")
	mdsCmd.Flags().StringVarP(&mdscfg.LogLocation, "log", "l", config.Get("mds.log_location"), "log location of the mds will print out")
}
