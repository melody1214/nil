package partstore

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chanyoung/nil/app/ds/application/cluster"
	"github.com/chanyoung/nil/app/ds/application/gencoding"
	"github.com/chanyoung/nil/app/ds/application/object"
	"github.com/chanyoung/nil/app/ds/domain/model/device"
	"github.com/chanyoung/nil/app/ds/domain/model/volume"
	"github.com/chanyoung/nil/app/ds/infrastructure/repository"
)

// service is the backend store service.
type service struct {
	pgs          map[string]*pg
	devs         map[string]*dev
	basePath     string
	requestQueue queue
	pushCh       chan interface{}
	devLock      sync.RWMutex
}

// NewService returns a new backend store service.
func NewService(basePath string) repository.Service {
	return &service{
		basePath: basePath,
		pgs:      map[string]*pg{},
		devs:     map[string]*dev{},
		pushCh:   make(chan interface{}, 1),
	}
}

// newService returns a new backend store service.
// This is only for unit test. Do not use in real service.
func newService(basePath string) *service {
	return &service{
		basePath: basePath,
		pgs:      map[string]*pg{},
		devs:     map[string]*dev{},
		pushCh:   make(chan interface{}, 1),
	}
}

// Run starts to serve backend store service.
func (s *service) Run() {
	checkTicker := time.NewTicker(100 * time.Millisecond)
	spinUpCheckTicker := time.NewTicker(10 * time.Second)
	spinDownTicker := time.NewTicker(15 * time.Second)

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
		case <-spinUpCheckTicker.C:
			// fmt.Printf("spinUpCheckTicker, uid : %d\n", os.Getuid())
			if os.Getuid() != 0 {
				if c, err := s.CheckSpinUp(); err == nil {
					fmt.Printf("Migrate Data..\n")
					go s.MigrateData(c)
				}
			} else {
				if c, err := s.checkSpinUp(); err == nil {
					fmt.Printf("Migrate data......\n")
					go s.MigrateData(c)
				}
			}
		case <-spinDownTicker.C:
			for p, dev := range s.devs {
				go s.SpinDown(p, dev)
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

	// Get a path of block dev files using mount point of cold partitions, and assign that to dev.
	for i := 1; i <= int(v.SubPartGroup.Cold.NumOfPart); i++ {
		partID := "cold_part" + strconv.Itoa(i)
		partPath := v.MntPoint + "/" + partID

		_, err := os.Stat(partPath)
		if os.IsNotExist(err) {
			os.MkdirAll(partPath, 0775)
		}

		s.devLock.RLock()
		_, ok := s.devs[partID]
		s.devLock.RUnlock()
		if ok {
			break
		}

		// For test
		re := regexp.MustCompile(`/dev/loop[0-9]+`)

		//re := regexp.MustCompile("/dev/sd.?")

		devPath, err := exec.Command("di", "-n", "-f", "S", partPath).CombinedOutput()
		if err != nil {
			return err
		}

		t := time.Now()
		tn := t.Nanosecond()

		s.devLock.Lock()
		s.devs[partID] = &dev{
			Name:      re.FindString(string(devPath)),
			Timestamp: uint(tn),
		}
		s.devLock.Unlock()
	}

	// Get a path of block dev files using mount point of hot partitions, and assign that to dev.
	for i := 1; i <= int(v.SubPartGroup.Hot.NumOfPart); i++ {
		partID := "hot_part" + strconv.Itoa(i)
		partPath := v.MntPoint + "/" + partID

		_, err := os.Stat(partPath)
		if os.IsNotExist(err) {
			os.MkdirAll(partPath, 0775)
		}

		s.devLock.RLock()
		_, ok := s.devs[partID]
		s.devLock.RUnlock()
		if ok {
			break
		}

		// For test
		re := regexp.MustCompile(`/dev/loop[0-9]+`)

		//re := regexp.MustCompile("/dev/sd.?")

		devPath, err := exec.Command("di", "-n", "-f", "S", partPath).CombinedOutput()
		if err != nil {
			return err
		}

		t := time.Now()
		tn := t.Nanosecond()

		s.devLock.Lock()
		s.devs[partID] = &dev{
			Name:      re.FindString(string(devPath)),
			State:     "Active",
			Timestamp: uint(tn),
		}
		s.devLock.Unlock()
	}

	// Check the user previllige and initially spin-down cold storages.
	s.devLock.Lock()
	if os.Getuid() != 0 {
		for partID, di := range s.devs {
			if strings.HasPrefix(partID, "cold_part") {
				di.State = "Standby"
			}
		}
	} else {
		for partID, di := range s.devs {
			if !strings.HasPrefix(partID, "cold_part") {
				continue
			}

			if strings.HasPrefix(di.Name, "/dev/sd") {
				_, err := exec.Command("hdparm", "-qy", di.Name).CombinedOutput()
				if err != nil {
					s.devLock.Unlock()
					return err
				}
			}
			di.State = "Standby"
		}
	}
	s.devLock.Unlock()

	// Set volume has running state.
	v.Status = repository.Active

	return nil
}

func getRWbytes(path string) (int, error) {
	f, err := os.Open("/proc/diskstats")
	if err != nil {
		return -1, err
	}

	defer f.Close()

	var diskstat []string
	var reads uint64
	var writes uint64

	r := bufio.NewReader(f)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		diskstat = append(diskstat, strings.Trim(line, "\n"))
	}

	for _, line := range diskstat {
		fields := strings.Fields(line)

		if len(fields) < 14 {
			// malformed line in /proc/diskstats
			continue
		}
		name := fields[2]

		if !strings.HasSuffix(path, name) {
			continue
		}

		reads, err = strconv.ParseUint((fields[3]), 10, 64)
		if err != nil {
			return -1, err
		}

		writes, err = strconv.ParseUint((fields[7]), 10, 64)
		if err != nil {
			return -1, err
		}
		break
	}

	// TODO: Need to change unit in bytes
	totalIO := reads + writes
	return int(totalIO), nil
}

