package lvstore

// // service is the backend store service.
// type service struct {
// 	lvs          map[string]*lv
// 	basePath     string
// 	requestQueue queue
// 	pushCh       chan interface{}
// }

// // NewService returns a new backend store service.
// func NewService(basePath string) repository.Service {
// 	return &service{
// 		basePath: basePath,
// 		lvs:      map[string]*lv{},
// 		pushCh:   make(chan interface{}, 1),
// 	}
// }

// // newService returns a new backend store service.
// // This is only for unit test. Do not use in real service.
// func newService(basePath string) *service {
// 	return &service{
// 		basePath: basePath,
// 		lvs:      map[string]*lv{},
// 		pushCh:   make(chan interface{}, 1),
// 	}
// }

// // Run starts to serve backend store service.
// func (s *service) Run() {
// 	checkTicker := time.NewTicker(100 * time.Millisecond)
// 	spinUpTicker := time.NewTicker(10 * time.Second)

// 	// TODO: change to do not polling.
// 	for {
// 		select {
// 		case <-s.pushCh:
// 			if c := s.requestQueue.pop(); c != nil {
// 				go s.handleCall(c)
// 			}
// 		case <-checkTicker.C:
// 			if c := s.requestQueue.pop(); c != nil {
// 				go s.handleCall(c)
// 			}
// 		case <-spinUpTicker.C:
// 			go s.MigrateData()
// 		}
// 	}
// }

// // Stop supports graceful stop of backend store service.
// func (s *service) Stop() {
// 	// TODO: graceful stop.
// 	// Tracking all jobs and wait them until finished.

// 	// Deletes all volumes in the store.
// 	for _, lv := range s.lvs {
// 		lv.Umount()
// 	}
// }

// // Push pushes an io request into the scheduling queue.
// func (s *service) Push(r *repository.Request) error {
// 	if err := r.Verify(); err != nil {
// 		return err
// 	}

// 	r.Wg.Add(1)
// 	s.requestQueue.push(r)
// 	s.pushCh <- nil

// 	return nil
// }

// // AddVolume adds a volume into the lv map.
// func (s *service) AddVolume(v *repository.Vol) error {
// 	if _, ok := s.lvs[v.Name]; ok {
// 		return fmt.Errorf("Volume name %s already exists", v.Name)
// 	}

// 	if err := v.Mount(); err != nil {
// 		return err
// 	}

// 	// Update filesystem stats.
// 	if err := v.UpdateStatFs(); err != nil {
// 		return err
// 	}

// 	// TODO: Set the disk speed.
// 	v.SetSpeed()

// 	s.lvs[v.Name] = &lv{
// 		Vol: v,
// 	}

// 	// Set volume has running state.
// 	v.Status = repository.Active

// 	return nil
// }

// // MigrateData moves data from service to archive archive.
// func (s *service) MigrateData() error {
// 	for _, lv := range s.lvs {
// 		for key, value := range lv.ChunkMap {
// 			// After check type of each chunks, migrate only parity chunk in hot storage into cold partition.
// 			if value.ChunkInfo.Type == "Data" {
// 				continue
// 			}
// 			PartID := value.PartID
// 			if strings.HasPrefix(PartID, "hot_") {
// 				if !strings.HasPrefix(key, "/G_") {
// 					continue
// 				}
// 				fChunkSrc, err := os.OpenFile(lv.MntPoint+"/"+PartID+"/"+value.ChunkInfo.LocGid+"/"+key, os.O_RDWR, 0775)
// 				if err != nil {
// 					return err
// 				}

// 				// Scheduling for cold partitions.
// 				lv.SubPartGroup.Cold.DiskSched = lv.SubPartGroup.Cold.DiskSched%lv.SubPartGroup.Cold.NumOfPart + 1
// 				DiskSched := lv.SubPartGroup.Cold.DiskSched
// 				DestPartID := "cold_part" + strconv.Itoa(int(DiskSched))

// 				// Create a directory for a local group if not exist.
// 				lgDir := lv.MntPoint + "/" + DestPartID + "/" + value.ChunkInfo.LocGid
// 				_, err = os.Stat(lgDir)
// 				if os.IsNotExist(err) {
// 					os.MkdirAll(lgDir, 0775)
// 				}

