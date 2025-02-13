package config

type JWTConfig struct {
	SecretKey       string `env:"JWT_SECRET" required:"true"`
	ExpirationHours int    `env:"JWT_EXPIRATION_HOURS" default:"24"`
}