func (s *service) SpinDown(p string, d *dev) {
	var oldTotalIO uint
	var newTotalIO uint
	var diffTotalIO uint
	if !strings.HasPrefix(p, "cold_part") {
		return
	}

	if d.State != "Active" {
		return
	}

	if os.Getuid() != 0 {
		s.devLock.RLock()
		oldTotalIO = d.TotalIO
		s.devLock.RUnlock()
	} else {
		oldTotalIO, err := getRWbytes(d.Name)
		if err != nil {
			return
		}
		d.TotalIO = uint(oldTotalIO)
	}

	timer := time.NewTimer(10 * time.Second)
	<-timer.C
	fmt.Println("Timer expired")

	if os.Getuid() != 0 {
		s.devLock.RLock()
		newTotalIO = d.TotalIO
		s.devLock.RUnlock()
	} else {
		newTotalIO, err := getRWbytes(d.Name)
		if err != nil {
			return
		}
		d.TotalIO = uint(newTotalIO)
	}

	diffTotalIO = newTotalIO - oldTotalIO
	// TODO: change threshold for diffTotalIO
	if diffTotalIO > 0 {
		return
	}

	t := time.Now()
	tn := t.Nanosecond()

	tDiff := uint(tn) - d.Timestamp
	if os.Getuid() != 0 {
		s.devLock.Lock()
		d.State = "Standby"
		d.Timestamp = uint(tn)
		d.ActiveTime = d.ActiveTime + tDiff
		s.devLock.Unlock()
		return
	}

	if strings.HasPrefix(d.Name, "/dev/sd") {
		_, err := exec.Command("hdparm", "-qy", d.Name).CombinedOutput()
		if err != nil {
			return
		}
	}

	s.devLock.Lock()
	d.State = "Standby"
	d.Timestamp = uint(tn)
	d.ActiveTime = d.ActiveTime + tDiff
	s.devLock.Unlock()

	return
}

