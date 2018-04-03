package gw

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/gw/delivery"
	"github.com/chanyoung/nil/app/gw/repository/inmem"
	"github.com/chanyoung/nil/app/gw/usecase/admin"
	"github.com/chanyoung/nil/app/gw/usecase/auth"
	"github.com/chanyoung/nil/app/gw/usecase/bucket"
	"github.com/chanyoung/nil/app/gw/usecase/object"
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
func Bootstrap(cfg config.Gw) error {
	// Setup logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return errors.Wrap(err, "init log failed")
	}
	log = mlog.GetLogger().WithField("package", "gw")
	if log == nil {
		return errors.New("init log failed: nil logger object")
	}
	ctxLogger := log.WithField("method", "New")
	ctxLogger.Info("Setting logger succeeded")

	// Generates gateway ID.
	cfg.ID = uuid.Gen()
	ctxLogger.WithField("uuid", cfg.ID).Info("Generating gateway UUID succeeded")

	// Setup repository.
	authCache := inmem.NewAuthRepository()

	// Setup request event factory.
	requestEventFactory := request.NewRequestEventFactory()

	// Setup each usecase handlers.
	authHandlers := auth.NewAuthHandlers(authCache)

	adminHandlers := admin.NewAdminHandlers()
	bucketHandlers := bucket.NewBucketHandlers(requestEventFactory, authHandlers)
	objectHandlers := object.NewObjectHandlers(requestEventFactory, authHandlers)

	// Setup cluster map.
	if err := cmap.Initial(cfg.FirstMds); err != nil {
		return errors.Wrap(err, "failed to init cluster map")
	}

	// Setup delivery.
	delivery, err := delivery.NewDeliveryService(&cfg, adminHandlers, bucketHandlers, objectHandlers)
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
			return nil
		}
	}
}
