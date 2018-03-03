package cmd

import (
	"log"

	"github.com/chanyoung/nil/pkg/ds"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var dsCfg config.Ds

var dsCmd = &cobra.Command{
	Use:   "ds",
	Short: "ds control commands",
	Long:  "ds control commands",
	Run:   dsRun,
}

func dsRun(cmd *cobra.Command, args []string) {
	d, err := ds.New(&dsCfg)
	if err != nil {
		log.Fatal(err)
	}

	d.Start()
}

func init() {
	dsCmd.Flags().StringVarP(&dsCfg.ServerAddr, "bind", "b", config.Get("ds.addr"), "address to which the ds will bind")
	dsCmd.Flags().StringVarP(&dsCfg.ServerPort, "port", "p", config.Get("ds.port"), "port on which the ds will listen")

	dsCmd.Flags().StringVarP(&dsCfg.Swim.CoordinatorAddr, "swim-coordinator-addr", "", config.Get("swim.coordinator_addr"), "swim coordinator address")

	dsCmd.Flags().StringVarP(&dsCfg.Security.CertsDir, "secure-certs-dir", "", config.Get("security.certs_dir"), "directory path of secure configuration files")
	dsCmd.Flags().StringVarP(&dsCfg.Security.RootCAPem, "secure-rootca-pem", "", config.Get("security.rootca_pem"), "file name of rootCA.pem")
	dsCmd.Flags().StringVarP(&dsCfg.Security.ServerKey, "secure-server-key", "", config.Get("security.server_key"), "file name of server key")
	dsCmd.Flags().StringVarP(&dsCfg.Security.ServerCrt, "secure-server-crt", "", config.Get("security.server_crt"), "file name of server crt")

	dsCmd.Flags().StringVarP(&dsCfg.LogLocation, "log", "l", config.Get("ds.log_location"), "log location of the ds will print out")
}
