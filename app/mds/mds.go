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
	"github.com/chanyoung/nil/app/mds/usecase/clustermap"
	"github.com/chanyoung/nil/app/mds/usecase/consensus"
	"github.com/chanyoung/nil/app/mds/usecase/membership"
	"github.com/chanyoung/nil/app/mds/usecase/object"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
	"github.com/chanyoung/nil/pkg/cluster"
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
		store := mysql.New(&cfg)
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

	// Setup membership service.
	clusterCfg := cluster.DefaultConfig()
	clusterCfg.Name = cluster.NodeName(cfg.ID)
	clusterCfg.Address = cluster.NodeAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	clusterCfg.Coordinator = cluster.NodeAddress(cfg.ServerAddr + ":" + cfg.ServerPort)
	clusterCfg.Type = cluster.MDS
	clusterService, err := cluster.NewService(*clusterCfg, mlog.GetPackageLogger("pkg/cluster"))
	if err != nil {
		return errors.Wrap(err, "failed to create cluster service")
	}
	_ = clusterService

	// Setup cluster map.
	clusterMap, err := cmap.NewController(cfg.ServerAddr + ":" + cfg.ServerPort)
	if err != nil {
		return errors.Wrap(err, "failed to init cluster map")
	}

	// Setup usecase handlers.
	adminHandlers := admin.NewHandlers(&cfg, adminStore)
	authHandlers := auth.NewHandlers(authStore)
	bucketHandlers := bucket.NewHandlers(bucketStore)
	consensusHandlers := consensus.NewHandlers(&cfg, consensusStore)
	clustermapHandlers := clustermap.NewHandlers(clusterMap, clustermapStore)
	membershipHandlers := membership.NewHandlers(&cfg, clusterMap, membershipStore)
	objectHandlers := object.NewHandlers(objectStore)
	recoveryHandlers, err := recovery.NewHandlers(&cfg, clusterMap, recoveryStore)
	if err != nil {
		return errors.Wrap(err, "failed to create recovery handler")
	}

	// Setup delivery service.
	delivery, err := delivery.NewDeliveryService(
		&cfg, adminHandlers, authHandlers, bucketHandlers, consensusHandlers,
		clustermapHandlers, membershipHandlers, objectHandlers, recoveryHandlers,
	)
	if err != nil {
		return err
	}
	if err := delivery.Run(); err != nil {
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
