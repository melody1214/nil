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

var dsVolumeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add volume [device path]",
	Long:  "add volume [device path]",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires a volume path")
		}
		if len(args) > 1 {
			return fmt.Errorf("requires only one volume path")
		}
		return nil
	},
	Run: dsVolumeAddRun,
}

func dsVolumeAddRun(cmd *cobra.Command, args []string) {
	devPath := args[0]

	conn, err := nilrpc.Dial(dsCfg.ServerAddr+":"+dsCfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.DCLAddVolumeRequest{
		DevicePath: devPath,
	}
	res := &nilrpc.DCLAddVolumeResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.DsClusterAddVolume.String(), req, res); err != nil {
		log.Fatal(err)
	}
}

func init() {
	dsVolumeAddCmd.Flags().StringVarP(&dsCfg.ServerAddr, "bind", "b", config.Get("ds.addr"), "will ask the ds of this address")
	dsVolumeAddCmd.Flags().StringVarP(&dsCfg.ServerPort, "port", "p", config.Get("ds.port"), "will ask the ds of this port")
}
