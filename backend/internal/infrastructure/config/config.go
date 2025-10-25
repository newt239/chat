package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Wasabi   WasabiConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  int // minutes
	RefreshTokenTTL int // days
}

type WasabiConfig struct {
	Endpoint        string
	Region          string
	BucketName      string
	AccessKeyID     string
	SecretAccessKey string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	// Load .env file if exists (optional in production)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/chat?sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTokenTTL:  getEnvInt("JWT_ACCESS_TOKEN_TTL", 15),
			RefreshTokenTTL: getEnvInt("JWT_REFRESH_TOKEN_TTL", 30),
		},
		Wasabi: WasabiConfig{
			Endpoint:        getEnv("WASABI_ENDPOINT", "https://s3.wasabisys.com"),
			Region:          getEnv("WASABI_REGION", "us-east-1"),
			BucketName:      getEnv("WASABI_BUCKET", "chat-attachments"),
			AccessKeyID:     getEnv("WASABI_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("WASABI_SECRET_ACCESS_KEY", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")},
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "change-me-in-production" && c.Server.Env == "production" {
		return fmt.Errorf("JWT_SECRET must be set in production")
	}
	if c.Wasabi.AccessKeyID == "" || c.Wasabi.SecretAccessKey == "" {
		return fmt.Errorf("Wasabi credentials must be set")
	}
	return nil
}
