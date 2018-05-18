package gencoding

import (
	"io"
	"net/rpc"
	"time"

	"fmt"

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
		n, err := s.cmapAPI.SearchCallNode().Name(cmap.NodeName(s.cfg.ID)).Do()
		if err != nil {
			// Wait to be updated.
			notiC := s.cmapAPI.GetUpdatedNoti(s.cmapAPI.GetLatestCMapVersion())
			<-notiC
			continue
		}

		s.nodeID = n.ID
		break
	}

	checkTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-checkTicker.C:
			s.updateUnencoded()
		}
	}
}

func (s *service) updateUnencoded() {
	egs := s.cmapAPI.FindEncodingGroupByLeader(s.nodeID)
	// fmt.Printf("egs: %v\n", egs)
	for _, eg := range egs {
		if err := s.doUpdateUnencoded(eg); err != nil {
			fmt.Println(errors.Wrap(err, "failed to update unencoded"))
		}
	}
}

func (s *service) doUpdateUnencoded(eg cmap.EncodingGroup) error {
	unenc, err := s.store.CountNonCodedChunk(eg.Vols[len(eg.Vols)-1].String(), eg.ID.String())
	if err != nil {
		fmt.Println(err)
		return err
	}

	if eg.Uenc == unenc {
		return nil
	}

	eg.Uenc = unenc
	return s.cmapAPI.UpdateEncodingGroupUnencoded(eg)
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

// PrepareEncoding selects an unencoded chunk from the given encoding group.
// Makes selected chunk ready to encode.
func (s *service) PrepareEncoding(req *nilrpc.DGEPrepareEncodingRequest, res *nilrpc.DGEPrepareEncodingResponse) error {
	chunk, err := s.store.GetNonCodedChunk(req.Vol.String(), req.EG.String())
	if err != nil {
		return err
	}
	if chunk == "" {
		fmt.Printf("Server: %s\nVol: %s, EG: %s\n", s.cfg.ID, req.Vol, req.EG)
		return fmt.Errorf("no unencoded chunk")
	}

	eg, err := s.cmapAPI.SearchCallEncGrp().ID(req.EG).Do()
	if err != nil {
		return err
	}

	for _, vID := range eg.Vols {
		v, err := s.cmapAPI.SearchCallVolume().ID(vID).Do()
		if err != nil {
			return err
		}

		n, err := s.cmapAPI.SearchCallNode().ID(v.Node).Do()
		if err != nil {
			return err
		}

		conn, err := nilrpc.Dial(n.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
		if err != nil {
			return err
		}
		defer conn.Close()

		renameReq := &nilrpc.DGERenameChunkRequest{
			Vol:      vID.String(),
			EncGrp:   eg.ID.String(),
			OldChunk: chunk,
			NewChunk: "E_" + req.Chunk,
		}
		renameRes := &nilrpc.DGERenameChunkResponse{}

		cli := rpc.NewClient(conn)
		if err := cli.Call(nilrpc.DsGencodingRenameChunk.String(), renameReq, renameRes); err != nil {
			return errors.Wrap(err, "failed to rename chunk")
		}
	}

	return nil
}

// Service provides handlers for global encoding.
type Service interface {
	RenameChunk(req *nilrpc.DGERenameChunkRequest, res *nilrpc.DGERenameChunkResponse) error
	TruncateChunk(req *nilrpc.DGETruncateChunkRequest, res *nilrpc.DGETruncateChunkResponse) error
	PrepareEncoding(req *nilrpc.DGEPrepareEncodingRequest, res *nilrpc.DGEPrepareEncodingResponse) error
}
