package ds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/ds/application/cluster"
	"github.com/chanyoung/nil/app/ds/application/gencoding"
	"github.com/chanyoung/nil/app/ds/application/object"
	"github.com/chanyoung/nil/app/ds/delivery"
	"github.com/chanyoung/nil/app/ds/domain/model/device"
	"github.com/chanyoung/nil/app/ds/domain/model/volume"
	"github.com/chanyoung/nil/app/ds/infrastructure/repository"
	"github.com/chanyoung/nil/app/ds/infrastructure/repository/partstore"
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
		store            repository.Service
		objectStore      object.Repository
		gencodingStore   gencoding.Repository
		deviceRepository device.Repository
		volumeRepository volume.Repository
	)
	if cfg.Store == "part" {
		store = partstore.NewService(cfg.WorkDir)
		objectStore = partstore.NewObjectRepository(store)
		gencodingStore = partstore.NewGencodingRepository(store)
		deviceRepository = store.NewDeviceRepository()
		volumeRepository = store.NewVolumeRepository()
	} else {
		return fmt.Errorf("not supported store type: %s", cfg.Store)
	}
	go store.Run()

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

	// Setup each application handlers.
	clusterService := cluster.NewService(&cfg, cmapService.SlaveAPI(), deviceRepository, volumeRepository)
	objectHandlers, err := object.NewHandlers(&cfg, cmapService.SlaveAPI(), requestEventFactory, objectStore)
	if err != nil {
		return errors.Wrap(err, "failed to setup object handler")
	}
	gencodingService := gencoding.NewService(&cfg, cmapService.SlaveAPI(), gencodingStore)

	// Setup delivery service.
	delivery, err := delivery.SetupDeliveryService(&cfg, clusterService, objectHandlers, cmapService, gencodingService)
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
