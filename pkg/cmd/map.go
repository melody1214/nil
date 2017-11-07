package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
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

	res, err := cli.PrintMap(context.Background(), &mdspb.PrintMapRequest{})
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range res.GetMemlist() {
		fmt.Print(node)
		fmt.Println(node.Status.String())
	}
}

func init() {
	mapCmd.Flags().StringVarP(&mdscfg.ServerAddr, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mapCmd.Flags().StringVarP(&mdscfg.ServerPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
}
