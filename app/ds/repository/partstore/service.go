package partstore

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/app/ds/usecase/cluster"
	"github.com/chanyoung/nil/app/ds/usecase/object"
)

// service is the backend store service.
type service struct {
	pgs          map[string]*pg
	basePath     string
	requestQueue queue
	pushCh       chan interface{}
}

// NewService returns a new backend store service.
func NewService(basePath string) repository.Service {
	return &service{
		basePath: basePath,
		pgs:      map[string]*pg{},
		pushCh:   make(chan interface{}, 1),
	}
}

// newService returns a new backend store service.
// This is only for unit test. Do not use in real service.
func newService(basePath string) *service {
	return &service{
		basePath: basePath,
		pgs:      map[string]*pg{},
		pushCh:   make(chan interface{}, 1),
	}
}

// Run starts to serve backend store service.
func (s *service) Run() {
	checkTicker := time.NewTicker(100 * time.Millisecond)

	// TODO: change to do not pollin.
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
	for _, pg := range s.pgs {
		pg.Umount()
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

// AddVolume adds a volume into the pg map.
func (s *service) AddVolume(v *repository.Vol) error {
	if _, ok := s.pgs[v.Name]; ok {
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

	s.pgs[v.Name] = &pg{
		Vol: v,
	}

	// Set volume has running state.
	v.Status = repository.Active

	return nil
}

func (s *service) GetObjectSize(pgID, objID string) (int64, bool) {
	pg, ok := s.pgs[pgID]
	if ok == false {
		return 0, false
	}

	pg.Lock.RLock()
	obj, ok := pg.ObjInfo[objID]
	pg.Lock.RUnlock()
	if ok == false {
		return 0, false
	}

	return obj.Size, true
}

func (s *service) GetObjectMD5(pgID, objID string) (string, bool) {
	pg, ok := s.pgs[pgID]
	if ok == false {
		return "", false
	}

	pg.Lock.RLock()
	obj, ok := pg.ObjInfo[objID]
	pg.Lock.RUnlock()
	if ok == false {
		return "", false
	}

	return obj.MD5, true
}

func (s *service) GetChunkHeaderSize() int64 {
	// TODO: fill the method
	return 100
}

func (s *service) GetObjectHeaderSize() int64 {
	// TODO: fill the method
	return 100
}

func (s *service) handleCall(r *repository.Request) {
	defer r.Wg.Done()

	switch r.Op {
	case repository.Read:
		s.read(r)
	case repository.Write:
		s.write(r)
	case repository.WriteAll:
		s.writeAll(r)
	case repository.Delete:
		s.delete(r)
	case repository.ReadAll:
		s.readAll(r)
	}
}

func (s *service) read(r *repository.Request) {
	// Find and get the requested logical volume.
	pg, ok := s.pgs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
		return
	}

	// Find and get the requested object.
	pg.Lock.RLock()
	chk, ok := pg.ChunkMap[r.Cid]
	pg.Lock.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no chunk of such object: %s", r.Oid)
		return
	}

	pg.Lock.RLock()
	obj, ok := chk.ObjMap[r.Oid]
	pg.Lock.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such object: %s", r.Oid)
		return
	}

	// Create a directory for a local group if not exist.
	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid
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
	_, err = io.CopyN(r.Out, fChunk, r.Osize)
	if err != nil {
		r.Err = err
		return
	}

	// Complete to read the requested object.
	r.Err = nil
	return
}

func (s *service) readAll(r *repository.Request) {
	// Find and get a logical volume.
	pg, ok := s.pgs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
		return
	}

	pg.Lock.RLock()
	chk, ok := pg.ChunkMap[r.Cid]
	if !ok {
		r.Err = fmt.Errorf("no chunk of such object: %s", r.Oid)
		return
	}
	pg.Lock.RUnlock()

	// Create a directory for a local group if not exist.
	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid
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

	// Complete to read the all contents from the chunk.
	r.Err = nil
	return
}

