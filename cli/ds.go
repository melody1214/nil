package cli

import (
	"log"
	"os"

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
	if err := os.Chdir(dsCfg.WorkDir); err != nil {
		log.Fatal(err)
	}

	d, err := ds.New(&dsCfg)
	if err != nil {
		log.Fatal(err)
	}

	d.Start()
}

func init() {
	dsCmd.AddCommand(dsVolumeCmd)

	dsCmd.Flags().StringVarP(&dsCfg.ServerAddr, "bind", "b", config.Get("ds.addr"), "address to which the ds will bind")
	dsCmd.Flags().StringVarP(&dsCfg.ServerPort, "port", "p", config.Get("ds.port"), "port on which the ds will listen")

	dsCmd.Flags().StringVarP(&dsCfg.Swim.CoordinatorAddr, "swim-coordinator-addr", "", config.Get("swim.coordinator_addr"), "swim coordinator address")
	dsCmd.Flags().StringVarP(&dsCfg.Swim.Period, "swim-period", "", config.Get("swim.period"), "swim ping period time")
	dsCmd.Flags().StringVarP(&dsCfg.Swim.Expire, "swim-expire", "", config.Get("swim.expire"), "swim ping expire time")

	dsCmd.Flags().StringVarP(&dsCfg.WorkDir, "work-dir", "", config.Get("ds.work_dir"), "working directory")

	dsCmd.Flags().StringVarP(&dsCfg.Store, "store", "", config.Get("ds.store"), "type of backend store")

	dsCmd.Flags().StringVarP(&dsCfg.Security.CertsDir, "secure-certs-dir", "", config.Get("security.certs_dir"), "directory path of secure configuration files")
	dsCmd.Flags().StringVarP(&dsCfg.Security.RootCAPem, "secure-rootca-pem", "", config.Get("security.rootca_pem"), "file name of rootCA.pem")
	dsCmd.Flags().StringVarP(&dsCfg.Security.ServerKey, "secure-server-key", "", config.Get("security.server_key"), "file name of server key")
	dsCmd.Flags().StringVarP(&dsCfg.Security.ServerCrt, "secure-server-crt", "", config.Get("security.server_crt"), "file name of server crt")

	dsCmd.Flags().StringVarP(&dsCfg.LogLocation, "log", "l", config.Get("ds.log_location"), "log location of the ds will print out")
}
