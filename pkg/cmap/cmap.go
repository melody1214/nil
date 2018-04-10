package cmap

import (
	"fmt"
)

// Version is the version of cluster map.
type Version int64

// Int64 returns built-in int64 type of cmap.Version
func (v Version) Int64() int64 {
	return int64(v)
}

// CMap is a cluster map which includes the information about nodes.
type CMap struct {
	Version Version         `xml:"version"`
	Time    string          `xml:"time"`
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
	out := "+--------------------------------------------------+\n"
	header := fmt.Sprintf(
		"| Cluster map version %3d, created at %s |\n",
		m.Version.Int64(),
		m.Time,
	)
	out += header
	out += "+--------------------------------------------------+\n"

	// Make human readable sentences for each nodes.
	out += "\n"
	out += "+-------------------+\n"
	out += "| Nodes information |\n"
	out += "+------+------+-----+----------------+---------+-------------------+-----------------+\n"
	out += "| ID   | Type | Address              | Status  | UUID              | Volumes         |\n"
	out += "+------+------+----------------------+---------+-------------------+-----------------+\n"
	for _, n := range m.Nodes {
		row := fmt.Sprintf(
			"| %-4s | %-4s | %-20s | %-7s | %-17s | %-15s |\n",
			n.ID.String(),
			n.Type.String(),
			n.Addr,
			n.Stat.String(),
			n.Name,
			ids2str(n.Vols),
		)
		out += row
	}
	out += "+------+------+----------------------+---------+-------------------+-----------------+\n"

	// Make human readable information for each volumes.
	out += "\n"
	out += "+---------------------+\n"
	out += "| Volumes information |\n"
	out += "+------+---------+----+----+---------+---------+---------+------+-----------------+\n"
	out += "| ID   | Size    | Free    | Used    | Speed   | Status  | Node | EncGrps         |\n"
	out += "+------+---------+---------+---------+---------+---------+------+-----------------+\n"
	for _, v := range m.Vols {
		row := fmt.Sprintf(
			"| %-4s | %-7d | %-7d | %-7d | %-7s | %-7s | %-4s | %-15s |\n",
			v.ID.String(),
			v.Size,
			v.Free,
			v.Used,
			v.Speed.String(),
			v.Status.String(),
			v.Node.String(),
			ids2str(v.EncGrps),
		)
		out += row
	}
	out += "+------+---------+---------+---------+---------+---------+------+-----------------+\n"

	// Make human readable information for each encoding groups.
	out += "\n"
	out += "+-----------------------------+\n"
	out += "| Encoding groups information |\n"
	out += "+------+---------+---------+--+------+-----------------+\n"
	out += "| ID   | Size    | Free    | Used    | Volumes         |\n"
	out += "+------+---------+---------+---------+-----------------+\n"
	for _, eg := range m.EncGrps {
		row := fmt.Sprintf(
			"| %-4s | %-7d | %-7d | %-7d | %-15s |\n",
			eg.ID.String(),
			eg.Size,
			eg.Free,
			eg.Used,
			ids2str(eg.Vols),
		)
		out += row
	}
	out += "+------+---------+---------+---------+-----------------+\n"

	return out
}

// Save stores the cluster map to the local file system.
func (m *CMap) Save() error {
	return store(m)
}
