package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ExporterConfig struct {
	Namespaces     []string `yaml:",flow"`
	FormatterClass string
	LogLevel       string
	LogFile        string
	KNative        bool
}

func ReadConfig() *ExporterConfig {
	yfile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	exporterConfig := ExporterConfig{}
	err = yaml.Unmarshal(yfile, &exporterConfig)

	if err != nil {
		log.Fatal(err)
	}
	return &exporterConfig
}
