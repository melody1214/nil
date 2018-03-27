package lvstore

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chanyoung/nil/pkg/ds/store/request"
	"github.com/chanyoung/nil/pkg/ds/store/volume"
)

// Service is the backend store service.
type Service struct {
	lvs          map[string]*lv
	basePath     string
	requestQueue queue
	pushCh       chan interface{}
}

// NewService returns a new backend store service.
func NewService(basePath string) *Service {
	return &Service{
		basePath: basePath,
		lvs:      map[string]*lv{},
		pushCh:   make(chan interface{}, 1),
	}
}

// Run starts to serve backend store service.
func (s *Service) Run() {
	checkTicker := time.NewTicker(100 * time.Millisecond)

	// TODO: change to do not polling.
	for {
		select {
		case <-s.pushCh:
			if c := s.requestQueue.pop(); c != nil {
				s.handleCall(c)
			}
		case <-checkTicker.C:
			if c := s.requestQueue.pop(); c != nil {
				s.handleCall(c)
			}
		}
	}
}

// Stop supports graceful stop of backend store service.
func (s *Service) Stop() {
	// TODO: graceful stop.
	// Tracking all jobs and wait them until finished.

	// Deletes all volumes in the store.
	for name, lv := range s.lvs {
		delete(s.lvs, name)
		lv.Umount()
	}
}

// Push pushes an io request into the scheduling queue.
func (s *Service) Push(r *request.Request) error {
	if err := r.Verify(); err != nil {
		return err
	}

	r.Wg.Add(1)
	s.requestQueue.push(r)
	s.pushCh <- nil

	return nil
}

// AddVolume adds a volume into the lv map.
func (s *Service) AddVolume(v *volume.Vol) error {
	if _, ok := s.lvs[v.Name]; ok {
		return fmt.Errorf("Volume name %s already exists", v.Name)
	}

	if err := v.Mount(); err != nil {
		return err
	}

	// Update filesystem stats.
	if err := v.UpdateStatFs(); err != nil {
		return err
	}

	// TODO: Set the disk speed.
	v.SetSpeed()

	s.lvs[v.Name] = &lv{
		Vol: v,
	}

	v.ChunkSize = 100000
	// Set volume has running state.
	v.Status = volume.Active

	return nil
}

func (s *Service) handleCall(r *request.Request) {
	defer r.Wg.Done()

	switch r.Op {
	case request.Read:
		s.read(r)
	case request.Write:
		s.write(r)
	case request.Delete:
		s.delete(r)
	}
}

func (s *Service) read(r *request.Request) {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	// fmt.Println("---In read---")
	// for key, value := range lv.Objs {
	// fmt.Println(key, value)
	// }

	obj, ok := lv.Objs[r.Oid]
	if !ok {
		r.Err = fmt.Errorf("no such object: %s", r.Oid)
		return
	}

	// fmt.Println("---In read---")
	// fmt.Println("obj.Cid : ", obj.Cid, " obj.Offset : ", obj.Offset)

	fChunk, err := os.Open(lv.MntPoint + "/" + obj.Cid)
	if err != nil {
		r.Err = err
		return
	}

	// fmt.Println("File open completed")

	defer fChunk.Close()

	_, err = fChunk.Seek(obj.Offset, os.SEEK_SET)
	if err != nil {
		r.Err = err
		return
	}

	if _, err = io.CopyN(r.Out, fChunk, r.Osize); err != nil {
		r.Err = err
		return
	}
}

func (s *Service) write(r *request.Request) error {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return nil
	}

	// Chunk file open
	fChunk, err := os.OpenFile(lv.MntPoint+"/"+r.Cid, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		r.Err = err
		return nil
	}
	defer fChunk.Close()

	// Check chunk file info
	fChunkName := fChunk.Name()

	fChunkInfo, err := os.Lstat(fChunkName)
	if err != nil {
		r.Err = err
		return nil
	}

	fChunkLen := fChunkInfo.Size()

	// fmt.Println("request : ", r.Op, r.Vol, r.Oid, r.Cid, r.Osize)
	// fmt.Println("fChunkLen : ", fChunkLen, ", lv.ChunkSize : ", lv.ChunkSize)

	if fChunkLen < lv.ChunkSize {
		if lv.ChunkSize-fChunkLen >= r.Osize {
			if _, err = io.CopyN(fChunk, r.In, r.Osize); err != nil {
				return nil
			}

			if lv.ChunkSize-fChunkLen == r.Osize {
				r.Err = fmt.Errorf("chunk full")
			}
			lv.Objs[r.Oid] = volume.ObjMap{Cid: r.Cid, Offset: fChunkLen}
		} else {
			if err := fChunk.Truncate(lv.ChunkSize); err != nil {
				r.Err = err
				return err
			}

			// fmt.Println("Truncated")
			r.Err = fmt.Errorf("truncated")
		}
	}

	return nil
}

func (s *Service) delete(r *request.Request) {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	r.Err = os.Remove(lv.MntPoint + "/" + r.Oid)
}
