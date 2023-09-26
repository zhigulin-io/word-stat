package server

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func LoadConfig(filename string) Config {
	storageConfigData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("cannot open server config file")
	}

	storageConfig := Config{}
	err = yaml.Unmarshal(storageConfigData, &storageConfig)
	if err != nil {
		log.Fatal("cannot convert yaml to config:", err)
	}

	return storageConfig
}
