package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Postgres PostgresConfig `yaml:"postgres"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type JWTConfig struct {
	SecretKey     string        `yaml:"secret_key"`
	TokenLifetime time.Duration `yaml:"token_lifetime"`
}

func (pc *PostgresConfig) GetConnectionString() string {
	sslMode := pc.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s pool_max_conns=50",
		pc.User, pc.Password, pc.Host, pc.Port, pc.DBName, sslMode)
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	cfg := &Config{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}
	if cfg.JWT.TokenLifetime <= 0 {
		return fmt.Errorf("JWT token lifetime must be positive")
	}
	if cfg.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}
	return nil
}
