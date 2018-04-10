package cli

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var mdsMapCmd = &cobra.Command{
	Use:   "map",
	Short: "print cluster map",
	Long:  "print cluster map",
	Run:   mdsMapRun,
}

var (
	mdsMapBind string
	mdsMapPort string
	mdsMapCert string
)

func mdsMapRun(cmd *cobra.Command, args []string) {
	conn, err := nilrpc.Dial(mdsMapBind+":"+mdsMapPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MCLGetClusterMapRequest{Version: 0}
	res := &nilrpc.MCLGetClusterMapResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClustermapGetClusterMap.String(), req, res); err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.ClusterMap.HumanReadable())
}

func init() {
	mdsMapCmd.Flags().StringVarP(&mdsMapBind, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mdsMapCmd.Flags().StringVarP(&mdsMapPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
	mdsMapCmd.Flags().StringVarP(&mdsMapCert, "cert", "c", config.Get("security.certs_dir")+"/"+config.Get("security.rootca_pem"), "will ask the mds of this port")
}
