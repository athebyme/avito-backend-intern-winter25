package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

// TODO : поменять зависимость auth от jwt.
type AuthConfig struct {
	Config JWTConfig `yaml:"auth"`
}

type JWTConfig struct {
	Sign string `yaml:"jwt_sign"`
}

func (c *AuthConfig) Init(filename string) (interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	config := &JWTConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
