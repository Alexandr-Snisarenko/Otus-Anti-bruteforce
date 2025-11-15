package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

// загрузка файла конфигурации.
func TestLoadConfig_FromYAML(t *testing.T) {
	t.Parallel()

	yaml := `
app:
  name: "calendar"

server:
  address: "0.0.0.0"
  port: 8080
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

logger:
  level: "info"
  file: "/var/log/calendar.log"

database:
  workmode: "postgresql"
  postgresql:
    dsn: ""
    user: "user1"
    password: "pass1"
    host: "localhost"
    port: 5432
    name: "caldb"
    pool:
      max_open_conns: 20
      max_idle_conns: 10
      conn_max_lifetime: "1h"
      conn_max_idle_time: "5m"
`

	dir := t.TempDir()
	cfgPath := writeTempFile(t, dir, "config.yaml", yaml)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if cfg.App.Name != "calendar" {
		t.Errorf("App.Name = %q, want %q", cfg.App.Name, "calendar")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Database.Workmode != "postgresql" {
		t.Errorf("Database.Workmode = %q, want %q", cfg.Database.Workmode, "postgresql")
	}
	if got, want := cfg.Database.Postgresql.Pool.MaxOpenConns, 20; got != want {
		t.Errorf("MaxOpenConns = %d, want %d", got, want)
	}
	if got, want := cfg.Database.Postgresql.Pool.ConnMaxLifetime, time.Hour; got != want {
		t.Errorf("ConnMaxLifetime = %v, want %v", got, want)
	}
	if got, want := cfg.Database.Postgresql.Pool.ConnMaxIdleTime, 5*time.Minute; got != want {
		t.Errorf("ConnMaxIdleTime = %v, want %v", got, want)
	}
}

// работа с переменными окружения.
func TestLoadConfig_EnvOverride_DSN(t *testing.T) {
	// Пустой файл, всё переопределим ENV
	// logger.level в файле не задан. для связки с переменной окружения в вайпере используется defoult
	dir := t.TempDir()

	cfgPath := writeTempFile(t, dir, "config.yaml", "database: { workmode: postgresql, postgresql: { dsn: \"\" } }")

	// ENV: RATELIMITER_DATABASE__POSTGRESQL__DSN
	t.Setenv("RATELIMITER_DATABASE__POSTGRESQL__DSN", "postgres://u:p@db:5432/x?sslmode=disable")
	t.Setenv("RATELIMITER_LOGGER__LEVEL", "debug")
	t.Setenv("RATELIMITER_SERVER__PORT", "9090")

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if got, want := cfg.Database.Postgresql.Dsn, "postgres://u:p@db:5432/x?sslmode=disable"; got != want {
		t.Errorf("DSN = %q, want %q", got, want)
	}
	if got, want := cfg.Logger.Level, "debug"; got != want {
		t.Errorf("Logger.Level = %q, want %q", got, want)
	}
	if got, want := cfg.Server.Port, 9090; got != want {
		t.Errorf("Server.Port = %d, want %d", got, want)
	}
}

// обработка типа Time.Duration.
func TestLoadConfig_DurationParsing(t *testing.T) {
	t.Parallel()

	yaml := `
database:
  postgresql:
    pool:
      conn_max_lifetime: "30m"
      conn_max_idle_time: "0s"
`
	dir := t.TempDir()
	cfgPath := writeTempFile(t, dir, "config.yaml", yaml)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if got, want := cfg.Database.Postgresql.Pool.ConnMaxLifetime, 30*time.Minute; got != want {
		t.Errorf("ConnMaxLifetime = %v, want %v", got, want)
	}
	if got, want := cfg.Database.Postgresql.Pool.ConnMaxIdleTime, 0*time.Second; got != want {
		t.Errorf("ConnMaxIdleTime = %v, want %v", got, want)
	}
}
