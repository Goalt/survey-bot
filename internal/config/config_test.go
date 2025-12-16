package config

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		want    Config
		wantErr bool
		envs    map[string]string
	}{
		{
			name: "ok",
			want: Config{
				Level: "error",
				Env:   "prod",
				Token: "abc",
				DB: DatabaseConfig{
					Host:         "localhost",
					Port:         "1111",
					Name:         "survey-bot",
					User:         "user",
					Pwd:          "pwd",
					SslMode:      "enable",
					MigrationsUp: false,
				},
				ReleaseVersion: "1.0.0",
				PollInterval:   10 * time.Minute,
				SentryTimeout:  5 * time.Second,
				AdminUserIDs:   []int64{-1},
				MetricsPort:    7777,
				APIPort:        8080,
			},
			wantErr: false,
			envs: map[string]string{
				"LEVEL":              "error",
				"ENV":                "prod",
				"TOKEN":              "abc",
				"DB_HOST":            "localhost",
				"DB_PORT":            "1111",
				"DB_NAME":            "survey-bot",
				"DB_USER":            "user",
				"DB_PWD":             "pwd",
				"DB_SSL_MODE":        "enable",
				"DB_SCHEMA":          "public2",
				"DB_MIGRATIONS_UP":   "false",
				"DB_MIGRATIONS_DOWN": "true",
				"POLL_DURATION":      "10m",
				"RELEASE_VERSION":    "1.0.0",
			},
		},
		{
			name:    "no passwd",
			want:    Config{},
			wantErr: true,
			envs: map[string]string{
				"LEVEL":              "error",
				"ENV":                "prod",
				"TOKEN":              "abc",
				"DB_HOST":            "localhost",
				"DB_PORT":            "1111",
				"DB_NAME":            "survey-bot",
				"DB_USER":            "user",
				"DB_SSL_MODE":        "enable",
				"DB_SCHEMA":          "public2",
				"DB_MIGRATIONS_UP":   "false",
				"DB_MIGRATIONS_DOWN": "true",
				"POLL_DURATION":      "10m",
			},
		},
		{
			name:    "fail, no release version",
			wantErr: true,
			envs: map[string]string{
				"LEVEL":              "error",
				"ENV":                "prod",
				"TOKEN":              "abc",
				"DB_HOST":            "localhost",
				"DB_PORT":            "1111",
				"DB_NAME":            "survey-bot",
				"DB_USER":            "user",
				"DB_PWD":             "pwd",
				"DB_SSL_MODE":        "enable",
				"DB_SCHEMA":          "public2",
				"DB_MIGRATIONS_UP":   "false",
				"DB_MIGRATIONS_DOWN": "true",
				"POLL_DURATION":      "10m",
			},
		},
		{
			name: "ok, migrations up",
			want: Config{
				Level:        "error",
				Env:          "prod",
				Token:        "abc",
				AdminUserIDs: []int64{-1},
				DB: DatabaseConfig{
					Host:         "localhost",
					Port:         "1111",
					Name:         "survey-bot",
					User:         "user",
					Pwd:          "pwd",
					SslMode:      "enable",
					MigrationsUp: true,
				},
				ReleaseVersion: "1.0.0",
				PollInterval:   10 * time.Minute,
				SentryTimeout:  5 * time.Second,
				MetricsPort:    7777,
				APIPort:        8080,
			},
			wantErr: false,
			envs: map[string]string{
				"LEVEL":              "error",
				"ENV":                "prod",
				"TOKEN":              "abc",
				"DB_HOST":            "localhost",
				"DB_PORT":            "1111",
				"DB_NAME":            "survey-bot",
				"DB_USER":            "user",
				"DB_PWD":             "pwd",
				"DB_SSL_MODE":        "enable",
				"DB_SCHEMA":          "public2",
				"DB_MIGRATIONS_UP":   "true",
				"DB_MIGRATIONS_DOWN": "true",
				"POLL_DURATION":      "10m",
				"RELEASE_VERSION":    "1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envs {
				_ = os.Setenv(key, val)
			}

			got, err := New()
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}

			for key := range tt.envs {
				_ = os.Setenv(key, "")
			}
		})
	}
}
