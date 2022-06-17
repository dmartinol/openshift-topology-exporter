package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ExporterConfig struct {
	Namespaces []string `yaml:",flow"`
}

func ReadConfig() *ExporterConfig {
	yfile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	exporterConfig := ExporterConfig{}
	err2 := yaml.Unmarshal(yfile, &exporterConfig)

	if err2 != nil {
		log.Fatal(err2)
	}
	return &exporterConfig
}