// 				fChunkDest, err := os.OpenFile(lgDir+"/"+key, os.O_CREATE|os.O_WRONLY, 0775)
// 				if err != nil {
// 					return err
// 				}

// 				// Copy the chunk in the hot storage to the cold storage.
// 				_, err = io.Copy(fChunkDest, fChunkSrc)

// 				// Update ChunkMap for complitely migrated chunk.
// 				value.PartID = DestPartID
// 				lv.ChunkMap[key] = value

// 				err = fChunkSrc.Close()
// 				if err != nil {
// 					return err
// 				}

// 				// Remove the chunk from the hot storage.
// 				err = os.Remove(lv.MntPoint + "/" + PartID + "/" + value.ChunkInfo.LocGid + "/" + key)
// 				if err != nil {
// 					return err
// 				}

// 				fChunkDest.Close()
// 			}
// 		}
// 	}

// 	return nil
// }

// func (s *service) ChunkExist(pgID, chkID string) bool {
// 	return false
// }

// func (s *service) GetObjectSize(lvID, objID string) (int64, bool) {
// 	lv, ok := s.lvs[lvID]
// 	if ok == false {
// 		return 0, false
// 	}

// 	lv.Lock.Obj.RLock()
// 	obj, ok := lv.ObjMap[objID]
// 	lv.Lock.Obj.RUnlock()
// 	if ok == false {
// 		return 0, false
// 	}

// 	return obj.ObjInfo.Size, true
// }

// func (s *service) GetObjectMD5(lvID, objID string) (string, bool) {
// 	lv, ok := s.lvs[lvID]
// 	if ok == false {
// 		return "", false
// 	}

// 	lv.Lock.Obj.RLock()
// 	obj, ok := lv.ObjMap[objID]
// 	lv.Lock.Obj.RUnlock()
// 	if ok == false {
// 		return "", false
// 	}

// 	return obj.ObjInfo.MD5, true
// }

// func (s *service) GetChunkHeaderSize() int64 {
// 	// TODO: fill the method
// 	return 8
// }

// func (s *service) GetObjectHeaderSize() int64 {
// 	// TODO: fill the method
// 	return 100
// }

// func (s *service) handleCall(r *repository.Request) {
// 	defer r.Wg.Done()

// 	switch r.Op {
// 	case repository.Read:
// 		s.read(r)
// 	case repository.Write:
// 		s.write(r)
// 	case repository.WriteAll:
// 		s.writeAll(r)
// 	case repository.Delete:
// 		s.delete(r)
// 	case repository.ReadAll:
// 		s.readAll(r)
// 	case repository.DeleteReal:
// 		s.deleteReal(r)
// 	}
// }

// func (s *service) read(r *repository.Request) {
// 	// Find and get the requested logical volume.
// 	lv, ok := s.lvs[r.Vol]
// 	if !ok {
// 		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
// 		return
// 	}

// 	// Find and get the requested object.
// 	lv.Lock.Obj.RLock()
// 	obj, ok := lv.ObjMap[r.Oid]
// 	lv.Lock.Obj.RUnlock()
// 	if !ok {
// 		r.Err = fmt.Errorf("no such object: %s", r.Oid)
// 		return
// 	}

// 	// Open a chunk requested by a client.
// 	lgDir := lv.MntPoint + "/" + r.LocGid
// 	fChunk, err := os.Open(lgDir + "/" + obj.Cid)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}
// 	defer fChunk.Close()

// 	// Seek offset beginning of the requested object in the chunk.
// 	_, err = fChunk.Seek(obj.Offset, os.SEEK_SET)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	oHeader := new(repository.ObjHeader)
// 	err = binary.Read(fChunk, binary.LittleEndian, oHeader)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Read contents of the requested object from the chunk.
// 	_, err = fChunk.Seek(oHeader.Offset, os.SEEK_SET)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	_, err = io.CopyN(r.Out, fChunk, r.Osize)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Complete to read the requested object.
// 	r.Err = nil
// 	return
// }

// func (s *service) readAll(r *repository.Request) {
// 	// Find and get a logical volume.
// 	lv, ok := s.lvs[r.Vol]
// 	if !ok {
// 		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
// 		return
// 	}

// 	// Open a chunk requested by a client.
// 	lgDir := lv.MntPoint + "/" + r.LocGid
// 	fChunk, err := os.Open(lgDir + "/" + r.Cid)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}
// 	defer fChunk.Close()

