package lvstore

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/app/ds/usecase/admin"
	"github.com/chanyoung/nil/app/ds/usecase/object"
	"github.com/chanyoung/nil/app/ds/usecase/recovery"
)

// service is the backend store service.
type service struct {
	lvs          map[string]*lv
	basePath     string
	requestQueue queue
	pushCh       chan interface{}
}

// NewService returns a new backend store service.
func NewService(basePath string) repository.Service {
	return &service{
		basePath: basePath,
		lvs:      map[string]*lv{},
		pushCh:   make(chan interface{}, 1),
	}
}

// newService returns a new backend store service.
// This is only for unit test. Do not use in real service.
func newService(basePath string) *service {
	return &service{
		basePath: basePath,
		lvs:      map[string]*lv{},
		pushCh:   make(chan interface{}, 1),
	}
}

// Run starts to serve backend store service.
func (s *service) Run() {
	checkTicker := time.NewTicker(100 * time.Millisecond)

	// TODO: change to do not polling.
	for {
		select {
		case <-s.pushCh:
			if c := s.requestQueue.pop(); c != nil {
				go s.handleCall(c)
			}
		case <-checkTicker.C:
			if c := s.requestQueue.pop(); c != nil {
				go s.handleCall(c)
			}
		}
	}
}

// Stop supports graceful stop of backend store service.
func (s *service) Stop() {
	// TODO: graceful stop.
	// Tracking all jobs and wait them until finished.

	// Deletes all volumes in the store.
	for _, lv := range s.lvs {
		lv.Umount()
	}
}

// Push pushes an io request into the scheduling queue.
func (s *service) Push(r *repository.Request) error {
	if err := r.Verify(); err != nil {
		return err
	}

	r.Wg.Add(1)
	s.requestQueue.push(r)
	s.pushCh <- nil

	return nil
}

// AddVolume adds a volume into the lv map.
func (s *service) AddVolume(v *repository.Vol) error {
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

	v.ChunkSize = 20000
	// Set volume has running state.
	v.Status = repository.Active

	return nil
}

func (s *service) handleCall(r *repository.Request) {
	defer r.Wg.Done()

	switch r.Op {
	case repository.Read:
		s.read(r)
	case repository.Write:
		s.write(r)
	case repository.Delete:
		s.delete(r)
	}
}

func (s *service) read(r *repository.Request) {
	if r.Oid == "" {
		s.readAll(r)
		return
	}

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

	LocGrpDir := lv.MntPoint + "/" + r.LocGid

	if _, err := os.Stat(LocGrpDir); os.IsNotExist(err) {
		os.MkdirAll(LocGrpDir, 0775)
	}

	// fmt.Println("---In read---")
	// fmt.Println("obj.Cid : ", obj.Cid, " obj.Offset : ", obj.Offset)

	fChunk, err := os.Open(LocGrpDir + "/" + obj.Cid)
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

func (s *service) readAll(r *repository.Request) {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	LocGrpDir := lv.MntPoint + "/" + r.LocGid

	if _, err := os.Stat(LocGrpDir); os.IsNotExist(err) {
		os.MkdirAll(LocGrpDir, 0775)
	}

	fChunk, err := os.Open(LocGrpDir + "/" + r.Cid)
	if err != nil {
		r.Err = err
		return
	}
	defer fChunk.Close()

	if _, err = io.Copy(r.Out, fChunk); err != nil {
		r.Err = err
		return
	}
}

func (s *service) write(r *repository.Request) error {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return nil
	}

	LocGrpDir := lv.MntPoint + "/" + r.LocGid

	if _, err := os.Stat(LocGrpDir); os.IsNotExist(err) {
		os.MkdirAll(LocGrpDir, 0775)
	}

	// Chunk file open
	fChunk, err := os.OpenFile(LocGrpDir+"/"+r.Cid, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
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
			lv.Objs[r.Oid] = repository.ObjMap{Cid: r.Cid, Offset: fChunkLen}
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

func (s *service) delete(r *repository.Request) {
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	r.Err = os.Remove(lv.MntPoint + "/" + r.LocGid + "/" + r.Cid)
}

// NewAdminRepository returns a new lv store inteface in a view of admin domain.
func NewAdminRepository(store repository.Service) admin.Repository {
	return store
}

// NewObjectRepository returns a new lv store inteface in a view of object domain.
func NewObjectRepository(store repository.Service) object.Repository {
	return store
}

// NewRecoveryRepository returns a new lv store inteface in a view of recovery domain.
func NewRecoveryRepository(store repository.Service) recovery.Repository {
	return store
}