// For test
func (s *service) CheckSpinUp() (string, error) {
	oldDev := dev{}
	var spinUpPart = ""

	s.devLock.RLock()
	for part, dev := range s.devs {
		fmt.Printf("%s\n", part)
		if dev.State != "Active" {
			continue
		}
		if !strings.HasPrefix(part, "cold_part") {
			continue
		}
		if !strings.HasPrefix(spinUpPart, "cold_part") {
			oldDev = *dev
			spinUpPart = part
			continue
		}
		if oldDev.Free < dev.Free {
			oldDev = *dev
			spinUpPart = part
			continue
		}
	}
	s.devLock.RUnlock()

	if !strings.HasPrefix(spinUpPart, "cold_part") {
		err := fmt.Errorf("there are no spin-up partitions")
		return "", err
	}
	return spinUpPart, nil
}

// For real experimental environment
func (s *service) checkSpinUp() (string, error) {
	oldDev := dev{}
	var spinUpPart = ""

	s.devLock.Lock()
	for part, dev := range s.devs {
		cmd := exec.Command("smartctl", "-n", "standby", dev.Name)

		err := cmd.Start()
		if err != nil {
			err = fmt.Errorf("error occurs during run of smartctl")
			s.devLock.Unlock()
			return "", err
		}

		err = cmd.Wait()
		if err != nil {
			continue
		}

		if !strings.HasPrefix(oldDev.Name, "/dev/sd") {
			oldDev = *dev
			spinUpPart = part
			continue
		}

		if oldDev.Free < dev.Free {
			oldDev = *dev
			spinUpPart = part
			continue
		}
	}
	s.devLock.Unlock()

	if !strings.HasPrefix(spinUpPart, "/dev/sd") {
		err := fmt.Errorf("there are no spin-up disks")
		return "", err
	}
	return spinUpPart, nil
}