// 	// Read all contents from the chunk to a writer stream.
// 	_, err = io.Copy(r.Out, fChunk)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Complete to read the all contents from the chunk.
// 	r.Err = nil
// 	return
// }

// func (s *service) write(r *repository.Request) {
// 	// Find and get a logical volume.
// 	lv, ok := s.lvs[r.Vol]
// 	if !ok {
// 		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
// 		return
// 	}

// 	lv.Lock.Chk.RLock()
// 	_, ok = lv.ChunkMap[r.Cid]
// 	lv.Lock.Chk.RUnlock()

// 	if !ok {
// 		lv.Lock.Chk.Lock()
// 		lv.ChunkMap[r.Cid] = repository.ChunkMap{
// 			ChunkInfo: repository.ChunkInfo{
// 				Type:   r.Type,
// 				LocGid: r.LocGid,
// 			},
// 		}
// 		lv.Lock.Chk.Unlock()
// 	}

// 	// Create a directory for a local group if not exist.
// 	lgDir := lv.MntPoint + "/" + r.LocGid
// 	_, err := os.Stat(lgDir)
// 	if os.IsNotExist(err) {
// 		os.MkdirAll(lgDir, 0775)
// 	}

// 	// Open a chunk that objects will be written to.
// 	fChunk, err := os.OpenFile(lgDir+"/"+r.Cid, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}
// 	defer fChunk.Close()

// 	// Obtain a path of the chunk.
// 	fChunkName := fChunk.Name()

// 	// Get an information of the chunk.
// 	fChunkInfo, err := os.Lstat(fChunkName)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Get current length of the chunk.
// 	fChunkLen := fChunkInfo.Size()

// 	// If the chunk is newly generated, write a chunk header.
// 	if fChunkLen == 0 {
// 		cHeader := repository.ChunkHeader{
// 			Magic:   [4]byte{0x7f, 'c', 'h', 'k'},
// 			Type:    [1]byte{},
// 			State:   [1]byte{'P'},
// 			Encoded: false,
// 		}

// 		cHeader.Type[0] = 'D'

// 		b := new(bytes.Buffer)
// 		bufio.NewWriter(b)
// 		err := binary.Write(b, binary.LittleEndian, cHeader)
// 		if err != nil {
// 			r.Err = err
// 			return
// 		}

// 		n, err := fChunk.Write(b.Bytes())
// 		//fmt.Printf("chunk written: %d\n", n)

// 		fChunkLen = fChunkLen + int64(n)

// 		if n != b.Len() {
// 			r.Err = err
// 			return
// 		}
// 	}
// 	// Get an information of the chunk.
// 	fChunkInfo, err = os.Lstat(fChunkName)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Get current length of the chunk.
// 	fChunkLen = fChunkInfo.Size()

// 	// Create an object header for requested object.
// 	oHeader := repository.ObjHeader{
// 		Magic:  [4]byte{0x7f, 'o', 'b', 'j'},
// 		Name:   [48]byte{},
// 		MD5:    [32]byte{},
// 		Size:   r.Osize,
// 		Offset: fChunkLen + s.GetObjectHeaderSize(),
// 	}

// 	for i := 0; i < len(r.Oid); i++ {
// 		oHeader.Name[i] = r.Oid[i]
// 	}
// 	//fmt.Println("len(r.Oid) : ", len(r.Oid))

// 	for i := len(r.Oid); i < len(oHeader.Name); i++ {
// 		oHeader.Name[i] = '0'
// 	}

// 	for i := 0; i < len(r.Md5); i++ {
// 		oHeader.MD5[i] = r.Md5[i]
// 	}

// 	b := new(bytes.Buffer)
// 	bufio.NewWriter(b)
// 	err = binary.Write(b, binary.LittleEndian, oHeader)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Check whether the chunk is full or has not enough space to write the object.
// 	if fChunkLen >= lv.ChunkSize {
// 		r.Err = fmt.Errorf("chunk full")
// 		return
// 	}
// 	if fChunkLen+int64(b.Len())+r.Osize > lv.ChunkSize {
// 		err = fChunk.Truncate(lv.ChunkSize)
// 		r.Err = fmt.Errorf("truncated")
// 		return
// 	}

// 	// Get an information of the chunk.
// 	fChunkInfo, err = os.Lstat(fChunkName)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Get current length of the chunk.
// 	fChunkLen = fChunkInfo.Size()

