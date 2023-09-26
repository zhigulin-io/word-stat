package storage

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func LoadConfig(filename string) Config {
	storageConfigData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("cannot open storage config file")
	}

	storageConfig := Config{}
	err = yaml.Unmarshal(storageConfigData, &storageConfig)
	if err != nil {
		log.Fatal("cannot convert yaml to config:", err)
	}

	return storageConfig
}
