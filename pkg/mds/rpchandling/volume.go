package rpchandling

import (
	"fmt"
	"strconv"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// RegisterVolume receives a new volume information from ds and register it to the database.
func (h *Handler) RegisterVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	// If the id field of request is empty, then the ds
	// tries to get an id of volume.
	if req.ID == "" {
		return h.insertNewVolume(req, res)
	}
	return h.updateVolume(req, res)
}

func (h *Handler) updateVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	log.Infof("update a member %v", req)

	q := fmt.Sprintf(
		`
		UPDATE volume
		SET volume_status='%s', size='%d', free='%d', used='%d', max_chain='%d', speed='%s' 
		WHERE volume_id in ('%s')
		`, req.Status, req.Size, req.Free, req.Used, calcMaxChain(req.Size), req.Speed, req.ID,
	)

	_, err := h.store.Execute(q)
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (h *Handler) insertNewVolume(req *nilrpc.RegisterVolumeRequest, res *nilrpc.RegisterVolumeResponse) error {
	log.Infof("insert a new volume %v", req)

	q := fmt.Sprintf(
		`
		INSERT INTO volume (node_id, volume_status, size, free, used, max_chain, speed)
		SELECT node_id, '%s', '%d', '%d', '%d', '%d', '%s' FROM node WHERE node_name = '%s'
		`, req.Status, req.Size, req.Free, req.Used, calcMaxChain(req.Size), req.Speed, req.Ds,
	)

	r, err := h.store.Execute(q)
	if err != nil {
		return err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return err
	}
	res.ID = strconv.FormatInt(id, 10)

	return nil
}

func calcMaxChain(volumeSize uint64) int {
	if volumeSize <= 0 {
		return 0
	}

	// Test, chain per 10MB,
	return int(volumeSize / 10)
}
