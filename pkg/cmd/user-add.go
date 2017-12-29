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

	conn, err := nilrpc.Dial(mdscfg.ServerAddr+":"+mdscfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.AddUserRequest{Name: name}
	res := &nilrpc.AddUserResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call("Server.AddUser", req, res); err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.AccessKey)
	fmt.Println(res.SecretKey)
}

func init() {
	userAddCmd.Flags().StringVarP(&mdscfg.ServerAddr, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	userAddCmd.Flags().StringVarP(&mdscfg.ServerPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
}