// MigrateData moves data from service to archive archive.
func (s *service) MigrateData(DestPartID string) error {
	for _, pg := range s.pgs {
		pg.Lock.Chk.Lock()
		defer pg.Lock.Chk.Unlock()

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
				//pg.SubPartGroup.Cold.DiskSched = pg.SubPartGroup.Cold.DiskSched%pg.SubPartGroup.Cold.NumOfPart + 1
				//DiskSched := pg.SubPartGroup.Cold.DiskSched
				//DestPartID := "cold_part" + strconv.Itoa(int(DiskSched))

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
				n, err := io.Copy(fChunkDest, fChunkSrc)

				s.devLock.RLock()
				dstDev, ok := s.devs[DestPartID]
				s.devLock.RUnlock()
				if !ok {
					err = fmt.Errorf("no such dev named as such partition")
					return err
				}

				s.devLock.RLock()
				srcDev, ok := s.devs[PartID]
				s.devLock.RUnlock()
				if !ok {
					err = fmt.Errorf("no such dev named as such partition")
					return err
				}

				s.devLock.Lock()
				srcDev.Used = srcDev.Used - uint(n)
				srcDev.Free = srcDev.Free + uint(n)
				srcDev.Size = srcDev.Used + srcDev.Free
				srcDev.TotalIO = srcDev.TotalIO + uint(n)
				s.devLock.Unlock()

				s.devLock.Lock()
				dstDev.Used = dstDev.Used + uint(n)
				dstDev.Free = dstDev.Free - uint(n)
				dstDev.Size = dstDev.Used + dstDev.Free
				dstDev.TotalIO = dstDev.TotalIO + uint(n)
				s.devLock.Unlock()

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

				err = fChunkDest.Close()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *service) ChunkExist(pgID, chkID string) bool {
	pg, ok := s.pgs[pgID]
	if ok == false {
		return false
	}

	pg.Lock.Chk.RLock()
	_, ok = pg.ChunkMap[chkID]
	pg.Lock.Chk.RUnlock()
	return ok
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

	objSize := obj.ObjInfo.Size

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
	return 8
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

	// Open a chunk requested by a client.
	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid
	fChunk, err := os.Open(lgDir + "/" + obj.Cid)
	if err != nil {
		r.Err = err
		return
	}
	defer fChunk.Close()

	// If the partition is one of the spin-down disks, spin-up the disk immediately.
	s.devLock.RLock()
	dev, ok := s.devs[chk.PartID]
	s.devLock.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such partition in partgroup : %s", chk.PartID)
		return
	}

	t := time.Now()
	tn := t.Nanosecond()
	tDiff := uint(tn) - dev.Timestamp
	dev.Timestamp = uint(tn)

	s.devLock.Lock()
	if dev.State == "Standby" {
		dev.State = "Active"
		dev.StandbyTime = dev.StandbyTime + tDiff
	} else {
		dev.ActiveTime = dev.ActiveTime + tDiff
	}
	s.devLock.Unlock()

	// Seek offset beginning of the requested object in the chunk.
	_, err = fChunk.Seek(obj.Offset, os.SEEK_SET)
	if err != nil {
		r.Err = err
		return
	}

	oHeader := new(repository.ObjHeader)
	err = binary.Read(fChunk, binary.LittleEndian, oHeader)
	//fmt.Printf("%s, %d, %d\n", oHeader.Name, oHeader.Size, oHeader.Offset)
	if err != nil {
		r.Err = err
		return
	}

	_, err = fChunk.Seek(oHeader.Offset, os.SEEK_SET)
	if err != nil {
		r.Err = err
		return
	}

	// Read contents of the requested object from the chunk.
	n, err := io.CopyN(r.Out, fChunk, r.Osize)
	if err != nil {
		r.Err = err
		return
	}

	if dev.Name != "" {
		dev.TotalIO = dev.TotalIO + uint(n)
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
	pg.Lock.Chk.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no chunk of such object: %s", r.Oid)
		return
	}

	// Open a chunk requested by a client.
	lgDir := pg.MntPoint + "/" + chk.PartID + "/" + r.LocGid
	fChunk, err := os.Open(lgDir + "/" + r.Cid)
	if err != nil {
		r.Err = err
		return
	}
	defer fChunk.Close()

	// If the partition is one of the spin-down disks, spin-up the disk immediately.
	s.devLock.RLock()
	dev, ok := s.devs[chk.PartID]
	s.devLock.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such partition in partgroup : %s", chk.PartID)
		return
	}

	t := time.Now()
	tn := t.Nanosecond()
	tDiff := uint(tn) - dev.Timestamp
	dev.Timestamp = uint(tn)

	s.devLock.Lock()
	if dev.State == "Standby" {
		dev.State = "Active"
		dev.StandbyTime = dev.StandbyTime + tDiff
	} else {
		dev.ActiveTime = dev.ActiveTime + tDiff
	}
	s.devLock.Unlock()

	// Read all contents from the chunk to a writer stream.
	n, err := io.Copy(r.Out, fChunk)
	if err != nil {
		r.Err = err
		return
	}

	if dev.Name != "" {
		dev.TotalIO = dev.TotalIO + uint(n)
	}
	// Complete to read the all contents from the chunk.
	r.Err = nil
	return
}

