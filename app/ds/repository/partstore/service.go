package partstore

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chanyoung/nil/app/ds/repository"
	"github.com/chanyoung/nil/app/ds/usecase/cluster"
	"github.com/chanyoung/nil/app/ds/usecase/gencoding"
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
	spinUpTicker := time.NewTicker(30 * time.Second)

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
		case <-spinUpTicker.C:
			go s.MigrateData()
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

// MigrateData moves data from service to archive archive.
func (s *service) MigrateData() error {
	for _, pg := range s.pgs {
		for key, value := range pg.ChunkMap {
			// After check type of each chunks, migrate only parity chunk in hot storage into cold partition.
			PartID := value.PartID
			if value.ChunkInfo.Type == "Data" {
				continue
			}
			if strings.HasPrefix(PartID, "hot_") {
				if !strings.HasPrefix(key, "G_") {
					continue
				}

				fChunkSrc, err := os.OpenFile(pg.MntPoint+"/"+PartID+"/"+value.ChunkInfo.LocGid+"/"+key, os.O_RDWR, 0775)
				if err != nil {
					return err
				}

				// Scheduling for cold partitions.
				pg.SubPartGroup.Cold.DiskSched = pg.SubPartGroup.Cold.DiskSched%pg.SubPartGroup.Cold.NumOfPart + 1
				DiskSched := pg.SubPartGroup.Cold.DiskSched
				DestPartID := "cold_part" + strconv.Itoa(int(DiskSched))

				// Create a directory for a local group if not exist.
				lgDir := pg.MntPoint + "/" + DestPartID + "/" + value.ChunkInfo.LocGid
				_, err = os.Stat(lgDir)
				if os.IsNotExist(err) {
					os.MkdirAll(lgDir, 0775)
				}

				fChunkDest, err := os.OpenFile(lgDir+"/"+key, os.O_CREATE|os.O_WRONLY, 0775)
				if err != nil {
					return err
				}

				// Copy the chunk in the hot storage to the cold storage.
				_, err = io.Copy(fChunkDest, fChunkSrc)

				// Update ChunkMap for complitely migrated chunk.
				value.PartID = DestPartID
				pg.ChunkMap[key] = value

				err = fChunkSrc.Close()
				if err != nil {
					return err
				}

				// Remove the chunk from the hot storage.
				err = os.Remove(pg.MntPoint + "/" + PartID + "/" + value.ChunkInfo.LocGid + "/" + key)
				if err != nil {
					return err
				}

				fChunkDest.Close()
			}
		}
	}

	return nil
}

func (s *service) GetObjectSize(pgID, objID string) (int64, bool) {
	pg, ok := s.pgs[pgID]
	if ok == false {
		return 0, false
	}

	pg.Lock.Obj.RLock()
	obj, ok := pg.ObjMap[objID]
	pg.Lock.Obj.RUnlock()
	if ok == false {
		return 0, false
	}

	objSize := obj.ObjInfo.Size - 140

	return objSize, true
}

func (s *service) GetObjectMD5(pgID, objID string) (string, bool) {
	pg, ok := s.pgs[pgID]
	if ok == false {
		return "", false
	}

	pg.Lock.Obj.RLock()
	obj, ok := pg.ObjMap[objID]
	pg.Lock.Obj.RUnlock()
	if ok == false {
		return "", false
	}

	return obj.ObjInfo.MD5, true
}

func (s *service) GetChunkHeaderSize() int64 {
	// TODO: fill the method
	return 135
}

func (s *service) GetObjectHeaderSize() int64 {
	// TODO: fill the method
	return 140
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
	case repository.DeleteReal:
		s.deleteReal(r)
	}
}

