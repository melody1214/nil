package admin

import (
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/chanyoung/nil/app/gw/delivery"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/security"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type handlers struct {
	cMap *cmap.CMap
}

func NewHandlers() delivery.AdminHandlers {
	log = mlog.GetLogger().WithField("package", "gw/usecase/admin")

	return &handlers{
		cMap: cmap.New(),
	}
}

func (h *handlers) Proxying(conn net.Conn) {
	ctxLogger := log.WithField("method", "handlers.Proxying")

	// 1. Prepare dialer with security config.
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	tlsConfig := security.DefaultTLSConfig()

	// 2. Lookup mds from cluster map.
	mds, err := h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
	if err != nil {
		h.updateClusterMap()
		mds, err = h.cMap.SearchCall().Type(cmap.MDS).Status(cmap.Alive).Do()
		if err != nil {
			ctxLogger.Error(errors.Wrap(err, "find alive mds failed"))
			return
		}
	}

	// 3. Dial with tls.
	remote, err := tls.DialWithDialer(dialer, "tcp", mds.Addr, tlsConfig)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "dial to mds failed"))
		return
	}

	// 4. Forwarding.
	go io.Copy(conn, remote)
	go io.Copy(remote, conn)
}

// updateClusterMap retrieves the latest cluster map from the mds.
func (h *handlers) updateClusterMap() {
	ctxLogger := log.WithField("method", "handlers.updateClusterMap")

	m, err := cmap.GetLatest(cmap.WithFromRemote(true))
	if err != nil {
		ctxLogger.Error(err)
		return
	}

	h.cMap = m
}
