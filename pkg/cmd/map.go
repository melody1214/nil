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
	"google.golang.org/grpc/credentials"
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
	creds, err := credentials.NewClientTLSFromFile(
		cert,
		"localhost",
	)
	if err != nil {
		log.Fatal(err)
	}

	cc, err := grpc.Dial(net.JoinHostPort(bind, port), grpc.WithTransportCredentials(creds))
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
	mapCmd.Flags().StringVarP(&bind, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mapCmd.Flags().StringVarP(&port, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
	mapCmd.Flags().StringVarP(&cert, "cert", "c", config.Get("security.certs_dir")+"/"+config.Get("security.rootca_pem"), "will ask the mds of this port")
}
