package admin

import (
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/chanyoung/nil/pkg/cluster"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type handlers struct {
	clusterAPI cluster.SlaveAPI
}

// NewHandlers creates an admin handlers with necessary dependencies.
func NewHandlers(clusterAPI cluster.SlaveAPI) Handlers {
	logger = mlog.GetPackageLogger("app/gw/usecase/admin")

	return &handlers{
		clusterAPI: clusterAPI,
	}
}

// Proxying forwards the rpc connection to the mds.
func (h *handlers) Proxying(conn net.Conn) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Proxying")

	// 1. Prepare dialer with security config.
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	tlsConfig := security.DefaultTLSConfig()

	// 2. Lookup mds from cluster map.
	mds, err := h.clusterAPI.SearchCallNode().Type(cluster.MDS).Status(cluster.Alive).Do()
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "find alive mds failed"))
		return
	}

	// 3. Dial with tls.
	remote, err := tls.DialWithDialer(dialer, "tcp", mds.Addr.String(), tlsConfig)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "dial to mds failed"))
		return
	}

	// 4. Forwarding.
	go io.Copy(conn, remote)
	go io.Copy(remote, conn)
}

// Handlers is the interface that provides admin rpc handlers.
type Handlers interface {
	Proxying(conn net.Conn)
}
