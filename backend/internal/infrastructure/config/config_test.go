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
			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Auth.PasswordAuthEnabled != tt.want {
				t.Errorf("PasswordAuthEnabled = %v, want %v", cfg.Auth.PasswordAuthEnabled, tt.want)
			}
		})
	}
}

func TestDatabasePoolFromEnv(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "4")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "30s")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.MaxOpenConns != 4 || cfg.Database.ConnMaxIdleTime != 30*time.Second {
		t.Errorf("環境変数が反映されていません: %+v", cfg.Database)
	}
	if cfg.Database.MaxIdleConns != 5 || cfg.Database.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("未設定の項目が既定値になっていません: %+v", cfg.Database)
	}
}

func TestValidateRequiresRedisInProduction(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DATABASE_URL", "postgres://localhost/chat")
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("PASSWORD_AUTH_ENABLED", "true")
	t.Setenv("REDIS_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("本番で REDIS_URL が空なら検証に失敗するはず")
	}
	cfg.Redis.URL = "redis://localhost:6379"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("REDIS_URL があれば通るはず: %v", err)
	}
}
