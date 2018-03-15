package cmap

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	baseDir       = "cmap"
	fileName      = "cluster_map"
	fileExtension = ".xml"
)

func filePath(ver int64) string {
	return baseDir + "/" + fileName + "_" + strconv.FormatInt(ver, 10) + fileExtension
}

func createFile(path string) error {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.Mkdir(baseDir, os.ModePerm)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}

func encode(m CMap, path string) error {
	// Create xml file.
	f, err := os.OpenFile(path, os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	// Encode to the file.
	xmlWriter := io.Writer(f)
	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("", "    ")
	return enc.Encode(m)
}

func decode(path string) (CMap, error) {
	// Open xml file.
	f, err := os.Open(path)
	if err != nil {
		return CMap{}, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)

	m := CMap{
		Nodes: make([]Node, 0),
	}
	if err := xml.Unmarshal(data, &m); err != nil {
		return CMap{}, err
	}

	return m, nil
}