// 	// Write the object header into the chunk.
// 	n, err := fChunk.Write(b.Bytes())

// 	// ToDo: implement more information of cHeader.
// 	if n != b.Len() {
// 		r.Err = err
// 		return
// 	}

// 	// Write the object into the chunk if it will not be full.
// 	_, err = io.CopyN(fChunk, r.In, r.Osize)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}

// 	// Store mapping information between the object and the chunk.
// 	lv.Lock.Obj.Lock()

// 	lv.ObjMap[r.Oid] = repository.ObjMap{
// 		Cid:    r.Cid,
// 		Offset: fChunkLen,
// 		ObjInfo: repository.ObjInfo{
// 			Size: r.Osize,
// 			MD5:  r.Md5,
// 		},
// 	}

// 	lv.Lock.Obj.Unlock()

// 	// Complete to write the object into the chunk.
// 	r.Err = nil
// 	return
// }

// func (s *service) writeAll(r *repository.Request) {
// 	// Find and get a logical volume.
// 	lv, ok := s.lvs[r.Vol]
// 	if !ok {
// 		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
// 		return
// 	}

// 	// Check if the requested object is in the object map.
// 	lv.Lock.Obj.RLock()
// 	_, ok = lv.ObjMap[r.Oid]
// 	lv.Lock.Obj.RUnlock()
// 	if ok {
// 		r.Err = fmt.Errorf("chunk is already existed: %s", r.Oid)
// 		return
// 	}

// 	lv.Lock.Chk.Lock()
// 	lv.ChunkMap[r.Cid] = repository.ChunkMap{
// 		ChunkInfo: repository.ChunkInfo{
// 			Type:   r.Type,
// 			LocGid: r.LocGid,
// 		},
// 	}
// 	lv.Lock.Chk.Unlock()

// 	// Create a directory for a local group if not exist.
// 	lgDir := lv.MntPoint + "/" + r.LocGid
// 	_, err := os.Stat(lgDir)
// 	if os.IsNotExist(err) {
// 		os.MkdirAll(lgDir, 0775)
// 	}

// 	// Open a chunk that objects will be written to.
// 	fChunk, err := os.OpenFile(lgDir+"/"+r.Cid, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
// 	if err != nil {
// 		r.Err = err
// 		return
// 	}
// 	defer fChunk.Close()

// 	// Check whether the chunk is full or has not enough space to write the object.
// 	if r.Osize != lv.ChunkSize {
// 		r.Err = fmt.Errorf("write all must writes same length with the chunk size")
// 		return
// 	}

// 	// Write the object into the chunk if it will not be full.
// 	n, err := io.CopyN(fChunk, r.In, lv.ChunkSize)
// 	if err != nil || n != lv.ChunkSize {
// 		r.Err = err
// 		return
// 	}

// 	// Complete to write the object into the chunk.
// 	r.Err = nil
// 	return
// }

// func (s *service) delete(r *repository.Request) {
// 	// Find and get a logical volume.
// 	lv, ok := s.lvs[r.Vol]
// 	if !ok {
// 		r.Err = fmt.Errorf("no such lv: %s", r.Vol)
// 		return
// 	}

// 	// Check if the requested object is in the object map.
// 	lv.Lock.Obj.RLock()
// 	_, ok = lv.ObjMap[r.Oid]
// 	lv.Lock.Obj.RUnlock()
// 	if !ok {
// 		r.Err = fmt.Errorf("no such object: %s", r.Oid)
// 		return
// 	}

// 	// Delete the object from the map.
// 	lv.Lock.Obj.Lock()
// 	delete(lv.ObjMap, r.Oid)
// 	lv.Lock.Obj.Unlock()

// 	// Complete to delete the object from the map.
// 	r.Err = nil
// 	return
// }

// func (s *service) deleteReal(r *repository.Request) {
// 	// Find and get a logical volume.
// 	lv, ok := s.lvs[r.Vol]
// 	if !ok {
// 		r.Err = fmt.Errorf("no such partition group: %s", r.Vol)
// 		return
// 	}

// 	lv.Lock.Obj.RLock()
// 	obj, ok := lv.ObjMap[r.Oid]
// 	lv.Lock.Obj.RUnlock()
// 	if ok {
// 		lgDir := lv.MntPoint + "/" + r.LocGid

