package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config interface{}

type Initialize interface {
	// Init TODO : не самый лучший вариант класть filename в init бд. получается мы зависим от filename string
	// Init TODO : а если придется подключаться не из filename, например.
	Init(filename string) (Config, error)
}

type AppConfig struct {
	DatabaseConfig DatabaseConfig `yaml:"database"`
	// AuthConfig     AuthConfig     `yaml:"auth_config"`
}

func (pc *AppConfig) Init(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(pc); err != nil {
		return nil, err
	}
	return pc, nil
}
