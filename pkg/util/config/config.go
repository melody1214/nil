package config

import (
	"log"

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
	json, err := gabs.ParseJSONFile(configFile)
	if err != nil {
		log.Panic(err)
	}

	config = json
}
