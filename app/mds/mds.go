package mds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyoung/nil/app/mds/application/account"
	"github.com/chanyoung/nil/app/mds/application/cluster"
	"github.com/chanyoung/nil/app/mds/application/gencoding"
	"github.com/chanyoung/nil/app/mds/application/object"
	"github.com/chanyoung/nil/app/mds/delivery"
	"github.com/chanyoung/nil/app/mds/domain/model/user"
	"github.com/chanyoung/nil/app/mds/domain/service/raft"
	"github.com/chanyoung/nil/app/mds/infrastructure/repository/mysql"
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
		userRepository    user.Repository
		objectStore       object.Repository
		gencodingStore    gencoding.Repository
		raftService       raft.Service
		raftSimpleService raft.SimpleService
	)
	if useMySQL := true; useMySQL {
		store := mysql.New(&cfg)
		userRepository = mysql.NewUserRepository(store)
		objectStore = mysql.NewObjectRepository(store)
		gencodingStore = mysql.NewGencodingRepository(store)
		raftService = store.NewRaftService()
		raftSimpleService = raftService.NewRaftSimpleService()
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

	// Setup application handlers.
	accountService := account.NewService(&cfg, raftSimpleService, userRepository)
	clusterService := cluster.NewService(&cfg, cmapService.MasterAPI(), raftService)
	objectHandlers := object.NewHandlers(objectStore)
	gencodingService, err := gencoding.NewService(&cfg, cmapService.SlaveAPI(), gencodingStore)
	if err != nil {
		return errors.Wrap(err, "failed to create global encoding service")
	}

	// Setup delivery service.
	delivery, err := delivery.SetupDeliveryService(
		&cfg, accountService, clusterService, cmapService, objectHandlers, gencodingService,
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
