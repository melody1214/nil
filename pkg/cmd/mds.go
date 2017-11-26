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
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLUser, "mysql-user", "", config.Get("mds.mysql_user"), "user id to mysql server")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLPassword, "mysql-password", "", config.Get("mds.mysql_password"), "password of mysql user")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLHost, "mysql-host", "", config.Get("mds.mysql_host"), "host address of mysql server")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLPort, "mysql-port", "", config.Get("mds.mysql_port"), "port number of mysql server")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLDatabase, "mysql-database", "", config.Get("mds.mysql_database"), "mysql schema name")
	mdsCmd.Flags().StringVarP(&mdscfg.LogLocation, "log", "l", config.Get("mds.log_location"), "log location of the mds will print out")
}
