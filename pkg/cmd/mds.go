package cmd

import (
	"log"

	"github.com/chanyoung/nil/pkg/mds"
	"github.com/spf13/cobra"
)

var (
	serverAddr string
	serverPort string
)

var mdsCmd = &cobra.Command{
	Use:   "mds",
	Short: "mds control commands",
	Long:  "mds control commands",
	Run:   mdsRun,
}

func mdsRun(cmd *cobra.Command, args []string) {
	m, err := mds.New(serverAddr, serverPort)
	if err != nil {
		log.Fatal(err)
	}

	m.Start()
}

func init() {
	mdsCmd.Flags().StringVarP(&serverAddr, "bind", "b", "127.0.0.1", "address to which the mds will bind")
	mdsCmd.Flags().StringVarP(&serverPort, "port", "p", "51000", "port on which the mds will listen")
}
