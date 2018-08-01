package gw

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/gw/application/admin"
	"github.com/chanyoung/nil/app/gw/application/auth"
	"github.com/chanyoung/nil/app/gw/application/client"
	"github.com/chanyoung/nil/app/gw/application/clustermap"
	"github.com/chanyoung/nil/app/gw/delivery"
	"github.com/chanyoung/nil/app/gw/infrastructure/repository/inmem"
	"github.com/chanyoung/nil/pkg/client/request"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Bootstrap build up the gateway service.
func Bootstrap(cfg config.Gw) error {
	// Setup logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return errors.Wrap(err, "init log failed")
	}
	logger = mlog.GetPackageLogger("app/gw")

	ctxLogger := mlog.GetFunctionLogger(logger, "Bootstrap")
	ctxLogger.Info("start bootstrap gw ...")

	// Generates gateway ID.
	cfg.ID = uuid.Gen()

	// Setup repository.
	authCache := inmem.NewCredRepository()

	// Setup request event factory.
	requestEventFactory := request.NewRequestEventFactory()

	// Setup cluster map service.
	// This service is maintained by cluster domain, however the all domains
	// require this service necessarily. So create service in bootstrap code
	// and inject the service to all domains.
	cmapService, err := cmap.NewService(mlog.GetPackageLogger("pkg/cmap"))
	if err != nil {
		return errors.Wrap(err, "failed to create cmap service")
	}

	// Setup each usecase handlers.
	authHandlers := auth.NewHandlers(cmapService.SlaveAPI(), authCache)
	adminHandlers := admin.NewHandlers(cmapService.SlaveAPI())
	clientHandlers := client.NewHandlers(cmapService.SlaveAPI(), requestEventFactory, authHandlers)
	cmapmapService := cmapmap.NewService(cmapService)

	// Starts to update cmap map.
	cmapmapService.Run()

	// Setup delivery service.
	delivery, err := delivery.NewDeliveryService(&cfg, adminHandlers, clientHandlers)
	if err != nil {
		return errors.Wrap(err, "failed to setup delivery")
	}
	delivery.Run()

	ctxLogger.Info("bootstrap gw succeeded")

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			ctxLogger.Info("Received stop signal from OS")
			delivery.Stop()
			return nil
		}
	}
}
