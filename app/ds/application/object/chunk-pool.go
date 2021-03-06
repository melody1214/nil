package object

// // chunkID is the ID of chunk. It is basically generated by UUID.
// type chunkID string

// const notFound chunkID = "not found"

// // vID is the ID of volume.
// type vID string

// // egID is the ID of encoding group.
// type egID string

// type chunkStatus string

// const (
// 	// W stands for writing.
// 	W chunkStatus = "W"
// 	// L stands for locally encoded.
// 	L chunkStatus = "L"
// 	// T stands for tmp.
// 	// Encoding or decoding or moving.
// 	// T state chunk can't be the object of recovery.
// 	T chunkStatus = "T"
// 	// E stands for globally encoding.
// 	E chunkStatus = "E"
// 	// G stands for globally encoded.
// 	G chunkStatus = "G"
// 	// R stands for recovering.
// 	R chunkStatus = "R"
// 	// F stands for faulty.
// 	// GC will collects and remove from the disk.
// 	F chunkStatus = "F"
// )

// func (s chunkStatus) String() string {
// 	switch s {
// 	case W, L, T, G, R, F:
// 		return string(s)
// 	default:
// 		return "unknown"
// 	}
// }

// // chunk holds the information which are required for object writing.
// type chunk struct {
// 	id            chunkID
// 	status        chunkStatus
// 	encodingGroup egID
// 	volume        vID
// 	encoding      bool
// 	shard         int

// 	// Space information.
// 	size int64
// 	free int64
// }

// // newChunkPool returns a new chunk pool with the given configurations.
// func newChunkPool(cmapAPI cmap.SlaveAPI, shardSize, chunkSize, chunkHeaderSize, objectHeaderSize, maximumSize int64) *chunkPool {
// 	return &chunkPool{
// 		pool:             make(map[chunkID]*chunk),
// 		writing:          make(map[chunkID]*chunk),
// 		encoding:         make(map[chunkID]*chunk),
// 		shardSize:        shardSize,
// 		chunkSize:        chunkSize,
// 		chunkHeaderSize:  chunkHeaderSize,
// 		objectHeaderSize: objectHeaderSize,
// 		maximumSize:      maximumSize,
// 		cmapAPI:          cmapAPI,
// 	}
// }

// // chunkPool is the pool of available chunk in the backend store.
// type chunkPool struct {
// 	pool             map[chunkID]*chunk
// 	writing          map[chunkID]*chunk
// 	encoding         map[chunkID]*chunk
// 	shardSize        int64
// 	chunkSize        int64
// 	chunkHeaderSize  int64
// 	objectHeaderSize int64
// 	maximumSize      int64

// 	cmapAPI cmap.SlaveAPI
// 	mu      sync.Mutex
// }

// // newChunk creates a new chunk object within the given volume.
// func (p *chunkPool) newChunk(encodingGroup egID, volume vID) (chunkID, error) {
// 	mds, err := p.cmapAPI.SearchCall().Node().Type(cmap.MDS).Status(cmap.NodeAlive).Do()
// 	if err != nil {
// 		return "", err
// 	}

// 	conn, err := nilrpc.Dial(mds.Addr.String(), nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		return "", errors.Wrap(err, "failed to dial")
// 	}
// 	defer conn.Close()

// 	eg, err := strconv.ParseInt(string(encodingGroup), 10, 64)
// 	if err != nil {
// 		return "", err
// 	}

// 	req := &nilrpc.MOBGetChunkRequest{
// 		EncodingGroup: cmap.ID(eg),
// 	}
// 	res := &nilrpc.MOBGetChunkResponse{}

// 	cli := rpc.NewClient(conn)
// 	if err := cli.Call(nilrpc.MdsObjectGetChunk.String(), req, res); err != nil {
// 		return "", errors.Wrap(err, "failed to get chunk id")
// 	}

// 	cid := chunkID(res.ID)

// 	p.pool[cid] = &chunk{
// 		id:            cid,
// 		status:        W,
// 		volume:        volume,
// 		encodingGroup: encodingGroup,
// 		shard:         1,
// 		size:          p.chunkSize,
// 		free:          p.chunkSize - p.chunkHeaderSize,
// 	}

// 	return cid, nil
// }

// func (p *chunkPool) moveChunk(egID egID, vID vID, cID chunkID, shard int) error {
// 	if int64(shard) > p.shardSize {
// 		return fmt.Errorf("unmatched shard size")
// 	}

// 	c := &chunk{
// 		id:            cID,
// 		status:        "W",
// 		encodingGroup: egID,
// 		volume:        vID,
// 		shard:         shard,
// 		encoding:      false,

// 		size: p.chunkSize,
// 	}

