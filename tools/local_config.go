package tools

import (
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
)

type local_config struct {
	Sentry string
}

var LocalConfig local_config

func InitConfig(filePath string) {
	if !PathExists(filePath) {
		return
	}

	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Unable to read configuration %v\r\n", filePath)
		return
	}

	conf := local_config{}

	err = yaml.Unmarshal([]byte(fileContents), &conf)
	if err != nil {
		log.Printf("Unable to parse local configuration %v\r\n", filePath)
		return
	}

	LocalConfig = conf
}
