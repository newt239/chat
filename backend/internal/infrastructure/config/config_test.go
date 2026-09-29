package config

import "testing"

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
