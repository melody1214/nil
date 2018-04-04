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
	if config == nil {
		setConfig()
	}
	return config.Path(path).Data().(string)
}

func setConfig() {
	json, err := gabs.ParseJSONFile("/etc/nil/" + configFile)
	if err != nil {
		log.Panic(err)
	}

	config = json
}
