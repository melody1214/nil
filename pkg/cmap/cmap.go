package cmap

import (
	"fmt"
	"strconv"
	"time"
)

// Version is the version of cluster map, identity of the CMap entity.
type Version int64

// Int64 returns built-in int64 type of CMapVersion.
func (v Version) Int64() int64 {
	return int64(v)
}

// Time is the time string of cluster map.
type Time string

// String returns built-in string type of CMapTime.
func (t Time) String() string {
	return string(t)
}

// Now returns the current time of CMapTime.
func Now() Time {
	return Time(time.Now().UTC().String())
}

// ID is the id of the member.
type ID int64

// String returns a string type of the ID.
func (i ID) String() string {
	return strconv.FormatInt(i.Int64(), 10)
}

// Int64 returns a int64 type of the ID.
func (i ID) Int64() int64 {
	return int64(i)
}

// CMap is the cluster map entity.
type CMap struct {
	Version Version         `xml:"version"`
	Time    Time            `xml:"time"`
	Nodes   []Node          `xml:"node"`
	Vols    []Volume        `xml:"volume"`
	EncGrps []EncodingGroup `xml:"encgrp"`
}

// HumanReadable returns a human readable map of the cluster.
func (m *CMap) HumanReadable() string {
	ids2str := func(ids []ID) string {
		r := "{"
		for i, id := range ids {
			r += id.String()
			if i < len(ids)-1 {
				r += ", "
			}
		}
		r += "}"
		return r
	}

	// Write cmap header.
	out := "+-------------------------+-----------------------------------------------------+\n"
	header := fmt.Sprintf(
		"| Cluster map version %3d | created at %-40s |\n",
		m.Version.Int64(),
		m.Time,
	)
	out += header
	out += "+-------------------------+-----------------------------------------------------+\n"

	// Make human readable sentences for each nodes.
	out += "\n"
	out += "+-------------------+\n"
	out += "| Nodes information |\n"
	out += "+------+------+-----+----------------+---------+-------------------+------------------------------+\n"
	out += "| ID   | Type | Address              | Status  | UUID              | Volumes                      |\n"
	out += "+------+------+----------------------+---------+-------------------+------------------------------+\n"
	for _, n := range m.Nodes {
		row := fmt.Sprintf(
			"| %-4s | %-4s | %-20s | %-7s | %-17s | %-28s |\n",
			n.ID.String(),
			n.Type.String(),
			n.Addr,
			n.Stat.String(),
			n.Name,
			ids2str(n.Vols),
		)
		out += row
	}
	out += "+------+------+----------------------+---------+-------------------+------------------------------+\n"

	// Make human readable information for each volumes.
	out += "\n"
	out += "+---------------------+\n"
	out += "| Volumes information |\n"
	out += "+------+---------+----+----+-----------+------+---------------------------------------------------+\n"
	out += "| ID   | Size    | Speed   | Status    | Node | EncGrps                                           |\n"
	out += "+------+---------+---------+-----------+------+---------------------------------------------------+\n"
	for _, v := range m.Vols {
		row := fmt.Sprintf(
			"| %-4s | %-7d | %-7s | %-9s | %-4s | %-49s |\n",
			v.ID.String(),
			v.Size,
			v.Speed.String(),
			v.Stat.String(),
			v.Node.String(),
			ids2str(v.EncGrps),
		)
		out += row
	}
	out += "+------+---------+---------+-----------+------+---------------------------------------------------+\n"

	// Make human readable information for each encoding groups.
	out += "\n"
	out += "+-----------------------------+\n"
	out += "| Encoding groups information |\n"
	out += "+------+---------+---------+--+------+---------+--------------------------------------+-----------+\n"
	out += "| ID   | Size    | Free    | Used    | Status  | Volumes                              | Unencoded |\n"
	out += "+------+---------+---------+---------+---------+--------------------------------------+-----------+\n"
	for _, eg := range m.EncGrps {
		row := fmt.Sprintf(
			"| %-4s | %-7d | %-7d | %-7d | %-7s | %-36s | %9d |\n",
			eg.ID.String(),
			eg.Size,
			eg.Free,
			eg.Used,
			eg.Stat.String(),
			ids2str(eg.Vols),
			eg.Uenc,
		)
		out += row
	}
	out += "+------+---------+---------+---------+---------+--------------------------------------+-----------+\n"

	return out
}

// Save stores the cluster map to the local file system.
func (m *CMap) Save() error {
	// 1. Get store file path.
	path := filePath(m.Version.Int64())

	// 2. Create empty file with the version.
	if err := createFile(path); err != nil {
		return err
	}

	// 3. Encode map data into the created file.
	if err := encode(*m, path); err != nil {
		removeFile(path)
		return err
	}

	return nil
}
