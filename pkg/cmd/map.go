package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/swim/swimpb"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "print cluster map",
	Long:  "print cluster map",
	Run:   mapRun,
}

func mapRun(cmd *cobra.Command, args []string) {
	cc, err := grpc.Dial(net.JoinHostPort(mdscfg.ServerAddr, mdscfg.ServerPort), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	cli := mdspb.NewMdsClient(cc)

	res, err := cli.GetClusterMap(context.Background(), &mdspb.GetClusterMapRequest{})
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range res.GetMemlist() {
		fmt.Printf(
			"| %4s | %s | %s:%s | %7s | Incarnation: %d |\n",
			swimpb.MemberType_name[int32(node.GetType())],
			node.GetUuid(),
			node.GetAddr(), node.GetPort(),
			swimpb.Status_name[int32(node.GetStatus())],
			node.GetIncarnation(),
		)
	}
}

func init() {
	mapCmd.Flags().StringVarP(&mdscfg.ServerAddr, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mapCmd.Flags().StringVarP(&mdscfg.ServerPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
}
