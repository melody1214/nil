package ds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/ds/delivery"
	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/app/ds/repository/lvstore"
	"github.com/chanyoung/nil/app/ds/usecase/admin"
	"github.com/chanyoung/nil/app/ds/usecase/membership"
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

var log *logrus.Entry

// Bootstrap build up the gateway service.
func Bootstrap(cfg config.Ds) error {
	// Setup logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return errors.Wrap(err, "init log failed")
	}
	log = mlog.GetLogger().WithField("package", "ds")

	ctxLogger := log.WithField("function", "Bootstrap")
	ctxLogger.Info("Setting logger succeeded")

	// Generates data server ID.
	cfg.ID = uuid.Gen()
	ctxLogger.WithField("uuid", cfg.ID).Info("Generating gateway UUID succeeded")

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
	} else {
		return fmt.Errorf("not supported store type: %s", cfg.Store)
	}
	_ = recoveryStore
	go store.Run()

	// Setup request event factory.
	requestEventFactory := request.NewRequestEventFactory()

	// Setup each usecase handlers.
	adminHandlers := admin.NewHandlers(&cfg, adminStore)
	objectHandlers := object.NewHandlers(requestEventFactory, objectStore)
	membershipHandlers := membership.NewHandlers(&cfg)

	// Setup cluster map.
	if err := cmap.Initial(cfg.Swim.CoordinatorAddr); err != nil {
		return errors.Wrap(err, "failed to init cluster map")
	}

	// Setup delivery service.
	delivery, err := delivery.NewDeliveryService(&cfg, adminHandlers, objectHandlers, membershipHandlers)
	if err != nil {
		return errors.Wrap(err, "failed to setup delivery")
	}
	go delivery.Run()

	// Make channel for Ctrl-C or other terminate signal is received.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sigc:
			ctxLogger.Info("Received stop signal from OS")
			delivery.Stop()
			store.Stop()
			return nil
		}
	}
}
