package config

import (
	"testing"
	"time"
)

func TestPasswordAuthEnabledDefault(t *testing.T) {
	tests := []struct {
		name string
		env  string
		flag string
		want bool
	}{
		{name: "開発環境では既定で有効", env: "development", want: true},
		{name: "本番環境では既定で無効", env: "production", want: false},
		{name: "本番環境でも明示すれば有効", env: "production", flag: "true", want: true},
		{name: "開発環境でも明示すれば無効", env: "development", flag: "false", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENV", tt.env)
			t.Setenv("PASSWORD_AUTH_ENABLED", tt.flag)
			cfg := Load()
			if cfg.Auth.PasswordAuthEnabled != tt.want {
				t.Errorf("PasswordAuthEnabled = %v, want %v", cfg.Auth.PasswordAuthEnabled, tt.want)
			}
		})
	}
}

func TestDatabasePoolFromEnv(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "4")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "30s")
	cfg := Load()
	if cfg.Database.MaxOpenConns != 4 || cfg.Database.ConnMaxIdleTime != 30*time.Second {
		t.Errorf("環境変数が反映されていません: %+v", cfg.Database)
	}
	if cfg.Database.MaxIdleConns != 5 {
		t.Errorf("未設定の項目が既定値になっていません: %+v", cfg.Database)
	}
}

func TestDispatchInterval(t *testing.T) {
	cfg := Load()
	if cfg.DispatchInterval != 10*time.Second {
		t.Errorf("既定値が 10 秒になっていません: %v", cfg.DispatchInterval)
	}
	t.Setenv("DISPATCH_INTERVAL", "15m")
	if cfg = Load(); cfg.DispatchInterval != 15*time.Minute {
		t.Errorf("環境変数が反映されていません: %v", cfg.DispatchInterval)
	}
}

func TestValidateRequiresConnections(t *testing.T) {
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("PASSWORD_AUTH_ENABLED", "true")
	required := map[string]string{
		"DATABASE_URL":    "postgres://db:5432/chat",
		"MEILISEARCH_URL": "http://meilisearch:7700",
		"REDIS_URL":       "redis://redis:6379",
		"PUBLIC_BASE_URL": "https://api.example.com",
	}
	for name, value := range required {
		t.Setenv(name, value)
	}
	if err := Load().Validate(); err != nil {
		t.Fatalf("接続先がそろっていれば通るはず: %v", err)
	}
	for name := range required {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, "")
			if err := Load().Validate(); err == nil {
				t.Fatalf("%s が空なら検証に失敗するはず", name)
			}
		})
	}
}
