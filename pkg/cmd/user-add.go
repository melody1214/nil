package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/chanyoung/nil/pkg/mds/mdspb"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add user [user name]",
	Long:  "add user [user name]",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires an user name")
		}
		if len(args) > 1 {
			return fmt.Errorf("requires only one user name")
		}
		return nil
	},
	Run: userAddRun,
}

func userAddRun(cmd *cobra.Command, args []string) {
	name := args[0]

	cc, err := grpc.Dial(net.JoinHostPort(mdscfg.ServerAddr, mdscfg.ServerPort), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	cli := mdspb.NewMdsClient(cc)

	res, err := cli.AddUser(context.Background(), &mdspb.AddUserRequest{
		Name: name,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.GetAccessKey())
	fmt.Println(res.GetSecretKey())
}

func init() {
	userAddCmd.Flags().StringVarP(&mdscfg.ServerAddr, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	userAddCmd.Flags().StringVarP(&mdscfg.ServerPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
}
