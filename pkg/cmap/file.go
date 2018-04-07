package cmap

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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

func removeFile(path string) error {
	return os.Remove(path)
}

func getLatestMapFile() (string, error) {
	files, err := listMapFiles()
	if err != nil {
		return "", err
	}

	latest := int64(0)
	name := ""
	for _, f := range files {
		var ver int64
		if _, err := fmt.Sscanf(f, fileName+"_%d"+fileExtension, &ver); err != nil {
			return "", err
		}

		if ver >= latest {
			name = f
		}
	}

	return baseDir + "/" + name, nil
}

func listMapFiles() ([]string, error) {
	maps := make([]string, 0)

	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return maps, fmt.Errorf("no map files")
	}

	for _, f := range files {
		if strings.Contains(f.Name(), fileName) {
			maps = append(maps, f.Name())
		}
	}

	return maps, nil
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

func store(m *CMap) error {
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
