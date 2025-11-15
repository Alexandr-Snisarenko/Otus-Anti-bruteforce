package config

import (
	"errors"
	"os"
	"strings"
	"time"

	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var ErrFileNotFound = errors.New(" file not found")

type App struct {
	Name string `mapstructure:"name"`
}

type Limits struct {
	LoginAttempts    int           `mapstructure:"login_attempts"`
	PasswordAttempts int           `mapstructure:"password_attempts"`
	IPAttempts       int           `mapstructure:"ip_attempts"`
	Window           time.Duration `mapstructure:"window"`
}

type Server struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
	TLS     struct {
		Enabled  bool   `mapstructure:"enabled"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
	} `mapstructure:"tls"`
}

type Logger struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

type Database struct {
	Workmode   string `mapstructure:"workmode"` // Режим работы memory или postgresql по умолчанию - postgresql
	Postgresql struct {
		// Параметры подключения могут задаваться либо в dsn, либо, если dsn не задан в следующих полях
		Dsn string `mapstructure:"dsn"`
		// Поля подключения к БД в случае, если dsn не задан
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Name     string `mapstructure:"name"`
		// параметры пула коннектов
		Pool struct {
			// Макс. число открытых соединений от этого процесса (по умолчанию - 20, без ограничений)
			MaxOpenConns int `mapstructure:"max_open_conns"`
			// Макс. число открытых неиспользуемых соединений
			MaxIdleConns int `mapstructure:"max_idle_conns"`
			// Макс. время жизни одного подключения
			ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
			// Макс. время ожидания подключения в пуле
			ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
		} `mapstructure:"pool"`
	} `mapstructure:"postgresql"`
}

type Config struct {
	App      App      `mapstructure:"app"`
	Limits   Limits   `mapstructure:"limits"`
	Server   Server   `mapstructure:"server"`
	Logger   Logger   `mapstructure:"logger"`
	Database Database `mapstructure:"database"`
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func setDefaults(v *viper.Viper) {
	// Дефолты
	v.SetDefault("database.workmode", "postgresql")
	v.SetDefault("database.postgresql.pool.max_open_conns", 20)
	v.SetDefault("database.postgresql.pool.max_idle_conns", 10)
	v.SetDefault("database.postgresql.pool.conn_max_lifetime", "1h")
	v.SetDefault("database.postgresql.pool.conn_max_idle_time", "10m")
	v.SetDefault("server.port", 8080)
	v.SetDefault("logger.level", "info")
	v.SetDefault("limits.login_attempts", 10)
	v.SetDefault("limits.password_attempts", 100)
	v.SetDefault("limits.ip_attempts", 1000)
	v.SetDefault("limits.window", "1m")

	// Бинды/ для работы без файла конфигурациии без дефолтов, или с нестандартными ключами окружения
	// _ = v.BindEnv("logger.level", "RATELIMITER_LOGGER__LEVEL")
}

func LoadConfig(cfgFilePath string) (*Config, error) {
	v := viper.New()

	// ENV с префиксом RATELIMITER
	v.SetEnvPrefix("RATELIMITER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__", "-", "_"))
	v.AutomaticEnv()

	// устанавливаем дефолты и бинды для загрузки из ENV
	setDefaults(v)

	// если конфиг не задан - ищем по стандартным путям
	if cfgFilePath == "" {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/calendar")
	} else {
		if !fileExists(cfgFilePath) {
			return nil, ErrFileNotFound
		}
		v.SetConfigFile(cfgFilePath)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// как вариант - без контроля наличия файла
	// _ = v.ReadInConfig()

	var cfg Config
	decoderCfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           &cfg,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
	}
	dec, err := mapstructure.NewDecoder(decoderCfg)
	if err != nil {
		return nil, err
	}
	if err := dec.Decode(v.AllSettings()); err != nil {
		return nil, err
	}
	return &cfg, nil
}
