package gencoding

import (
	"fmt"
	"io"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// service manages the cluster map.
type service struct {
	cfg     *config.Ds
	cmapAPI cmap.SlaveAPI
	store   Repository

	nodeID cmap.ID
}

// NewService returns a new instance of a global encoding service.
func NewService(cfg *config.Ds, cmapAPI cmap.SlaveAPI, store Repository) Service {
	logger = mlog.GetPackageLogger("app/ds/usecase/gencoding")

	s := &service{
		cfg:     cfg,
		cmapAPI: cmapAPI,
		store:   store,
	}
	go s.run()

	return s
}

func (s *service) run() {
	// Get node information
	for {
		c := s.cmapAPI.SearchCall()
		n, err := c.Node().Name(cmap.NodeName(s.cfg.ID)).Do()
		if err != nil {
			// Wait to be updated.
			notiC := s.cmapAPI.GetUpdatedNoti(c.Version())
			<-notiC
			continue
		}

		s.nodeID = n.ID
		break
	}

	// 	checkTicker := time.NewTicker(10 * time.Second)
	// 	for {
	// 		select {
	// 		case <-checkTicker.C:
	// 			s.updateUnencoded()
	// 		}
	// 	}
}

func (s *service) updateUnencoded() {
	c := s.cmapAPI.SearchCall()
	myVols, err := c.Volume().Node(s.nodeID).DoAll()
	if err != nil {
		if err != cmap.ErrNotFound {
			logger.Error(errors.Wrap(err, "failed to find owned volumes"))
		}
		return
	}

	for _, v := range myVols {
		egs, err := c.EncGrp().LeaderVol(v.ID).DoAll()
		if err != nil {
			if err != cmap.ErrNotFound {
				logger.Error(errors.Wrap(err, "failed to find owned volumes"))
			}
			continue
		}

		for _, eg := range egs {
			uenc, err := s.store.CountNonCodedChunk(eg.LeaderVol().String(), eg.ID.String())
			if err != nil {
				logger.Error(errors.Wrap(err, "failed to count unencoded chunk"))
				continue
			}

			if err := s.cmapAPI.UpdateUnencoded(eg.ID, uenc); err != nil {
				logger.Error(errors.Wrap(err, "failed to update unencoded"))
				continue
			}
		}
	}
}

func (s *service) RenameChunk(req *nilrpc.DGERenameChunkRequest, res *nilrpc.DGERenameChunkResponse) error {
	return s.store.RenameChunk(req.OldChunk, req.NewChunk, req.Vol, req.EncGrp)
}

func (s *service) TruncateChunk(req *nilrpc.DGETruncateChunkRequest, res *nilrpc.DGETruncateChunkResponse) error {
	truncateReq := &repository.Request{
		Op:     repository.Write,
		Vol:    req.Vol,
		LocGid: req.EncGrp,
		Oid:    "fake, just for truncating",
		Osize:  1000000000,
		Cid:    req.Chunk,
		In:     &io.PipeReader{},
		Md5:    "fakemd5stringfakemd5stringfakemd",
	}
	if err := s.store.Push(truncateReq); err != nil {
		return errors.Wrap(err, "failed to push truncated request")
	}
	if err := truncateReq.Wait(); err == nil {
		return fmt.Errorf("truncate request returns no error")
	} else if err.Error() != "truncated" {
		return err
	}

	return nil
}

// Service provides handlers for global encoding.
type Service interface {
	RenameChunk(req *nilrpc.DGERenameChunkRequest, res *nilrpc.DGERenameChunkResponse) error
	TruncateChunk(req *nilrpc.DGETruncateChunkRequest, res *nilrpc.DGETruncateChunkResponse) error
	GetCandidateChunk(req *nilrpc.DGEGetCandidateChunkRequest, res *nilrpc.DGEGetCandidateChunkResponse) error
	Encode(req *nilrpc.DGEEncodeRequest, res *nilrpc.DGEEncodeResponse) error
}