// 		fChunk, err := os.OpenFile(lgDir+"/"+obj.Cid, os.O_RDWR, 0775)
// 		if err != nil {
// 			r.Err = err
// 			return
// 		}
// 		defer fChunk.Close()

// 		// Obtain a path of the chunk.
// 		fChunkName := fChunk.Name()

// 		// Get an information of the chunk.
// 		fChunkInfo, err := os.Lstat(fChunkName)
// 		if err != nil {
// 			r.Err = err
// 			return
// 		}

// 		// Get current length of the chunk.
// 		fChunkLen := fChunkInfo.Size()

// 		if obj.Offset+obj.ObjInfo.Size+s.GetObjectHeaderSize() != fChunkLen {
// 			r.Err = fmt.Errorf("can remove only a last object of a chunk")
// 			return
// 		}
// 		fChunk.Seek(obj.Offset, io.SeekStart)
// 		fChunk.Truncate(fChunkLen - (fChunkLen - obj.Offset))

// 		lv.Lock.Obj.Lock()
// 		delete(lv.ObjMap, r.Oid)
// 		lv.Lock.Obj.Unlock()

// 		r.Err = nil
// 		return
// 	}

// 	lgDir := lv.MntPoint + "/" + r.LocGid
// 	chk := lgDir + "/" + r.Cid

// 	lv.Lock.Obj.Lock()
// 	// Remove all metadata of chunk
// 	for key, value := range lv.ObjMap {
// 		if value.Cid == r.Cid {
// 			delete(lv.ObjMap, key)
// 		}
// 	}
// 	lv.Lock.Obj.Unlock()

// 	err := os.Remove(chk)
// 	if err != nil {
// 		r.Err = fmt.Errorf("no such chunk: %s", r.Cid)
// 		return
// 	}

// 	lv.Lock.Chk.Lock()
// 	delete(lv.ChunkMap, r.Cid)
// 	lv.Lock.Chk.Unlock()

// 	r.Err = nil
// 	return

// }

// // renameChunk renames oldpath to newpath of chunk.
// func (s *service) RenameChunk(src string, dest string, Vol string, LocGid string) error {
// 	if Vol == "" || src == "" || dest == "" || LocGid == "" {
// 		err := fmt.Errorf("invalid arguments: %s, %s, %s, %s", src, dest, Vol, LocGid)
// 		return err
// 	}

// 	lv, ok := s.lvs[Vol]
// 	if !ok {
// 		err := fmt.Errorf("no such partition group: %s", Vol)
// 		return err
// 	}

// 	lv.Lock.Chk.RLock()
// 	chk, ok := lv.ChunkMap[src]
// 	lv.Lock.Chk.RUnlock()
// 	if !ok {
// 		err := fmt.Errorf("no such chunk: %s", src)
// 		return err
// 	}

// 	lgDir := lv.MntPoint + "/" + chk.PartID + "/" + LocGid
// 	err := os.Rename(lgDir+"/"+src, lgDir+"/"+dest)
// 	if err != nil {
// 		return err
// 	}

// 	lv.Lock.Obj.Lock()
// 	for key, value := range lv.ObjMap {
// 		if value.Cid == src {
// 			value.Cid = dest
// 			lv.ObjMap[key] = value
// 		}
// 	}
// 	lv.Lock.Obj.Unlock()

// 	lv.Lock.Chk.Lock()
// 	lv.ChunkMap[dest] = repository.ChunkMap{
// 		ChunkInfo: repository.ChunkInfo{
// 			Type:   chk.ChunkInfo.Type,
// 			LocGid: chk.ChunkInfo.LocGid,
// 		},
// 	}
// 	delete(lv.ChunkMap, src)
// 	lv.Lock.Chk.Unlock()

// 	return nil
// }

// func (s *service) CountNonCodedChunk(Vol string, LocGid string) (int, error) {
// 	if Vol == "" || LocGid == "" {
// 		err := fmt.Errorf("invalid arguements: %s, %s", Vol, LocGid)
// 		return -1, err
// 	}

// 	lv, ok := s.lvs[Vol]
// 	if !ok {
// 		err := fmt.Errorf("no such partition group: %s", Vol)
// 		return -1, err
// 	}

// 	dir := lv.MntPoint

// 	count := 0

