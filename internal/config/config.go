package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type HTTPConfig struct {
	Host              string
	Port              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	ConnectTimeout  time.Duration
}

type JWTConfig struct {
	Secret string
}

func Load() (*Config, error) {
	// Настоящие переменные окружения имеют приоритет
	// над значениями из .env.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	cfg := &Config{
		HTTP: HTTPConfig{
			Host: getEnv("HTTP_HOST", "127.0.0.1"),
			Port: getEnv("HTTP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
	}

	var err error

	if cfg.HTTP.ReadHeaderTimeout, err = getDuration(
		"HTTP_READ_HEADER_TIMEOUT", 5*time.Second,
	); err != nil {
		return nil, err
	}

	if cfg.HTTP.ReadTimeout, err = getDuration(
		"HTTP_READ_TIMEOUT", 15*time.Second,
	); err != nil {
		return nil, err
	}

	if cfg.HTTP.WriteTimeout, err = getDuration(
		"HTTP_WRITE_TIMEOUT", 60*time.Second,
	); err != nil {
		return nil, err
	}

	if cfg.HTTP.IdleTimeout, err = getDuration(
		"HTTP_IDLE_TIMEOUT", 60*time.Second,
	); err != nil {
		return nil, err
	}

	if cfg.HTTP.ShutdownTimeout, err = getDuration(
		"HTTP_SHUTDOWN_TIMEOUT", 10*time.Second,
	); err != nil {
		return nil, err
	}

	if cfg.Database.MaxConns, err = getInt32(
		"DB_MAX_CONNS", 10,
	); err != nil {
		return nil, err
	}

	if cfg.Database.MinConns, err = getInt32(
		"DB_MIN_CONNS", 1,
	); err != nil {
		return nil, err
	}

	if cfg.Database.MaxConnLifetime, err = getDuration(
		"DB_MAX_CONN_LIFETIME", time.Hour,
	); err != nil {
		return nil, err
	}

	if cfg.Database.ConnectTimeout, err = getDuration(
		"DB_CONNECT_TIMEOUT", 5*time.Second,
	); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if net.ParseIP(c.HTTP.Host) == nil &&
		c.HTTP.Host != "localhost" {
		return fmt.Errorf("invalid HTTP_HOST")
	}

	port, err := strconv.Atoi(c.HTTP.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}

	if strings.TrimSpace(c.Database.URL) == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	if c.Database.MinConns < 0 ||
		c.Database.MaxConns < 1 ||
		c.Database.MinConns > c.Database.MaxConns {
		return fmt.Errorf("invalid database pool limits")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}

	return nil
}

func (c *Config) HTTPAddress() string {
	return net.JoinHostPort(c.HTTP.Host, c.HTTP.Port)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}

	return duration, nil
}

func getInt32(key string, fallback int32) (int32, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	number, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}

	return int32(number), nil
}
