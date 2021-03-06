package cli

import (
	"log"
	"os"

	"github.com/chanyoung/nil/app/gw"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var (
	gwCfg config.Gw
)

var gwCmd = &cobra.Command{
	Use:   "gw",
	Short: "gateway control commands",
	Long:  "gateway control commands",
	Run:   gwRun,
}

func gwRun(cmd *cobra.Command, args []string) {
	if err := os.Chdir(gwCfg.WorkDir); err != nil {
		log.Fatal(err)
	}

	if err := gw.Bootstrap(gwCfg); err != nil {
		log.Fatal(err)
	}
}

func init() {
	gwCmd.Flags().StringVarP(&gwCfg.ServerAddr, "bind", "b", config.Get("gw.addr"), "address to which the gateway will bind")
	gwCmd.Flags().StringVarP(&gwCfg.ServerPort, "port", "p", config.Get("gw.port"), "port on which the gateway will listen")
	gwCmd.Flags().StringVarP(&gwCfg.LogLocation, "log", "l", config.Get("gw.log_location"), "log location of the gateway will print out")
	gwCmd.Flags().StringVarP(&gwCfg.FirstMds, "first-mds", "", config.Get("gw.first_mds"), "mds address to get local cluster information in initialize routine")

	gwCmd.Flags().StringVarP(&gwCfg.WorkDir, "work-dir", "", config.Get("gw.work_dir"), "working directory")

	gwCmd.Flags().StringVarP(&gwCfg.Security.CertsDir, "secure-certs-dir", "", config.Get("security.certs_dir"), "directory path of secure configuration files")
	gwCmd.Flags().StringVarP(&gwCfg.Security.RootCAPem, "secure-rootca-pem", "", config.Get("security.rootca_pem"), "file name of rootCA.pem")
	gwCmd.Flags().StringVarP(&gwCfg.Security.ServerKey, "secure-server-key", "", config.Get("security.server_key"), "file name of server key")
	gwCmd.Flags().StringVarP(&gwCfg.Security.ServerCrt, "secure-server-crt", "", config.Get("security.server_crt"), "file name of server crt")
}
