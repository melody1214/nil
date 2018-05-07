package membership

// var logger *logrus.Entry

// type handlers struct {
// 	cfg *config.Ds

// 	swimSrv    *swim.Server
// 	swimTransL *nilmux.SwimTransportLayer

// 	needCMapUpdate chan interface{}
// 	clusterService *cmap.Service
// }

// // NewHandlers creates a client handlers with necessary dependencies.
// func NewHandlers(cfg *config.Ds, clusterService *cmap.Service) Handlers {
// 	logger = mlog.GetPackageLogger("app/ds/usecase/membership")

// 	return &handlers{
// 		cfg:            cfg,
// 		needCMapUpdate: make(chan interface{}, 1),
// 		clusterService: clusterService,
// 	}
// }

// // Create makes swim server.
// func (h *handlers) Create(swimL *nilmux.Layer) (err error) {
// 	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Create")

// 	h.swimTransL = nilmux.NewSwimTransportLayer(swimL)

// 	// Setup configuration.
// 	swimConf := swim.DefaultConfig()
// 	swimConf.ID = swim.ServerID(h.cfg.ID)
// 	swimConf.Address = swim.ServerAddress(h.cfg.ServerAddr + ":" + h.cfg.ServerPort)
// 	swimConf.Coordinator = swim.ServerAddress(h.cfg.Swim.CoordinatorAddr)
// 	if t, err := time.ParseDuration(h.cfg.Swim.Period); err != nil {
// 		ctxLogger.Error(err)
// 	} else {
// 		swimConf.PingPeriod = t
// 	}
// 	if t, err := time.ParseDuration(h.cfg.Swim.Expire); err != nil {
// 		ctxLogger.Error(err)
// 	} else {
// 		swimConf.PingExpire = t
// 	}
// 	swimConf.Type = swim.DS

// 	h.swimSrv, err = swim.NewServer(swimConf, h.swimTransL)
// 	if err != nil {
// 		ctxLogger.Error(err)
// 		return
// 	}

// 	h.swimSrv.RegisterCustomHeader("cmap_ver", "0",
// 		func(have, rcv string) bool {
// 			intHave, err := strconv.Atoi(have)
// 			if err != nil {
// 				return false
// 			}

// 			intRcv, err := strconv.Atoi(rcv)
// 			if err != nil {
// 				return false
// 			}

// 			return intHave < intRcv
// 		}, h.needCMapUpdate)

// 	return nil
// }

// // Run starts swim service.
// func (h *handlers) Run() {
// 	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Run")

// 	sc := make(chan swim.PingError, 1)
// 	go h.swimSrv.Serve(sc)

// 	cmapUpdatedNotiC := h.cMap.GetUpdatedNoti(cmap.Version(0))
// 	for {
// 		select {
// 		case err := <-sc:
// 			ctxLogger.WithFields(logrus.Fields{
// 				"server":       "swim",
// 				"message type": err.Type,
// 				"destID":       err.DestID,
// 			}).Error(err.Err)
// 		case <-h.needCMapUpdate:
// 			h.cMap.Outdated()
// 		case <-cmapUpdatedNotiC:
// 			latest := h.cMap.LatestVersion()
// 			h.swimSrv.SetCustomHeader("cmap_ver", strconv.FormatInt(latest.Int64(), 10))
// 			cmapUpdatedNotiC = h.cMap.GetUpdatedNoti(latest)
// 		}
// 	}
// }

// func (h *handlers) Run(swimL *nilmux.Layer) (err error) {
// 	ctxLogger := mlog.GetMethodLogger(logger, "handlers.Run")

// 	// Setup configuration.
// 	clusterConf := cluster.DefaultConfig()
// 	clusterConf.Name = cluster.NodeName(h.cfg.ID)
// 	clusterConf.Address = cluster.NodeAddress(h.cfg.ServerAddr + ":" + h.cfg.ServerPort)
// 	clusterConf.Coordinator = cluster.NodeAddress(h.cfg.Swim.CoordinatorAddr)
// 	if t, err := time.ParseDuration(h.cfg.Swim.Period); err != nil {
// 		ctxLogger.Error(err)
// 	} else {
// 		clusterConf.PingPeriod = t
// 	}
// 	if t, err := time.ParseDuration(h.cfg.Swim.Expire); err != nil {
// 		ctxLogger.Error(err)
// 	} else {
// 		clusterConf.PingExpire = t
// 	}
// 	clusterConf.Type = cluster.DS

// 	h.clusterService.StartMembershipServer(*clusterConf, nilmux.NewSwimTransportLayer(swimL))
// 	return nil
// }

// // Handlers is the interface that provides client http handlers.
// type Handlers interface {
// 	// Create(swimL *nilmux.Layer) error
// 	// Run()
// 	Run(swimL *nilmux.Layer) error
// }
