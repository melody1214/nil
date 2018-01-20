package cmd

import (
	"log"

	"github.com/chanyoung/nil/pkg/osd"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var osdCfg config.Osd

var osdCmd = &cobra.Command{
	Use:   "osd",
	Short: "osd control commands",
	Long:  "osd control commands",
	Run:   osdRun,
}

func osdRun(cmd *cobra.Command, args []string) {
	o, err := osd.New(&osdCfg)
	if err != nil {
		log.Fatal(err)
	}

	o.Start()
}

func init() {
	osdCmd.Flags().StringVarP(&osdCfg.ServerAddr, "bind", "b", config.Get("osd.addr"), "address to which the osd will bind")
	osdCmd.Flags().StringVarP(&osdCfg.ServerPort, "port", "p", config.Get("osd.port"), "port on which the osd will listen")

	osdCmd.Flags().StringVarP(&osdCfg.Security.CertsDir, "secure-certs-dir", "", config.Get("security.certs_dir"), "directory path of secure configuration files")
	osdCmd.Flags().StringVarP(&osdCfg.Security.RootCAPem, "secure-rootca-pem", "", config.Get("security.rootca_pem"), "file name of rootCA.pem")
	osdCmd.Flags().StringVarP(&osdCfg.Security.ServerKey, "secure-server-key", "", config.Get("security.server_key"), "file name of server key")
	osdCmd.Flags().StringVarP(&osdCfg.Security.ServerCrt, "secure-server-crt", "", config.Get("security.server_crt"), "file name of server crt")

	osdCmd.Flags().StringVarP(&osdCfg.LogLocation, "log", "l", config.Get("osd.log_location"), "log location of the osd will print out")
}