func (s *service) write(r *repository.Request) {
	var totalWrite uint

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
			Type:    [1]byte{},
			State:   [1]byte{'P'},
			Encoded: false,
		}

		cHeader.Type[0] = 'D'

		b := new(bytes.Buffer)
		bufio.NewWriter(b)
		err := binary.Write(b, binary.LittleEndian, cHeader)
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

		totalWrite = totalWrite + uint(n)
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
		Name:   [48]byte{},
		MD5:    [32]byte{},
		Size:   r.Osize,
		Offset: fChunkLen + s.GetObjectHeaderSize(),
	}

	for i := 0; i < len(r.Oid); i++ {
		oHeader.Name[i] = r.Oid[i]
	}
	//fmt.Println("len(r.Oid) : ", len(r.Oid))

	for i := len(r.Oid); i < len(oHeader.Name); i++ {
		oHeader.Name[i] = '\x00'
	}

	//fmt.Printf("len(r.Md5) : %d, len(oHeader.MD5) : %d\n", len(r.Md5), len(oHeader.MD5))
	for i := 0; i < len(r.Md5); i++ {
		oHeader.MD5[i] = r.Md5[i]
	}

	b := new(bytes.Buffer)
	bufio.NewWriter(b)
	err = binary.Write(b, binary.LittleEndian, oHeader)
	if err != nil {
		r.Err = err
		return
	}

	//fmt.Println(b.Len())

	// If the partition is one of the spin-down disks, spin-up the disk immediately.
	s.devLock.RLock()
	dev, ok := s.devs[chk.PartID]
	s.devLock.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such partition in partgroup : %s", chk.PartID)
		return
	}

	t := time.Now()
	tn := t.Nanosecond()
	tDiff := uint(tn) - dev.Timestamp
	dev.Timestamp = uint(tn)

	s.devLock.Lock()
	if dev.State == "Standby" {
		dev.State = "Active"
		dev.StandbyTime = dev.StandbyTime + tDiff
	} else {
		dev.ActiveTime = dev.ActiveTime + tDiff
	}
	s.devLock.Unlock()

	// Check whether the chunk is full or has not enough space to write the object.
	if fChunkLen >= pg.ChunkSize {
		r.Err = fmt.Errorf("chunk full")
		return
	} else if fChunkLen+int64(b.Len())+r.Osize > pg.ChunkSize {
		err = fChunk.Truncate(pg.ChunkSize)
		totalWrite = totalWrite + uint(pg.ChunkSize-fChunkLen)
		if dev.Name != "" {
			dev.TotalIO = dev.TotalIO + totalWrite
			dev.Free = dev.Free - totalWrite
			dev.Used = dev.Used + totalWrite
		}
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
	totalWrite = totalWrite + uint(n)
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
		Cid: r.Cid,
		ObjInfo: repository.ObjInfo{
			Size: r.Osize,
			MD5:  r.Md5,
		},
		Offset: fChunkLen,
	}

	pg.Lock.Obj.Unlock()

	if dev.Name != "" {
		dev.TotalIO = dev.TotalIO + totalWrite + uint(r.Osize)
		dev.Free = dev.Free - (totalWrite + uint(r.Osize))
		dev.Used = dev.Used + (totalWrite + uint(r.Osize))
	}

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

	// Create a directory for a local group if not exist.
	lgDir := pg.MntPoint + "/" + PartID + "/" + r.LocGid
	_, err := os.Stat(lgDir)
	if os.IsNotExist(err) {
		os.MkdirAll(lgDir, 0775)
	}

	// Open a chunk that objects will be written to.
	fChunk, err := os.OpenFile(lgDir+"/"+r.Cid, os.O_CREATE|os.O_WRONLY, 0775)
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

	pg.Lock.Chk.Lock()
	pg.ChunkMap[r.Cid] = repository.ChunkMap{
		PartID: PartID,
		ChunkInfo: repository.ChunkInfo{
			Type:   r.Type,
			LocGid: r.LocGid,
		},
	}
	pg.Lock.Chk.Unlock()

	// If the partition is one of the spin-down disks, spin-up the disk immediately.
	s.devLock.RLock()
	dev, ok := s.devs[chk.PartID]
	s.devLock.RUnlock()
	if !ok {
		r.Err = fmt.Errorf("no such partition in partgroup : %s", chk.PartID)
		return
	}

	t := time.Now()
	tn := t.Nanosecond()
	tDiff := uint(tn) - dev.Timestamp
	dev.Timestamp = uint(tn)

	s.devLock.Lock()
	if dev.State == "Standby" {
		dev.State = "Active"
		dev.StandbyTime = dev.StandbyTime + tDiff
	} else {
		dev.ActiveTime = dev.ActiveTime + tDiff
	}
	s.devLock.Unlock()

	if dev.Name != "" {
		dev.TotalIO = dev.TotalIO + uint(n)
		dev.Free = dev.Free - uint(n)
		dev.Used = dev.Used + uint(n)
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

		fChunkInfo, err := os.Lstat(lgDir + "/" + r.Cid)
		if err != nil {
			r.Err = err
			return
		}

		fChunkLen := fChunkInfo.Size()

		err = os.Remove(lgDir + "/" + r.Cid)
		if err != nil {
			r.Err = fmt.Errorf("no such chunk: %s", r.Cid)
			return
		}

		pg.Lock.Chk.Lock()
		delete(pg.ChunkMap, r.Cid)
		pg.Lock.Chk.Unlock()

		s.devLock.RLock()
		dev, ok := s.devs[chk.PartID]
		if ok {
			dev.TotalIO = dev.TotalIO + uint(fChunkLen)
			dev.Used = dev.Used - uint(fChunkLen)
			dev.Free = dev.Free + uint(fChunkLen)
			dev.Size = dev.Used + dev.Free
		}
		s.devLock.RUnlock()

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

	if obj.Offset+obj.ObjInfo.Size+s.GetObjectHeaderSize() != fChunkLen {
		r.Err = fmt.Errorf("can remove only a last object of a chunk")
		return
	}
	fChunk.Seek(obj.Offset, io.SeekStart)
	fChunk.Truncate(fChunkLen - (fChunkLen - obj.Offset))

	pg.Lock.Obj.Lock()
	delete(pg.ObjMap, r.Oid)
	pg.Lock.Obj.Unlock()

	s.devLock.RLock()
	dev, ok := s.devs[chk.PartID]
	s.devLock.RUnlock()

	if ok {
		dev.TotalIO = dev.TotalIO + uint(fChunkLen-obj.Offset)
		dev.Free = dev.Free + uint(fChunkLen-obj.Offset)
		dev.Used = dev.Free - uint(fChunkLen-obj.Offset)
		dev.Size = dev.Free + dev.Used
	}

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

func (s *service) BuildObjectMap(Vol string, cid string) error {
	if Vol == "" || cid == "" {
		return fmt.Errorf("Invalid arguments :%s, %s", Vol, cid)
	}

	pg, ok := s.pgs[Vol]
	if !ok {
		return fmt.Errorf("no such partition group: %s", Vol)
	}

	pg.Lock.Chk.RLock()
	chk, ok := pg.ChunkMap[cid]
	pg.Lock.Chk.RUnlock()
	if !ok {
		return fmt.Errorf("no such chunk: %s", cid)
	}

	fChunk, err := os.OpenFile(pg.MntPoint+"/"+chk.PartID+"/"+chk.ChunkInfo.LocGid+"/"+cid, os.O_RDWR, 0775)
	if err != nil {
		return err
	}
	defer fChunk.Close()

	cHeader := new(repository.ChunkHeader)
	err = binary.Read(fChunk, binary.LittleEndian, cHeader)
	if err != nil {
		return err
	}

	//fmt.Printf("cHeader.Magic : %s, cHeader.Type : %s", cHeader.Magic, cHeader.Type)

	for {
		oHeader := new(repository.ObjHeader)
		err := binary.Read(fChunk, binary.LittleEndian, oHeader)
		if err == io.EOF {
			break
		}

		ObjID := strings.Trim(string(oHeader.Name[:]), "\x00")
		MD5 := strings.Trim(string(oHeader.MD5[:]), "\x00")

		//fmt.Printf("Object id : %s, MD5 : %x\n", ObjID, MD5)

		pg.Lock.Obj.RLock()
		_, ok := pg.ObjMap[ObjID]
		pg.Lock.Obj.RUnlock()
		if ok {
			return fmt.Errorf("object is already existed : %s", ObjID)
		}
		pg.Lock.Obj.Lock()
		pg.ObjMap[ObjID] = repository.ObjMap{
			Cid: cid,
			ObjInfo: repository.ObjInfo{
				Size: oHeader.Size,
				MD5:  MD5,
			},
			Offset: oHeader.Offset - s.GetObjectHeaderSize(),
		}

		pg.Lock.Obj.Unlock()

		fChunk.Seek(oHeader.Size, os.SEEK_CUR)
	}

	return nil
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

// Refactoring //

type devRepository struct {
	*service
}

func (r *devRepository) Create(given *device.Device) error {
	return nil
}

func (s *service) NewDeviceRepository() device.Repository {
	return &devRepository{
		service: s,
	}
}

type volumeRepository struct {
	*service
}

func (r *volumeRepository) Find(name volume.Name) (volume.Volume, error) {
	return volume.Volume{}, nil
}

func (r *volumeRepository) FindAll() []volume.Volume {
	return nil
}

func (s *service) NewVolumeRepository() volume.Repository {
	return &volumeRepository{
		service: s,
	}
}
