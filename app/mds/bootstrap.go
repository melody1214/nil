package mds

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/repository/mysql"
	"github.com/chanyoung/nil/app/mds/usecase/admin"
	"github.com/chanyoung/nil/app/mds/usecase/auth"
	"github.com/chanyoung/nil/app/mds/usecase/bucket"
	"github.com/chanyoung/nil/app/mds/usecase/clustermap"
	"github.com/chanyoung/nil/app/mds/usecase/membership"
	"github.com/chanyoung/nil/app/mds/usecase/object"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/uuid"
)

// var log *logrus.Entry

// Bootstrap build up the metadata server.
func Bootstrap(cfg config.Mds) error {
	// // Setup logger.
	// if err := mlog.Init(cfg.LogLocation); err != nil {
	// 	return errors.Wrap(err, "init log failed")
	// }
	// log = mlog.GetLogger().WithField("package", "mds")

	// ctxLogger := log.WithField("method", "Bootstrap")
	// ctxLogger.Info("Setting logger succeeded")

	// Generates mds ID.
	cfg.ID = uuid.Gen()
	// ctxLogger.WithField("uuid", cfg.ID).Info("Generating metadata server UUID succeeded")

	// Setup repositories.
	var (
		store           repository.Store
		adminStore      admin.Repository
		authStore       auth.Repository
		bucketStore     bucket.Repository
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
		clustermapStore = mysql.NewClusterMapRepository(store)
		membershipStore = mysql.NewMembershipRepository(store)
		objectStore = mysql.NewObjectRepository(store)
		recoveryStore = mysql.NewRecoveryRepository(store)
	} else {
		return fmt.Errorf("not supported store type")
	}
	_ = adminStore
	_ = authStore
	_ = bucketStore
	_ = clustermapStore
	_ = membershipStore
	_ = objectStore
	_ = recoveryStore

	return nil
}
