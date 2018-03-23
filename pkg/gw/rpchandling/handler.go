package rpchandling

import (
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// TypeBytes returns rpc type bytes which is used to multiplexing.
func TypeBytes() []byte {
	return []byte{
		0x01, // rpcRaft
		0x02, // rpcNil
	}
}

// Handler offers methods that can handle rpc requests.
type Handler struct {
	clusterMap *cmap.CMap
}

// NewHandler returns a new rpc handler.
func NewHandler() *Handler {
	log = mlog.GetLogger().WithField("package", "gw/rpchandling")

	return &Handler{
		clusterMap: cmap.New(),
	}
}

// Proxying forwards to mds requests.
func (h *Handler) Proxying(conn net.Conn) {
	// 1. Prepare dialer with security config.
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	tlsConfig := security.DefaultTLSConfig()

	// 2. Lookup mds from cluster map.
	mds, err := h.clusterMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		h.updateClusterMap()
		mds, err = h.clusterMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			log.WithField("method", "Handler.Proxying").Error(errors.Wrap(err, "find alive mds failed"))
			return
		}
	}

	// 3. Dial with tls.
	remote, err := tls.DialWithDialer(dialer, "tcp", mds.Addr, tlsConfig)
	if err != nil {
		log.WithField("method", "Handler.Proxying").Error(errors.Wrap(err, "dial to mds failed"))
		return
	}

	// 4. Forwarding.
	go io.Copy(conn, remote)
	go io.Copy(remote, conn)
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *Handler) updateClusterMap() {
	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		log.WithField("method", "Handler.updateClusterMap").Error(err)
		return
	}

	h.clusterMap = m
}
