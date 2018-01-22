package cmd

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "print cluster map",
	Long:  "print cluster map",
	Run:   mapRun,
}

var (
	bind string
	port string
	cert string
)

func mapRun(cmd *cobra.Command, args []string) {
	conn, err := nilrpc.Dial(bind+":"+port, nilrpc.RPCNil, time.Duration(2*time.Second))
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

	for _, node := range res.Members {
		fmt.Printf(
			"| %4s | %s | %s | %7s | Incarnation: %d |\n",
			node.Type.String(),
			node.ID,
			node.Address,
			node.Status.String(),
			node.Incarnation,
		)
	}
}

func init() {
	mapCmd.Flags().StringVarP(&bind, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mapCmd.Flags().StringVarP(&port, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
	mapCmd.Flags().StringVarP(&cert, "cert", "c", config.Get("security.certs_dir")+"/"+config.Get("security.rootca_pem"), "will ask the mds of this port")
}