func (s *service) read(r *repository.Request) {
	// Find and get the requested logical volume.
	pg, ok := s.pgs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
		return
	}

	pg.Lock.Obj.RLock()
	obj, ok := pg.ObjMap[r.Oid]
	pg.Lock.Obj.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such object: %s", r.Oid)
		return
	}

	// Find and get the requested object.
	pg.Lock.Chk.RLock()
	chk, ok := pg.ChunkMap[obj.Cid]
	pg.Lock.Chk.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no chunk of such object: %s", r.Oid)
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

	oHeader := new(repository.ObjHeader)
	dec := gob.NewDecoder(fChunk)
	err = dec.Decode(oHeader)

	//fmt.Println(oHeader.Name, oHeader.Size, oHeader.Offset)

	_, err = fChunk.Seek(oHeader.Offset, os.SEEK_SET)
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
	pg.Lock.Chk.RLock()
	chk, ok := pg.ChunkMap[r.Cid]
	if !ok {
		r.Err = fmt.Errorf("no chunk of such object: %s", r.Oid)
		return
	}
	pg.Lock.Chk.RUnlock()

	// Create a directory for a local group if not exist.
	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid

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

	pg.Lock.Obj.RLock()
	_, ok = pg.ObjMap[r.Oid]
	pg.Lock.Obj.RUnlock()
	if ok {
		r.Err = fmt.Errorf("object is already existed: %s", r.Oid)
		return
	}

	pg.Lock.Chk.RLock()
	chk, ok := pg.ChunkMap[r.Cid]
	pg.Lock.Chk.RUnlock()
	if !ok {
		pg.SubPartGroup.Hot.DiskSched = pg.SubPartGroup.Hot.DiskSched%pg.SubPartGroup.Hot.NumOfPart + 1
		DiskSched := pg.SubPartGroup.Hot.DiskSched
		PartID := "hot_part" + strconv.Itoa(int(DiskSched))

		pg.Lock.Chk.Lock()
		pg.ChunkMap[r.Cid] = repository.ChunkMap{
			PartID: PartID,
			ChunkInfo: repository.ChunkInfo{
				Type:   r.Type,
				LocGid: r.LocGid,
			},
		}
		pg.Lock.Chk.Unlock()

		pg.Lock.Chk.RLock()
		chk = pg.ChunkMap[r.Cid]
		pg.Lock.Chk.RUnlock()
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
		cHeader := repository.ChunkHeader{
			Magic:   [4]byte{0x7f, 'c', 'h', 'k'},
			Type:    make([]byte, 1),
			State:   [1]byte{'P'},
			Encoded: false,
		}

		cHeader.Type = append(cHeader.Type, 'D')

		b := new(bytes.Buffer)
		enc := gob.NewEncoder(b)
		err := enc.Encode(cHeader)
		if err != nil {
			r.Err = err
			return
		}

		n, err := fChunk.Write(b.Bytes())
		//fmt.Printf("chunk written: %d\n", n)

		if n != b.Len() {
			r.Err = err
			return
		}
	}
	// Get an information of the chunk.
	fChunkInfo, err = os.Lstat(fChunkName)
	if err != nil {
		r.Err = err
		return
	}

	// Get current length of the chunk.
	fChunkLen = fChunkInfo.Size()

	// Create an object header for requested object.
	oHeader := repository.ObjHeader{
		Magic:  [4]byte{0x7f, 'o', 'b', 'j'},
		Name:   make([]byte, 32),
		Size:   r.Osize,
		Offset: fChunkLen + 140,
	}

	oHeader.Name = append([]byte(r.Oid))
	//fmt.Println("len(r.Oid) : ", len(r.Oid))

	for i := len(r.Oid); i <= 32; i++ {
		oHeader.Name = append(oHeader.Name, byte(0))
	}

	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err = enc.Encode(&oHeader)
	if err != nil {
		r.Err = err
		return
	}

	//fmt.Println(b.Len())

	// Check whether the chunk is full or has not enough space to write the object.
	if fChunkLen >= pg.ChunkSize {
		r.Err = fmt.Errorf("chunk full")
		return
	}
	if fChunkLen+int64(b.Len())+r.Osize > pg.ChunkSize {
		err = fChunk.Truncate(pg.ChunkSize)
		r.Err = fmt.Errorf("truncated")
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

	// Write the object header into the chunk.
	n, err := fChunk.Write(b.Bytes())
	//fmt.Printf("buf len: %d, object written: %d\n", b.Len(), n)

	// ToDo: implement more information of cHeader.
	if n != b.Len() {
		r.Err = err
		return
	}

	// Write the object into the chunk if it will not be full.
	_, err = io.CopyN(fChunk, r.In, r.Osize)
	if err != nil {
		r.Err = err
		return
	}

	// Store mapping information between the object and the chunk.
	pg.Lock.Obj.Lock()

	pg.ObjMap[r.Oid] = repository.ObjMap{
		Cid:    r.Cid,
		Offset: fChunkLen,
		ObjInfo: repository.ObjInfo{
			Size: r.Osize + int64(n),
			MD5:  r.Md5,
		},
	}

	pg.Lock.Obj.Unlock()

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

	pg.Lock.Chk.RLock()
	chk, ok := pg.ChunkMap[r.Cid]
	pg.Lock.Chk.RUnlock()
	if ok {
		r.Err = fmt.Errorf("chunk is already exists: %s", r.Cid)
		return
	}

	pg.SubPartGroup.Hot.DiskSched = pg.SubPartGroup.Hot.DiskSched%pg.SubPartGroup.Hot.NumOfPart + 1
	DiskSched := pg.SubPartGroup.Hot.DiskSched
	PartID := "hot_part" + strconv.Itoa(int(DiskSched))

	pg.Lock.Chk.Lock()
	pg.ChunkMap[r.Cid] = repository.ChunkMap{
		PartID: PartID,
		ChunkInfo: repository.ChunkInfo{
			Type:   r.Type,
			LocGid: r.LocGid,
		},
	}
	pg.Lock.Chk.Unlock()

	pg.Lock.Chk.RLock()
	chk = pg.ChunkMap[r.Cid]
	pg.Lock.Chk.RUnlock()

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

	// Write the object into the chunk if it will not be full.
	n, err := io.CopyN(fChunk, r.In, pg.ChunkSize)
	if err != nil || n != pg.ChunkSize {
		r.Err = err
		return
	}

	// Completely write the chunk.
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
	pg.Lock.Obj.RLock()
	_, ok = pg.ObjMap[r.Oid]
	pg.Lock.Obj.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such object: %s", r.Oid)
		return
	}

	// Delete the object from the map.
	pg.Lock.Obj.Lock()
	delete(pg.ObjMap, r.Oid)
	pg.Lock.Obj.Unlock()

	// Complete to delete the object from the map.
	r.Err = nil
	return
}

