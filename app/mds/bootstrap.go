package mds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/repository/mysql"
	"github.com/chanyoung/nil/app/mds/usecase/admin"
	"github.com/chanyoung/nil/app/mds/usecase/auth"
	"github.com/chanyoung/nil/app/mds/usecase/bucket"
	"github.com/chanyoung/nil/app/mds/usecase/clustermap"
	"github.com/chanyoung/nil/app/mds/usecase/consensus"
	"github.com/chanyoung/nil/app/mds/usecase/membership"
	"github.com/chanyoung/nil/app/mds/usecase/object"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/chanyoung/nil/pkg/util/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// Bootstrap build up the metadata server.
func Bootstrap(cfg config.Mds) error {
	// Setup logger.
	if err := mlog.Init(cfg.LogLocation); err != nil {
		return errors.Wrap(err, "init log failed")
	}
	log = mlog.GetLogger().WithField("package", "mds")

	ctxLogger := log.WithField("method", "Bootstrap")
	ctxLogger.Info("Setting logger succeeded")

	// Generates mds ID.
	cfg.ID = uuid.Gen()
	ctxLogger.WithField("uuid", cfg.ID).Info("Generating metadata server UUID succeeded")

	// Setup repositories.
	var (
		store           repository.Store
		adminStore      admin.Repository
		authStore       auth.Repository
		bucketStore     bucket.Repository
		consensusStore  consensus.Repository
		clustermapStore clustermap.Repository
		membershipStore membership.Repository
		objectStore     object.Repository
		recoveryStore   recovery.Repository
	)
	if useMySQL := true; useMySQL {
		store = mysql.New(&cfg)
		adminStore = mysql.NewAdminRepository(store)
		authStore = mysql.NewAuthRepository(store)
		bucketStore = mysql.NewBucketRepository(store)
		consensusStore = mysql.NewConsensusRepository(store)
		clustermapStore = mysql.NewClusterMapRepository(store)
		membershipStore = mysql.NewMembershipRepository(store)
		objectStore = mysql.NewObjectRepository(store)
		recoveryStore = mysql.NewRecoveryRepository(store)
	} else {
		return fmt.Errorf("not supported store type")
	}
	_ = objectStore

	// Setup usecase handlers.
	adminHandlers := admin.NewHandlers(&cfg, adminStore)
	authHandlers := auth.NewHandlers(authStore)
	bucketHandlers := bucket.NewHandlers(bucketStore)
	consensusHandlers := consensus.NewHandlers(&cfg, consensusStore)
	clustermapHandlers := clustermap.NewHandlers(clustermapStore)
	membershipHandlers := membership.NewHandlers(&cfg, membershipStore)
	recoveryHandlers := recovery.NewHandlers(&cfg, recoveryStore)

	// Setup cluster map.
	if err := cmap.Initial(cfg.ServerAddr + ":" + cfg.ServerPort); err != nil {
		return err
	}

	// Setup delivery service.
	delivery, err := delivery.NewDeliveryService(
		&cfg, adminHandlers, authHandlers, bucketHandlers, consensusHandlers,
		clustermapHandlers, membershipHandlers, recoveryHandlers,
	)
	if err != nil {
		return err
	}
	if err := delivery.Run(); err != nil {
		return err
	}

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
