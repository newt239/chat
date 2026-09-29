package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Auth     AuthConfig
	Storage  StorageConfig
	Wasabi   WasabiConfig
	CORS     CORSConfig
	Search   SearchConfig
	Firebase FirebaseConfig
	Redis    RedisConfig
}

// RedisConfig はレプリカ間で WebSocket の配信・閲覧者一覧・レート制限を共有する Redis。未設定ならプロセス内で完結する
type RedisConfig struct {
	URL string
}

// FirebaseConfig はプッシュ通知 (FCM) の送信先プロジェクト。未設定なら通知を送らない。認証は ADC を使う
type FirebaseConfig struct {
	ProjectID string
}

// SearchConfig はメッセージの全文検索に使う Meilisearch への接続先です
type SearchConfig struct {
	MeilisearchURL    string
	MeilisearchAPIKey string
}

// StorageConfig は添付ファイルの保存先。Driver は wasabi（S3 互換）か local（開発用）
type StorageConfig struct {
	Driver string
	// local のときの保存先ディレクトリと、署名付き URL に使うバックエンドの公開 URL
	LocalDir      string
	PublicBaseURL string
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  int // minutes
	RefreshTokenTTL int // days
}

// AuthConfig は Google ログインとパスワード認証の設定。パスワード認証は production では既定で無効
type AuthConfig struct {
	GoogleOAuthClientID string
	PasswordAuthEnabled bool
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
	_ = godotenv.Load()

	env := getEnv("ENV", "development")
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  env,
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/chat?sslmode=disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTokenTTL:  getEnvInt("JWT_ACCESS_TOKEN_TTL", 15),
			RefreshTokenTTL: getEnvInt("JWT_REFRESH_TOKEN_TTL", 30),
		},
		Auth: AuthConfig{
			GoogleOAuthClientID: getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
			PasswordAuthEnabled: getEnvBool("PASSWORD_AUTH_ENABLED", env != "production"),
		},
		Storage: StorageConfig{
			Driver:        getEnv("STORAGE_DRIVER", "wasabi"),
			LocalDir:      getEnv("LOCAL_STORAGE_DIR", "tmp/storage"),
			PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:"+getEnv("PORT", "8080")),
		},
		Wasabi: WasabiConfig{
			Endpoint:        getEnv("WASABI_ENDPOINT", "https://s3.wasabisys.com"),
			Region:          getEnv("WASABI_REGION", "us-east-1"),
			BucketName:      getEnv("WASABI_BUCKET", "chat-attachments"),
			AccessKeyID:     getEnv("WASABI_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("WASABI_SECRET_ACCESS_KEY", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		},
		Search: SearchConfig{
			MeilisearchURL:    getEnv("MEILISEARCH_URL", "http://localhost:7700"),
			MeilisearchAPIKey: getEnv("MEILISEARCH_API_KEY", ""),
		},
		Firebase: FirebaseConfig{
			ProjectID: getEnv("FIREBASE_PROJECT_ID", ""),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
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

// getEnvList はカンマ区切りの環境変数をリストとして読み込みます
func getEnvList(key, defaultVal string) []string {
	values := strings.Split(getEnv(key, defaultVal), ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if d, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return d
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if b, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return b
	}
	return defaultVal
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "change-me-in-production" && c.Server.Env == "production" {
		return fmt.Errorf("JWT_SECRET must be set in production")
	}
	if c.Storage.Driver != "wasabi" && c.Storage.Driver != "local" {
		return fmt.Errorf("STORAGE_DRIVER must be wasabi or local: %q", c.Storage.Driver)
	}
	if c.Server.Env == "production" && c.Storage.Driver == "wasabi" && (c.Wasabi.AccessKeyID == "" || c.Wasabi.SecretAccessKey == "") {
		return fmt.Errorf("wasabi credentials must be set in production")
	}
	if c.Auth.GoogleOAuthClientID == "" && !c.Auth.PasswordAuthEnabled {
		return fmt.Errorf("GOOGLE_OAUTH_CLIENT_ID or PASSWORD_AUTH_ENABLED must be set to allow login")
	}
	if c.Server.Env == "production" && os.Getenv("DATABASE_URL") == "" {
		return fmt.Errorf("DATABASE_URL must be set in production")
	}
	// 本番は複数レプリカで動かすため、配信などを Redis で共有しないと他のレプリカの接続に届かない
	if c.Server.Env == "production" && c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL must be set in production")
	}
	return nil
}
