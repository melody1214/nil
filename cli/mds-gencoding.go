package cli

import (
	"log"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/spf13/cobra"
)

var mdsGGGCmd = &cobra.Command{
	Use:   "ggg",
	Short: "generate global encoding group",
	Long:  "generate global encoding group",
	Run:   mdsGGGRun,
}

var (
	mdsGGGBind string
	mdsGGGPort string
	mdsGGGCert string
)

func mdsGGGRun(cmd *cobra.Command, args []string) {
	conn, err := nilrpc.Dial(mdsGGGBind+":"+mdsGGGPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MGEGGGRequest{}
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
