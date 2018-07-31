package cli

import (
	"log"
	"os"

	"github.com/chanyoung/nil/app/mds"
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
	if err := os.Chdir(mdscfg.WorkDir); err != nil {
		log.Fatal(err)
	}

	if err := mds.Bootstrap(mdscfg); err != nil {
		log.Fatal(err)
	}
}

func init() {
	mdsCmd.AddCommand(mdsMapCmd)
	mdsCmd.AddCommand(mdsUserCmd)
	mdsCmd.AddCommand(mdsJobCmd)
	mdsCmd.AddCommand(mdsGGGCmd)

	mdsCmd.Flags().StringVarP(&mdscfg.ServerAddr, "bind", "b", config.Get("mds.addr"), "address to which the mds will bind")
	mdsCmd.Flags().StringVarP(&mdscfg.ServerPort, "port", "p", config.Get("mds.port"), "port on which the mds will listen")

	mdsCmd.Flags().StringVarP(&mdscfg.WorkDir, "work-dir", "", config.Get("mds.work_dir"), "working directory")

	mdsCmd.Flags().StringVarP(&mdscfg.MySQLUser, "mysql-user", "", config.Get("mds.mysql_user"), "user id to mysql server")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLPassword, "mysql-password", "", config.Get("mds.mysql_password"), "password of mysql user")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLHost, "mysql-host", "", config.Get("mds.mysql_host"), "host address of mysql server")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLPort, "mysql-port", "", config.Get("mds.mysql_port"), "port number of mysql server")
	mdsCmd.Flags().StringVarP(&mdscfg.MySQLDatabase, "mysql-database", "", config.Get("mds.mysql_database"), "mysql schema name")

	mdsCmd.Flags().StringVarP(&mdscfg.Rebalance, "rebalance", "", config.Get("mds.rebalance"), "peroid for checking balance of the cluster")

	mdsCmd.Flags().StringVarP(&mdscfg.Raft.LocalClusterAddr, "raft-local-cluster-addr", "", config.Get("raft.local_cluster_addr"), "raft local cluster end point")
	mdsCmd.Flags().StringVarP(&mdscfg.Raft.LocalClusterRegion, "raft-local-cluster-region", "", config.Get("raft.local_cluster_region"), "region name of the local cluster")
	mdsCmd.Flags().StringVarP(&mdscfg.Raft.GlobalClusterAddr, "raft-global-cluster-addr", "", config.Get("raft.global_cluster_addr"), "global raft cluster end point")
	mdsCmd.Flags().StringVarP(&mdscfg.Raft.ClusterJoin, "raft-cluster-join", "", config.Get("raft.cluster_join"), "join an existing raft cluster")
	mdsCmd.Flags().StringVarP(&mdscfg.Raft.RaftDir, "raft-dir", "", config.Get("raft.raft_dir"), "directory path of raft log store")
	mdsCmd.Flags().StringVarP(&mdscfg.Raft.ElectionTimeout, "raft-election-timeout", "", config.Get("raft.election_timeout"), "raft election timeout")

	mdsCmd.Flags().StringVarP(&mdscfg.Swim.CoordinatorAddr, "swim-coordinator-addr", "", config.Get("swim.coordinator_addr"), "swim coordinator address")
	mdsCmd.Flags().StringVarP(&mdscfg.Swim.Period, "swim-period", "", config.Get("swim.period"), "swim ping period time")
	mdsCmd.Flags().StringVarP(&mdscfg.Swim.Expire, "swim-expire", "", config.Get("swim.expire"), "swim ping expire time")

	mdsCmd.Flags().StringVarP(&mdscfg.Security.CertsDir, "secure-certs-dir", "", config.Get("security.certs_dir"), "directory path of secure configuration files")
	mdsCmd.Flags().StringVarP(&mdscfg.Security.RootCAPem, "secure-rootca-pem", "", config.Get("security.rootca_pem"), "file name of rootCA.pem")
	mdsCmd.Flags().StringVarP(&mdscfg.Security.ServerKey, "secure-server-key", "", config.Get("security.server_key"), "file name of server key")
	mdsCmd.Flags().StringVarP(&mdscfg.Security.ServerCrt, "secure-server-crt", "", config.Get("security.server_crt"), "file name of server crt")

	mdsCmd.Flags().StringVarP(&mdscfg.LocalEncodingMatrices, "local-encoding-matrices", "", config.Get("mds.local_encoding_matrices"), "number of matrices for local encoding")
	mdsCmd.Flags().StringVarP(&mdscfg.GlobalEncodingMatrices, "global-encoding-matrices", "", config.Get("mds.global_encoding_matrices"), "number of matrices for global encoding")

	mdsCmd.Flags().StringVarP(&mdscfg.LogLocation, "log", "l", config.Get("mds.log_location"), "log location of the mds will print out")
}
