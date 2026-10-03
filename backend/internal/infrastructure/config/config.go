package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
	// ScheduledMessage は予約投稿の送信処理。間隔を空けると DB にアクセスしない時間ができ、Neon などが停止できる
	ScheduledMessage ScheduledMessageConfig
}

type ScheduledMessageConfig struct {
	DispatchInterval time.Duration
}

// RedisConfig はレプリカ間で WebSocket の配信・閲覧者一覧・レート制限を共有する Redis
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
	// local のときに署名付き URL に使うバックエンドの公開 URL
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
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	Secret string
}

// AuthConfig は Google ログインとパスワード認証の設定。パスワード認証は production では既定で無効
type AuthConfig struct {
	GoogleOAuthClientID string
	PasswordAuthEnabled bool
	// ネイティブアプリの Google ログイン用。ClientSecret と RedirectURL が空ならネイティブアプリでは Google ログインを使えない
	GoogleOAuthClientSecret string
	GoogleOAuthRedirectURL  string
	NativeAppRedirectURL    string
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

func Load() *Config {
	env := getEnv("ENV", "development")
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  env,
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
		},
		Auth: AuthConfig{
			GoogleOAuthClientID: getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
			PasswordAuthEnabled: getEnvBool("PASSWORD_AUTH_ENABLED", env != "production"),

			GoogleOAuthClientSecret: getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
			GoogleOAuthRedirectURL:  getEnv("GOOGLE_OAUTH_REDIRECT_URL", ""),
			NativeAppRedirectURL:    getEnv("NATIVE_APP_REDIRECT_URL", "dev.newt239.chat://auth/callback"),
		},
		Storage: StorageConfig{
			Driver:        getEnv("STORAGE_DRIVER", "wasabi"),
			PublicBaseURL: getEnv("PUBLIC_BASE_URL", ""),
		},
		Wasabi: WasabiConfig{
			Endpoint:        getEnv("WASABI_ENDPOINT", "https://s3.wasabisys.com"),
			Region:          getEnv("WASABI_REGION", "us-east-1"),
			BucketName:      getEnv("WASABI_BUCKET", "chat-attachments"),
			AccessKeyID:     getEnv("WASABI_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("WASABI_SECRET_ACCESS_KEY", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS"),
		},
		Search: SearchConfig{
			MeilisearchURL:    getEnv("MEILISEARCH_URL", ""),
			MeilisearchAPIKey: getEnv("MEILISEARCH_API_KEY", ""),
		},
		Firebase: FirebaseConfig{
			ProjectID: getEnv("FIREBASE_PROJECT_ID", ""),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
		},
		ScheduledMessage: ScheduledMessageConfig{
			DispatchInterval: getEnvDuration("SCHEDULED_MESSAGE_DISPATCH_INTERVAL", 10*time.Second),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvList はカンマ区切りの環境変数をリストとして読み込みます
func getEnvList(key string) []string {
	values := strings.Split(os.Getenv(key), ",")
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
	for name, value := range map[string]string{"DATABASE_URL": c.Database.URL, "MEILISEARCH_URL": c.Search.MeilisearchURL, "REDIS_URL": c.Redis.URL} {
		if value == "" {
			return fmt.Errorf("%s must be set", name)
		}
	}
	return nil
}
