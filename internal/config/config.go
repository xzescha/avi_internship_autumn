// Package config содержит структуры конфигурации приложения и функцию их загрузки.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPPort = "8080"

	defaultDBHost     = "postgres_PR_db"
	defaultDBPort     = 5432
	defaultDBUser     = "admin_PR"
	defaultDBPassword = "appPass_QWERTY"
	defaultDBName     = "PR_serv"
	defaultDBSSLMode  = "disable"

	defaultDBMaxOpenConns    = 10
	defaultDBMaxIdleConns    = 5
	defaultDBConnMaxLifetime = 30 * time.Minute
)

// HTTPConfig содержит настройки HTTP-сервера.
type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DBConfig содержит настройки подключения к базе данных.
type DBConfig struct {
	// Либо готовый DSN из DATABASE_DSN,
	// либо поля ниже, из которых DSN будет собран. * DSN - имя источника данных. В нашем случае ссылка на БД
	DSN      string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Config агрегирует конфигурацию всех подсистем приложения.
type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
}

// DSNString возвращает строку подключения для database/sql.
func (c DBConfig) DSNString() string {
	if c.DSN != "" {
		return c.DSN
	}

	// postgres://user:pass@host:port/dbname?sslmode=mode
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// Load считывает конфиг из переменных окружения.
func Load() (Config, error) {
	httpCfg := HTTPConfig{
		Port:         getEnv("HTTP_PORT", defaultHTTPPort),
		ReadTimeout:  getDurationEnv("HTTP_READ_TIMEOUT", 5*time.Second),
		WriteTimeout: getDurationEnv("HTTP_WRITE_TIMEOUT", 5*time.Second),
		IdleTimeout:  getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
	}

	dbPort := getIntEnv("DB_PORT", defaultDBPort)
	dbMaxOpen := getIntEnv("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns)
	dbMaxIdle := getIntEnv("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns)
	dbConnLife := getDurationEnv("DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime)

	dbCfg := DBConfig{
		DSN:      os.Getenv("DATABASE_DSN"),
		Host:     getEnv("DB_HOST", defaultDBHost),
		Port:     dbPort,
		User:     getEnv("DB_USER", defaultDBUser),
		Password: getEnv("DB_PASSWORD", defaultDBPassword),
		Name:     getEnv("DB_NAME", defaultDBName),
		SSLMode:  getEnv("DB_SSLMODE", defaultDBSSLMode),

		MaxOpenConns:    dbMaxOpen,
		MaxIdleConns:    dbMaxIdle,
		ConnMaxLifetime: dbConnLife,
	}

	cfg := Config{
		HTTP: httpCfg,
		DB:   dbCfg,
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getIntEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
