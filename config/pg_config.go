package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type PostgresConfig struct {
	Config DatabaseConfiguration `yaml:"postgres"`
}

func (pc *PostgresConfig) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pc.Config.User, pc.Config.Password, pc.Config.Host, pc.Config.Port, pc.Config.DBName)
}

func (pc *PostgresConfig) Init(filename string) (interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	config := &PostgresConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
