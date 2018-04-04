package mds

import (
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

	return nil
}