func (s *service) write(r *repository.Request) {
	// Find and get a logical volume.
	pg, ok := s.pgs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
		return
	}

	pg.Lock.RLock()
	chk, ok := pg.ChunkMap[r.Cid]
	pg.Lock.RUnlock()
	if !ok {
		pg.DiskSched = pg.DiskSched%pg.NumOfPart + 1
		pg.Lock.Lock()
		pg.ChunkMap[r.Cid] = repository.StChunkMap{
			PartID: "part" + strconv.Itoa(int(pg.DiskSched)),
			ObjMap: make(map[string]repository.StObjMap),
		}
		pg.Lock.Unlock()
		chk = pg.ChunkMap[r.Cid]
	}

	// Create a directory for a local group if not exist.
	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid
	_, err := os.Stat(lgDir)
	if os.IsNotExist(err) {
		os.MkdirAll(lgDir, 0775)
	}

	// Open a chunk that objects will be written to.
	fChunk, err := os.OpenFile(lgDir+"/"+r.Cid, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		r.Err = err
		return
	}
	defer fChunk.Close()

	// Obtain a path of the chunk.
	fChunkName := fChunk.Name()

	// Get an information of the chunk.
	fChunkInfo, err := os.Lstat(fChunkName)
	if err != nil {
		r.Err = err
		return
	}

	// Get current length of the chunk.
	fChunkLen := fChunkInfo.Size()

	// If the chunk is newly generated, write a chunk header.
	if fChunkLen == 0 {
		cHeader := make([]byte, 1)

		// ToDo: implement more information of cHeader.
		cHeader[0] = 0x01
		n, err := fChunk.Write(cHeader)

		if n != len(cHeader) {
			r.Err = err
			return
		}
	}
	// Create an object header for requested object.
	oHeader := make([]string, 8)
	oHeader[0] = r.User

	// Check whether the chunk is full or has not enough space to write the object.
	if fChunkLen >= pg.ChunkSize {
		r.Err = fmt.Errorf("chunk full")
		return
	}
	if fChunkLen+int64(len(oHeader))+r.Osize > pg.ChunkSize {
		err = fChunk.Truncate(pg.ChunkSize)
		r.Err = fmt.Errorf("truncated")
		return
	}

	// Write the object header into the chunk.
	n, err := fChunk.WriteString(oHeader[0])

	// ToDo: implement more information of cHeader.
	if n != len(oHeader[0]) {
		r.Err = err
		return
	}

	// Get an information of the chunk.
	fChunkInfo, err = os.Lstat(fChunkName)
	if err != nil {
		r.Err = err
		return
	}

	// Get current length of the chunk.
	fChunkLen = fChunkInfo.Size()

	// Write the object into the chunk if it will not be full.
	_, err = io.CopyN(fChunk, r.In, r.Osize)
	if err != nil {
		r.Err = err
		return
	}

	// Store mapping information between the object and the chunk.
	pg.Lock.Lock()

	chk.ObjMap[r.Oid] = repository.StObjMap{
		Cid:    r.Cid,
		Offset: fChunkLen,
	}

	pg.ObjInfo[r.Oid] = repository.StObjInfo{
		Size: r.Osize,
		MD5:  r.Md5,
	}

	pg.Lock.Unlock()

	// Complete to write the object into the chunk.
	r.Err = nil
	return
}

func (s *service) writeAll(r *repository.Request) {
	// Find and get a logical volume.
	pg, ok := s.pgs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
		return
	}

	// Check if the requested object is in the object map.
	_, ok = pg.ChunkMap[r.Cid].ObjMap[r.Oid]
	if ok {
		r.Err = fmt.Errorf("same name of the chunk is existed: %s", r.Oid)
		return
	}

	// Create a directory for a local group if not exist.
	lgDir := pg.MntPoint + "/" + r.LocGid
	_, err := os.Stat(lgDir)
	if os.IsNotExist(err) {
		os.MkdirAll(lgDir, 0775)
	}

	// Open a chunk that objects will be written to.
	fChunk, err := os.OpenFile(lgDir+"/"+r.Cid, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		r.Err = err
		return
	}
	defer fChunk.Close()

	// Check whether the chunk is full or has not enough space to write the object.
	if r.Osize != pg.ChunkSize {
		r.Err = fmt.Errorf("write all must writes same length with the chunk size")
		return
	}

	// Write the object into the chunk if it will not be full.
	_, err = io.CopyN(fChunk, r.In, r.Osize)
	if err != nil {
		r.Err = err
		return
	}

	// Store mapping information between the object and the chunk.
	pg.Lock.Lock()
	pg.Lock.Unlock()

	// Complete to write the object into the chunk.
	r.Err = nil
	return
}

func (s *service) delete(r *repository.Request) {
	// Find and get a logical volume.
	pg, ok := s.pgs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
		return
	}

	// Check if the requested object is in the object map.
	_, ok = pg.ChunkMap[r.Cid].ObjMap[r.Oid]
	if !ok {
		r.Err = fmt.Errorf("no such object: %s", r.Oid)
		return
	}

	// Delete the object from the map.
	pg.Lock.Lock()
	delete(pg.ChunkMap[r.Cid].ObjMap, r.Oid)
	pg.Lock.Unlock()

	// Complete to delete the object from the map.
	r.Err = nil
	return
}

// NewClusterRepository returns a new part store inteface in a view of cluster domain.
func NewClusterRepository(store repository.Service) cluster.Repository {
	return store
}

// NewObjectRepository returns a new part store inteface in a view of object domain.
func NewObjectRepository(store repository.Service) object.Repository {
	return store
}