// 	if int64(shard) == p.shardSize {
// 		c.free = 0
// 		p.encoding[cID] = c
// 		return nil
// 	}
// 	c.free = c.size - p.chunkHeaderSize
// 	c.shard++
// 	p.pool[cID] = c
// 	return nil
// }

// // FindAvailableChunk returns chunkID which is available for writing.
// // This method is never failed, because it will creates a new chunk
// // when there is no available chunk.
// func (p *chunkPool) FindAvailableChunk(encodingGroup egID, volume vID, writingSize int64) (chunkID, error) {
// 	// TODO: remove volume in parameters.
// 	// read directly from cmap by using egID.
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	// var shard string
// 	// cid := notFound
// 	var c *chunk
// 	for i := range p.pool {
// 		if p.pool[i].encodingGroup != encodingGroup {
// 			continue
// 		}

// 		if p.pool[i].free < writingSize {
// 			continue
// 		}

// 		c = p.pool[i]
// 		break
// 	}

// 	if c == nil {
// 		cid, err := p.newChunk(encodingGroup, volume)
// 		if err != nil {
// 			return "", errors.Wrap(err, "failed to get new chunk id from mds")
// 		}
// 		c = p.pool[cid]
// 	}

// 	p.writing[c.id] = p.pool[c.id]
// 	delete(p.pool, c.id)
// 	// for id, c := range p.pool {
// 	// 	if c.encodingGroup != encodingGroup {
// 	// 		continue
// 	// 	}

// 	// 	if c.free < writingSize {
// 	// 		continue
// 	// 	}

// 	// 	cid = id
// 	// 	shard = strconv.Itoa(c.shard)
// 	// 	break
// 	// }

// 	// var err error
// 	// if cid == notFound {
// 	// 	cid, err = p.newChunk(encodingGroup, volume)
// 	// 	if err != nil {
// 	// 		return "", errors.Wrap(err, "failed to get new chunk id from mds")
// 	// 	}
// 	// 	shard = strconv.Itoa(p.pool[cid].shard)
// 	// }

// 	// p.writing[cid] = p.pool[cid]
// 	// delete(p.pool, cid)

// 	// chunk ID = "cid" + "_" + "shard"
// 	// return chunkID(string(cid) + "_" + shard), nil

// 	// chunk ID = "cid" + "_" + "shard"
// 	return chunkID(c.status.String() + "_" + string(c.id) + "_" + strconv.Itoa(c.shard)), nil
// }

// // FinishWriting moves chunk with the given id into the other pools.
// // If the left free space of the chunk is less than allowed maximum
// // chunk size, then push it in the encoding pool. The endec will
// // encode it batch. If not, then push int64o the waiting pool to wait
// // another writing requests from the clients.
// func (p *chunkPool) FinishWriting(cid chunkID, writingSize int64) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	cid = chunkID(strings.Split(string(cid), "_")[1])

// 	c, ok := p.writing[cid]
// 	if ok == false {
// 		return
// 	}
// 	defer delete(p.writing, cid)

// 	c.free = c.free - writingSize
// 	if c.free > p.maximumSize {
// 		p.pool[cid] = c
// 		return
// 	}

// 	if int64(c.shard) < p.shardSize {
// 		c.shard++
// 		c.free = c.size - p.chunkHeaderSize
// 		p.pool[cid] = c
// 		return
// 	}

// 	p.encoding[cid] = c
// }

// // GetChunk returns chunk with the given chunk id.
// func (p *chunkPool) GetChunk(cid chunkID) (c chunk, ok bool) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	cid = chunkID(strings.Split(string(cid), "_")[1])

// 	for id, c := range p.pool {
// 		if id == cid {
// 			return *c, true
// 		}
// 	}

// 	for id, c := range p.writing {
// 		if id == cid {
// 			return *c, true
// 		}
// 	}

// 	for id, c := range p.encoding {
// 		if id == cid {
// 			return *c, true
// 		}
// 	}

// 	return c, false
// }

// // GetNeedEncodingChunk returns a chunk object that need to be encoded.
// func (p *chunkPool) GetNeedEncodingChunk() (c chunk, ok bool) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	for _, chunk := range p.encoding {
// 		if chunk.encoding {
// 			continue
// 		}

// 		chunk.encoding = true
// 		return *chunk, true
// 	}

// 	return c, false
// }

// // EncodingFailed set the encoding field to false.
// func (p *chunkPool) EncodingFailed(cid chunkID) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	for id, chunk := range p.encoding {
// 		if id != cid {
// 			continue
// 		}

// 		chunk.encoding = false
// 		return
// 	}
// }

// // EncodingSuccess remove chunk with the given id from encoding list.
// func (p *chunkPool) EncodingSuccess(cid chunkID) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	delete(p.encoding, cid)
// }
