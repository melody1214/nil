package mds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/app/mds/repository/mysql"
	"github.com/chanyoung/nil/app/mds/usecase/admin"
	"github.com/chanyoung/nil/app/mds/usecase/auth"
	"github.com/chanyoung/nil/app/mds/usecase/bucket"
	"github.com/chanyoung/nil/app/mds/usecase/cluster"
	"github.com/chanyoung/nil/app/mds/usecase/consensus"
	"github.com/chanyoung/nil/app/mds/usecase/object"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Bootstrap build up the metadata server.
func Bootstrap(cfg config.Mds) error {
	// Setup logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return errors.Wrap(err, "init log failed")
	}
	logger = mlog.GetPackageLogger("app/mds")

	ctxLogger := mlog.GetFunctionLogger(logger, "Bootstrap")
	ctxLogger.Info("start bootstrap mds ...")

	// Generates mds ID.
	cfg.ID = uuid.Gen()

	// Setup repositories.
	var (
		adminStore     admin.Repository
		authStore      auth.Repository
		bucketStore    bucket.Repository
		consensusStore consensus.Repository
		clusterStore   cluster.Repository
		objectStore    object.Repository
		recoveryStore  recovery.Repository
	)
	if useMySQL := true; useMySQL {
		store := mysql.New(&cfg)
		adminStore = mysql.NewAdminRepository(store)
		authStore = mysql.NewAuthRepository(store)
		bucketStore = mysql.NewBucketRepository(store)
		consensusStore = mysql.NewConsensusRepository(store)
		clusterStore = mysql.NewClusterRepository(store)
		objectStore = mysql.NewObjectRepository(store)
		recoveryStore = mysql.NewRecoveryRepository(store)
	} else {
		return fmt.Errorf("not supported store type")
	}

	// Setup cluster map service.
	// This service is maintained by cluster domain, however the all domains
	// require this service necessarily. So create service in bootstrap code
	// and inject the service to all domains.
	cmapService, err := cmap.NewService(
		cmap.NodeAddress(cfg.Swim.CoordinatorAddr),
		mlog.GetPackageLogger("pkg/cmap"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cmap service")
	}

	// Setup usecase handlers.
	adminHandlers := admin.NewHandlers(&cfg, adminStore)
	authHandlers := auth.NewHandlers(authStore)
	bucketHandlers := bucket.NewHandlers(bucketStore)
	consensusHandlers := consensus.NewHandlers(&cfg, consensusStore)
	clusterService := cluster.NewService(&cfg, cmapService.MasterAPI(), clusterStore)
	objectHandlers := object.NewHandlers(objectStore)
	recoveryHandlers, err := recovery.NewHandlers(&cfg, cmapService.SlaveAPI(), recoveryStore)
	if err != nil {
		return errors.Wrap(err, "failed to create recovery handler")
	}

	// Setup delivery service.
	delivery, err := delivery.SetupDeliveryService(
		&cfg, adminHandlers, authHandlers, bucketHandlers, consensusHandlers,
		clusterService, cmapService, objectHandlers, recoveryHandlers,
	)
	if err != nil {
		return err
	}
	ctxLogger.Info("bootstrap mds succeeded")

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
