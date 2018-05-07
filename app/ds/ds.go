package ds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/ds/delivery"
	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/app/ds/repository/lvstore"
	"github.com/chanyoung/nil/app/ds/repository/partstore"
	"github.com/chanyoung/nil/app/ds/usecase/admin"
	"github.com/chanyoung/nil/app/ds/usecase/object"
	"github.com/chanyoung/nil/app/ds/usecase/recovery"
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
func Bootstrap(cfg config.Ds) error {
	// Setup logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return errors.Wrap(err, "init log failed")
	}
	logger = mlog.GetPackageLogger("app/ds")

	ctxLogger := mlog.GetFunctionLogger(logger, "Bootstrap")
	ctxLogger.Info("start bootstrap ds ...")

	// Generates data server ID.
	cfg.ID = uuid.Gen()

	// Setup repository.
	var (
		store         repository.Service
		adminStore    admin.Repository
		objectStore   object.Repository
		recoveryStore recovery.Repository
	)
	if cfg.Store == "lv" {
		store = lvstore.NewService(cfg.WorkDir)
		adminStore = lvstore.NewAdminRepository(store)
		objectStore = lvstore.NewObjectRepository(store)
		recoveryStore = lvstore.NewRecoveryRepository(store)
	} else if cfg.Store == "part" {
		store = partstore.NewService(cfg.WorkDir)
		adminStore = partstore.NewAdminRepository(store)
		objectStore = partstore.NewObjectRepository(store)
		recoveryStore = partstore.NewRecoveryRepository(store)
	} else {
		return fmt.Errorf("not supported store type: %s", cfg.Store)
	}
	_ = recoveryStore
	go store.Run()

	// Setup request event factory.
	requestEventFactory := request.NewRequestEventFactory()

	// Setup cmap map.
	// cmapMap, err := cmap.NewController(cfg.Swim.CoordinatorAddr)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to init cmap map")
	// }
	cmapService, err := cmap.NewService(cmap.NodeAddress(cfg.Swim.CoordinatorAddr), mlog.GetPackageLogger("pkg/cmap"))
	if err != nil {
		return errors.Wrap(err, "failed to create cmap service")
	}

	// Setup each usecase handlers.
	adminHandlers := admin.NewHandlers(&cfg, cmapService.SlaveAPI(), adminStore)
	objectHandlers, err := object.NewHandlers(&cfg, cmapService.SlaveAPI(), requestEventFactory, objectStore)
	if err != nil {
		return errors.Wrap(err, "failed to setup object handler")
	}
	// membershipHandlers := membership.NewHandlers(&cfg, cmapService)
	// cmapmapService := cmapmap.NewService(cmapService.SlaveAPI())

	// // Starts to update cmap map.
	// cmapmapService.Run()

	// Setup delivery service.
	delivery, err := delivery.SetupDeliveryService(&cfg, adminHandlers, objectHandlers, cmapService)
	if err != nil {
		return errors.Wrap(err, "failed to setup delivery")
	}
	// delivery.Run()

	ctxLogger.Info("bootstrap ds succeeded")

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			ctxLogger.Info("received stop signal from OS")
			delivery.Stop()
			store.Stop()
			return nil
		}
	}
}
