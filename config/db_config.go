package config

type DatabaseConfig interface {
	GetConnectionString() string
}

type DatabaseConfiguration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
}
