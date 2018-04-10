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
	// TODO: Receive chunksize from the config.
	v.ChunkSize = 1000000

	s.lvs[v.Name] = &lv{
		Vol: v,
	}

	// Set volume has running state.
	v.Status = repository.Active

	return nil
}

func (s *service) GetObjectSize(lvID, objID string) (int64, bool) {
	lv, ok := s.lvs[lvID]
	if ok == false {
		return 0, false
	}

	obj, ok := lv.Objs[objID]
	if ok == false {
		return 0, false
	}

	return obj.Size, true
}

func (s *service) GetObjectMD5(lvID, objID string) (string, bool) {
	lv, ok := s.lvs[lvID]
	if ok == false {
		return "", false
	}

	obj, ok := lv.Objs[objID]
	if ok == false {
		return "", false
	}

	return obj.MD5, true
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
	// Q: 이 부분 의미 질문
	if r.Oid == "" {
		s.readAll(r)
		return
	}

	// Find and get the requested logical volume.
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	// Find and get the requested object.
	obj, ok := lv.Objs[r.Oid]
	if !ok {
		r.Err = fmt.Errorf("no such object: %s", r.Oid)
		return
	}

	if r.Osize == 0 {
		r.Osize = obj.Size
	}

	// Create a directory for a local group if not exist.
	lgDir := lv.MntPoint + "/" + r.LocGid
	_, err := os.Stat(lgDir)
	if os.IsNotExist(err) {
		os.MkdirAll(lgDir, 0775)
	}

	// Open a chunk requested by a client.
	fChunk, err := os.Open(lgDir + "/" + obj.Cid)
	if err != nil {
		r.Err = err
		return
	}
	defer fChunk.Close()

	// Seek offset beginning of the requested object in the chunk.
	_, err = fChunk.Seek(obj.Offset, os.SEEK_SET)
	if err != nil {
		r.Err = err
		return
	}

	// Read contents of the requested object from the chunk.
	if _, err = io.CopyN(r.Out, fChunk, r.Osize); err != nil {
		r.Err = err
		return
	}
}

func (s *service) readAll(r *repository.Request) {
	// Find and get a logical volume.
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return
	}

	// Create a directory for a local group if not exist.
	lgDir := lv.MntPoint + "/" + r.LocGid
	_, err := os.Stat(lgDir)
	if os.IsNotExist(err) {
		os.MkdirAll(lgDir, 0775)
	}

	// Open a chunk requested by a client.
	fChunk, err := os.Open(lgDir + "/" + r.Cid)
	if err != nil {
		r.Err = err
		return
	}
	defer fChunk.Close()

	// Read all contents from the chunk to a writer stream.
	_, err = io.Copy(r.Out, fChunk)
	if err != nil {
		r.Err = err
		return
	}
}

func (s *service) write(r *repository.Request) error {
	// Find and get a logical volume.
	lv, ok := s.lvs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
		return nil
	}

	// Create a directory for a local group if not exist.
	lgDir := lv.MntPoint + "/" + r.LocGid
	_, err := os.Stat(lgDir)
	if os.IsNotExist(err) {
		os.MkdirAll(lgDir, 0775)
	}

	// Open a chunk that objects will be written to.
	fChunk, err := os.OpenFile(lgDir+"/"+r.Cid, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		r.Err = err
		return nil
	}
	defer fChunk.Close()

	// Obtain a path of the chunk.
	fChunkName := fChunk.Name()

	// Get an information of the chunk.
	fChunkInfo, err := os.Lstat(fChunkName)
	if err != nil {
		r.Err = err
		return nil
	}

	// Get current length of the chunk.
	fChunkLen := fChunkInfo.Size()

	// Write the object into the chunk if it will not be full.
	if fChunkLen >= lv.ChunkSize {
		r.Err = fmt.Errorf("chunk full")
		return nil
	}

	if fChunkLen+r.Osize > lv.ChunkSize {
		err = fChunk.Truncate(lv.ChunkSize)
		r.Err = fmt.Errorf("truncated")
		return nil
	}

	_, err = io.CopyN(fChunk, r.In, r.Osize)
	if err != nil {
		r.Err = err
		return nil
	}

	// Store mapping information between the object and the chunk.
	lv.Objs[r.Oid] = repository.ObjMap{
		Cid:    r.Cid,
		Offset: fChunkLen,
		Size:   r.Osize,
		MD5:    r.Md5,
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
