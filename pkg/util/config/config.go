package config

import (
	"log"
	"os"

	"github.com/Jeffail/gabs"
)

const configFile = "config.json"

var config *gabs.Container

// Get returns config data with the given path.
// Config data is only allowed in string type.
func Get(path string) string {
	return config.Path(path).Data().(string)
}

func init() {
	// TODO: change to get config file path cleverly.
	json, err := gabs.ParseJSONFile(os.Getenv("GOPATH") + "/src/github.com/chanyoung/nil/" + configFile)
	if err != nil {
		log.Panic(err)
	}

	config = json
}
