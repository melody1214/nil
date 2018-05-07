package gw

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/gw/delivery"
	"github.com/chanyoung/nil/app/gw/repository/inmem"
	"github.com/chanyoung/nil/app/gw/usecase/admin"
	"github.com/chanyoung/nil/app/gw/usecase/auth"
	"github.com/chanyoung/nil/app/gw/usecase/client"
	"github.com/chanyoung/nil/app/gw/usecase/clustermap"
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
	authCache := inmem.NewAuthRepository()

	// Setup request event factory.
	requestEventFactory := request.NewRequestEventFactory()

	// // Setup cmap map.
	// cmapMap, err := cmap.NewController(cfg.FirstMds)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to init cmap map")
	// }
	cmapService, err := cmap.NewService(cmap.NodeAddress(cfg.FirstMds), mlog.GetPackageLogger("pkg/cmap"))
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
