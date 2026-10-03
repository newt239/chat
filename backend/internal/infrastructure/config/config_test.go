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
	if cfg.Database.MaxIdleConns != 5 || cfg.Database.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("未設定の項目が既定値になっていません: %+v", cfg.Database)
	}
}

func TestScheduledMessageDispatchInterval(t *testing.T) {
	cfg := Load()
	if cfg.ScheduledMessage.DispatchInterval != 10*time.Second {
		t.Errorf("既定値が 10 秒になっていません: %v", cfg.ScheduledMessage.DispatchInterval)
	}
	t.Setenv("SCHEDULED_MESSAGE_DISPATCH_INTERVAL", "15m")
	if cfg = Load(); cfg.ScheduledMessage.DispatchInterval != 15*time.Minute {
		t.Errorf("環境変数が反映されていません: %v", cfg.ScheduledMessage.DispatchInterval)
	}
}

func TestValidateRequiresRedis(t *testing.T) {
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("PASSWORD_AUTH_ENABLED", "true")
	t.Setenv("REDIS_URL", "")
	cfg := Load()
	if err := cfg.Validate(); err == nil {
		t.Fatal("REDIS_URL が空なら検証に失敗するはず")
	}
	cfg.Redis.URL = "redis://localhost:6379"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("REDIS_URL があれば通るはず: %v", err)
	}
}