// 	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			err := fmt.Errorf("prevent panic by handling failure accessing a path %q: %v", dir, err)
// 			return err
// 		}
// 		ok, err := regexp.MatchString("/"+LocGid+"/L_", path)
// 		if ok {
// 			count++
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		return -1, err
// 	}

// 	return count, nil
// }

// func (s *service) GetNonCodedChunk(Vol string, LocGid string) (string, error) {
// 	if Vol == "" || LocGid == "" {
// 		err := fmt.Errorf("invalid arguements: %s, %s", Vol, LocGid)
// 		return "", err
// 	}

// 	lv, ok := s.lvs[Vol]
// 	if !ok {
// 		err := fmt.Errorf("no such partition group: %s", Vol)
// 		return "", err
// 	}

// 	dir := lv.MntPoint
// 	encPath := "/" + LocGid + "/L_"
// 	var cid string

// 	fmt.Println("getnoncodedchunk")

// 	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if err == nil {
// 			if cid != "" {
// 				return nil
// 			}
// 		}
// 		if err != nil {
// 			err := fmt.Errorf("prevent panic by handling failure accessing a path %q: %v", dir, err)
// 			return err
// 		}
// 		ok, err := regexp.MatchString(encPath, path)
// 		if ok {
// 			cid = filepath.Base(path)
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		return "", err
// 	}

// 	return cid, nil
// }

// func (s *service) BuildObjectMap(Vol string, cid string) error {
// 	if Vol == "" || cid == "" {
// 		return fmt.Errorf("Invalid arguments :%s, %s", Vol, cid)
// 	}

// 	lv, ok := s.lvs[Vol]
// 	if !ok {
// 		return fmt.Errorf("no such partition group: %s", Vol)
// 	}

// 	lv.Lock.Chk.RLock()
// 	chk, ok := lv.ChunkMap[cid]
// 	lv.Lock.Chk.RUnlock()
// 	if !ok {
// 		return fmt.Errorf("no such chunk: %s", cid)
// 	}

// 	fChunk, err := os.OpenFile(lv.MntPoint+"/"+chk.PartID+"/"+chk.ChunkInfo.LocGid+"/"+cid, os.O_RDWR, 0775)
// 	if err != nil {
// 		return err
// 	}
// 	defer fChunk.Close()

// 	cHeader := new(repository.ChunkHeader)
// 	err = binary.Read(fChunk, binary.LittleEndian, cHeader)
// 	if err != nil {
// 		return err
// 	}

// 	//fmt.Printf("cHeader.Magic : %s, cHeader.Type : %s", cHeader.Magic, cHeader.Type)

// 	for {
// 		oHeader := new(repository.ObjHeader)
// 		err := binary.Read(fChunk, binary.LittleEndian, oHeader)
// 		if err == io.EOF {
// 			break
// 		}

// 		ObjID := strings.Trim(string(oHeader.Name[:]), "\x00")
// 		MD5 := strings.Trim(string(oHeader.MD5[:]), "\x00")

// 		//fmt.Printf("Object id : %s, MD5 : %x\n", ObjID, MD5)

// 		lv.Lock.Obj.RLock()
// 		_, ok := lv.ObjMap[ObjID]
// 		lv.Lock.Obj.RUnlock()
// 		if ok {
// 			return fmt.Errorf("object is already existed : %s", ObjID)
// 		}
// 		lv.Lock.Obj.Lock()
// 		lv.ObjMap[ObjID] = repository.ObjMap{
// 			Cid: cid,
// 			ObjInfo: repository.ObjInfo{
// 				Size: oHeader.Size,
// 				MD5:  MD5,
// 			},
// 			Offset: oHeader.Offset - s.GetObjectHeaderSize(),
// 		}

// 		lv.Lock.Obj.Unlock()

// 		fChunk.Seek(oHeader.Size, os.SEEK_CUR)
// 	}

// 	return nil
// }

// // NewClusterRepository returns a new lv store inteface in a view of cluster domain.
// func NewClusterRepository(store repository.Service) cluster.Repository {
// 	return store
// }

// // NewObjectRepository returns a new lv store inteface in a view of object domain.
// func NewObjectRepository(store repository.Service) object.Repository {
// 	return store
// }

// // NewGencodingRepository returns a new part store inteface in a view of gencoding domain.
// func NewGencodingRepository(store repository.Service) gencoding.Repository {
// 	return store
// }
