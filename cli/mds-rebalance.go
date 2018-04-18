package cli

import (
	"log"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var mdsRebalanceCmd = &cobra.Command{
	Use:   "rebalance",
	Short: "rebalance cluster",
	Long:  "rebalance cluster",
	Run:   mdsRebalanceRun,
}

var (
	mdsRebalanceBind string
	mdsRebalancePort string
	mdsRebalanceCert string
)

func mdsRebalanceRun(cmd *cobra.Command, args []string) {
	conn, err := nilrpc.Dial(mdsRebalanceBind+":"+mdsRebalancePort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MRERecoveryRequest{Type: nilrpc.Rebalance}
	res := &nilrpc.MRERecoveryResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsRecoveryRecovery.String(), req, res); err != nil {
		log.Fatal(err)
	}
}

func init() {
	mdsRebalanceCmd.Flags().StringVarP(&mdsRebalanceBind, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mdsRebalanceCmd.Flags().StringVarP(&mdsRebalancePort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
	mdsRebalanceCmd.Flags().StringVarP(&mdsRebalanceCert, "cert", "c", config.Get("security.certs_dir")+"/"+config.Get("security.rootca_pem"), "will ask the mds of this port")
}
