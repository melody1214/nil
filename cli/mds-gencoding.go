package cli

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var mdsGGGCmd = &cobra.Command{
	Use:   "ggg [region1,region2,region3,region4]",
	Short: "generate global encoding group",
	Long:  "generate global encoding group",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires region names")
		}
		if len(args) > 1 {
			return fmt.Errorf("includes invalid arguments")
		}
		return nil
	},
	Run: mdsGGGRun,
}

var (
	mdsGGGBind string
	mdsGGGPort string
	mdsGGGCert string
)

func mdsGGGRun(cmd *cobra.Command, args []string) {
	regions := strings.Split(args[0], ",")
	if len(regions) != 4 {
		log.Fatal(fmt.Errorf("invalid region numbers"))
	}

	conn, err := nilrpc.Dial(mdsGGGBind+":"+mdsGGGPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MGEGGGRequest{
		Regions: regions,
	}
	res := &nilrpc.MGEGGGResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsGencodingGGG.String(), req, res); err != nil {
		log.Fatal(err)
	}
}

func init() {
	mdsGGGCmd.Flags().StringVarP(&mdsGGGBind, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mdsGGGCmd.Flags().StringVarP(&mdsGGGPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
	mdsGGGCmd.Flags().StringVarP(&mdsGGGCert, "cert", "c", config.Get("security.certs_dir")+"/"+config.Get("security.rootca_pem"), "will ask the mds of this port")
}
