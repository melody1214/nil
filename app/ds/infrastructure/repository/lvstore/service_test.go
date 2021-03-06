package lvstore

// func TestServiceAPIs(t *testing.T) {
// 	dir := "testServiceAPIs"
// 	os.Mkdir(dir, 0775)
// 	defer os.RemoveAll(dir)

// 	os.Mkdir(dir+"/lv1", 0775)
// 	lv1 := &lv{
// 		Vol: &repository.Vol{
// 			Name:      "lv1",
// 			MntPoint:  dir + "/lv1",
// 			Size:      1024,
// 			Speed:     repository.High,
// 			ChunkSize: 10000000,
// 			ChunkMap:  make(map[string]repository.ChunkMap),
// 			ObjMap:    make(map[string]repository.ObjMap),
// 		},
// 	}

// 	os.Mkdir(dir+"/lv2", 0775)
// 	lv2 := &lv{
// 		Vol: &repository.Vol{
// 			Name:      "lv2",
// 			MntPoint:  dir + "/lv2",
// 			Size:      1024,
// 			Speed:     repository.High,
// 			ChunkSize: 10000000,
// 			ChunkMap:  make(map[string]repository.ChunkMap),
// 			ObjMap:    make(map[string]repository.ObjMap),
// 		},
// 	}

// 	s := newService(dir)
// 	s.lvs[lv1.Name] = lv1
// 	s.lvs[lv2.Name] = lv2

// 	go s.Run()
// 	runtime.Gosched()

// 	testCases := []struct {
// 		op                 repository.Operation
// 		lv, lgid, oid, cid string
// 		typ                string
// 		size               int64
// 		Md5                string
// 		content            string
// 		result             error
// 	}{
// 		{repository.Read, "lv1", "lg1", "banana", "chunk1", "Parity", int64(len("banana is good\n")), "", "banana is good\n",
// 			fmt.Errorf("no such object: banana"),
// 		},
// 		{repository.Write, "lv1", "lg1", "banana", "chunk1", "Parity", int64(len("banana is good\n")), "", "banana is good\n", nil},
// 		{repository.Read, "lv1", "lg1", "banana", "chunk1", "Data", int64(len("banana is good\n")), "", "banana is good\n", nil},
// 		{repository.Delete, "lv2", "lg2", "apple", "chunk1", "Data", int64(len("apple is sweet\n")), "", "apple is sweet\n",
// 			fmt.Errorf("no such object: apple"),
// 		},
// 		{repository.Read, "lv2", "lg2", "apple", "chunk2", "Data", int64(len("apple is sweet\n")), "", "apple is sweet\n",
// 			fmt.Errorf("no such object: apple"),
// 		},
// 		{repository.Write, "lv2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
// 		{repository.Read, "lv2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
// 		{repository.Delete, "lv2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
// 		{repository.Write, "lv1", "lg1", "orange", "chunk1", "Parity", int64(len("orange is good\n")), "", "orange is good\n", nil},
// 		{repository.Write, "lv1", "lg1", "strawberry", "chunk3", "Parity", int64(len("orange is good\n")), "", "orange is good\n", nil},
// 		{repository.Write, "lv1", "lg1", "pineapple", "chunk1", "Parity", int64(len("pineapple is good\n")), "", "pineapple is good\n", nil},
// 		{repository.Write, "lv1", "lg1", "watermelon", "chunk1", "Parity", int64(len("watermelon is good\n")), "", "watermelon is good\n", nil},
// 		{repository.DeleteReal, "lv1", "lg1", "orange", "", "Parity", int64(len("orange is good\n")), "", "orange is good\n",
// 			fmt.Errorf("can remove only a last object of a chunk")},
// 		{repository.DeleteReal, "lv1", "lg1", "pineapple", "", "Parity", int64(len("pineapple is good\n")), "", "pineapple is good\n",
// 			fmt.Errorf("can remove only a last object of a chunk")},
// 		{repository.DeleteReal, "lv1", "lg1", "watermelon", "", "Parity", int64(len("watermelon is good\n")), "", "watermelon is good\n", nil},
// 		{repository.DeleteReal, "lv2", "lg2", "", "chunk2", "Parity", int64(len("pineapple is good\n")), "", "pineapple is good\n", nil},
// 	}

// 	h := md5.New()
// 	for _, c := range testCases {
// 		var b bytes.Buffer
// 		writer := bufio.NewWriter(&b)

// 		io.WriteString(h, c.content)
// 		md5 := string(h.Sum(nil))

// 		r := &repository.Request{
// 			Op:     c.op,
// 			Vol:    c.lv,
// 			LocGid: c.lgid,
// 			Oid:    c.oid,
// 			Cid:    c.cid,
// 			Osize:  c.size,
// 			Md5:    md5,
// 			In:     strings.NewReader(c.content),
// 			Out:    writer,
// 		}

// 		s.Push(r)

// 		err := r.Wait()
// 		if err != nil && c.result != nil {
// 			if err.Error() != c.result.Error() {
// 				t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.lv, c.oid, c.result, err)
// 			}
// 			continue
// 		} else if err != c.result {
// 			t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.lv, c.oid, c.result, err)
// 			continue
// 		}

// 		if c.op == repository.Read && b.String() != c.content {
// 			t.Errorf("%v %s/%s: expected data %v, got %v", c.op, c.lv, c.oid, c.content, b.String())
// 			continue
// 		}
// 	}

// 	s.RenameChunk("chunk3", "g_chunk3", "lv1", "lg1")
// }