func (s *service) deleteReal(r *repository.Request) {
	// Find and get a logical volume.
	pg, ok := s.pgs[r.Vol]
	if !ok {
		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
		return
	}

	// Check if the requested object is in the object map.
	pg.Lock.Chk.RLock()
	chk, ok := pg.ChunkMap[r.Cid]
	pg.Lock.Chk.RUnlock()

	// Remove chunk
	if ok {
		lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid

		// Remove all metadata of chunk
		pg.Lock.Obj.Lock()
		for key, value := range pg.ObjMap {
			if value.Cid == r.Cid {
				delete(pg.ObjMap, key)
			}
		}
		pg.Lock.Obj.Unlock()

		err := os.Remove(lgDir + "/" + r.Cid)
		if err != nil {
			r.Err = fmt.Errorf("no such chunk: %s", r.Cid)
			return
		}

		pg.Lock.Chk.Lock()
		delete(pg.ChunkMap, r.Cid)
		pg.Lock.Chk.Unlock()

		r.Err = nil
		return
	}

	// Remove object
	pg.Lock.Obj.RLock()
	obj, ok := pg.ObjMap[r.Oid]
	pg.Lock.Obj.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such object: %s", r.Oid)
		return
	}

	pg.Lock.Obj.RLock()
	chk, ok = pg.ChunkMap[obj.Cid]
	pg.Lock.Obj.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no chunk including such object: %s", r.Cid)
		return
	}

	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid
	fChunk, err := os.OpenFile(lgDir+"/"+obj.Cid, os.O_RDWR, 0775)
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

	if obj.Offset+obj.ObjInfo.Size != fChunkLen {
		r.Err = fmt.Errorf("can remove only a last object of a chunk")
		return
	}
	fChunk.Seek(obj.Offset, io.SeekStart)
	fChunk.Truncate(fChunkLen - (fChunkLen - obj.Offset))

	pg.Lock.Obj.Lock()
	delete(pg.ObjMap, r.Oid)
	pg.Lock.Obj.Unlock()

	r.Err = nil
	return
}

