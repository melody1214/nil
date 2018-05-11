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

var mdsJobCmd = &cobra.Command{
	Use:   "job",
	Short: "listing jobs",
	Long:  "listing jobs",
	Run:   mdsJobRun,
}

var (
	mdsJobBind string
	mdsJobPort string
	mdsJobCert string
)

func mdsJobRun(cmd *cobra.Command, args []string) {
	conn, err := nilrpc.Dial(mdsJobBind+":"+mdsJobPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MCLListJobRequest{}
	res := &nilrpc.MCLListJobResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsClusterListJob.String(), req, res); err != nil {
		log.Fatal(err)
	}

	for _, j := range res.List {
		fmt.Println(j)
	}
}

func init() {
	mdsJobCmd.Flags().StringVarP(&mdsJobBind, "bind", "b", config.Get("mds.addr"), "will ask the mds of this address")
	mdsJobCmd.Flags().StringVarP(&mdsJobPort, "port", "p", config.Get("mds.port"), "will ask the mds of this port")
	mdsJobCmd.Flags().StringVarP(&mdsJobCert, "cert", "c", config.Get("security.certs_dir")+"/"+config.Get("security.rootca_pem"), "will ask the mds of this port")
}
