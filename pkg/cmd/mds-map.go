package cmd

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
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

	req := &nilrpc.GetClusterMapRequest{}
	res := &nilrpc.GetClusterMapResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.GetClusterMap.String(), req, res); err != nil {
		log.Fatal(err)
	}

	m := &cmap.CMap{}
	for _, n := range res.Nodes {
		m.Nodes = append(m.Nodes, cmap.Node{
			ID:   cmap.ID(n.ID),
			Name: n.Name,
			Addr: n.Addr,
			Type: cmap.Type(n.Type),
			Stat: cmap.Status(n.Stat),
		})
	}

	fmt.Println(m.HumanReadable())
}

func init() {
	mdsMapCmd.Flags().StringVarP(&mdsMapBind, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mdsMapCmd.Flags().StringVarP(&mdsMapPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
	mdsMapCmd.Flags().StringVarP(&mdsMapCert, "cert", "c", config.Get("security.certs_dir")+"/"+config.Get("security.rootca_pem"), "will ask the mds of this port")
}