// RenameChunk renames oldpath to newpath of chunk.
func (s *service) RenameChunk(src string, dest string, Vol string, LocGid string) error {
	if Vol == "" || src == "" || dest == "" || LocGid == "" {
		err := fmt.Errorf("invalid arguments: %s, %s, %s, %s", src, dest, Vol, LocGid)
		return err
	}

	pg, ok := s.pgs[Vol]
	if !ok {
		err := fmt.Errorf("no such partition group: %s", Vol)
		return err
	}

	pg.Lock.Chk.RLock()
	chk, ok := pg.ChunkMap[src]
	pg.Lock.Chk.RUnlock()
	if !ok {
		err := fmt.Errorf("no such chunk: %s", src)
		return err
	}

	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + LocGid
	err := os.Rename(lgDir+"/"+src, lgDir+"/"+dest)
	if err != nil {
		return err
	}

	pg.Lock.Obj.Lock()
	for key, value := range pg.ObjMap {
		if value.Cid == src {
			value.Cid = dest
			pg.ObjMap[key] = value
		}
	}
	pg.Lock.Obj.Unlock()

	pg.Lock.Chk.Lock()
	pg.ChunkMap[dest] = repository.ChunkMap{
		PartID: chk.PartID,
		ChunkInfo: repository.ChunkInfo{
			Type:   chk.ChunkInfo.Type,
			LocGid: chk.ChunkInfo.LocGid,
		},
	}
	delete(pg.ChunkMap, src)
	pg.Lock.Chk.Unlock()

	return nil
}

func (s *service) CountNonCodedChunk(Vol string, LocGid string) (int, error) {
	if Vol == "" || LocGid == "" {
		err := fmt.Errorf("invalid arguements: %s, %s", Vol, LocGid)
		return -1, err
	}

	pg, ok := s.pgs[Vol]
	if !ok {
		err := fmt.Errorf("no such partition group: %s", Vol)
		return -1, err
	}

	dir := pg.MntPoint
	encPath := "/" + LocGid + "/L_"
	count := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			err := fmt.Errorf("prevent panic by handling failure accessing a path %q: %v", dir, err)
			return err
		}
		ok, err := regexp.MatchString(encPath, path)
		if ok {
			count++
		}
		return nil
	})

	if err != nil {
		return -1, err
	}

	return count, nil
}

func (s *service) GetNonCodedChunk(Vol string, LocGid string) (string, error) {
	if Vol == "" || LocGid == "" {
		err := fmt.Errorf("invalid arguements: %s, %s", Vol, LocGid)
		return "", err
	}

	pg, ok := s.pgs[Vol]
	if !ok {
		err := fmt.Errorf("no such partition group: %s", Vol)
		return "", err
	}

	dir := pg.MntPoint
	encPath := "/" + LocGid + "/L_"
	var cid string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if cid != "" {
				return nil
			}
		}
		if err != nil {
			err := fmt.Errorf("prevent panic by handling failure accessing a path %q: %v", dir, err)
			return err
		}
		ok, err := regexp.MatchString(encPath, path)
		if ok {
			cid = filepath.Base(path)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return cid, nil
}

// NewClusterRepository returns a new part store inteface in a view of cluster domain.
func NewClusterRepository(store repository.Service) cluster.Repository {
	return store
}

// NewObjectRepository returns a new part store inteface in a view of object domain.
func NewObjectRepository(store repository.Service) object.Repository {
	return store
}

// NewGencodingRepository returns a new part store inteface in a view of gencoding domain.
func NewGencodingRepository(store repository.Service) gencoding.Repository {
	return store
}
