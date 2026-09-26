package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	HTTPPort     string
	PostgresDSN  string
	RedisAddr    string
	RedisDB      int
	AdminToken   string
	RateLimitRPS int
}

func Load() (*Config, error) {
	rateLimitRPS, err := strconv.Atoi(os.Getenv("RATE_LIMIT_RPS"))
	if err != nil {
		return nil, errors.New("RATE_LIMIT_RPS must be an integer")
	}

	cfg := &Config{
		HTTPPort:     os.Getenv("HTTP_PORT"),
		PostgresDSN:  os.Getenv("POSTGRES_DSN"),
		RedisAddr:    os.Getenv("REDIS_ADDR"),
		RedisDB:      0,
		AdminToken:   os.Getenv("ADMIN_TOKEN"),
		RateLimitRPS: rateLimitRPS,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.HTTPPort == "" {
		return errors.New("HTTP_PORT is required")
	}

	if c.PostgresDSN == "" {
		return errors.New("POSTGRES_DSN is required")
	}

	if c.RedisAddr == "" {
		return errors.New("REDIS_ADDR is required")
	}

	if c.AdminToken == "" {
		return errors.New("ADMIN_TOKEN is required")
	}

	if c.RedisDB < 0 {
		return errors.New("REDIS_DB must be greater than or equal to 0")
	}

	if c.RateLimitRPS <= 0 {
		return errors.New("RATE_LIMIT_RPS must be greater than 0")
	}

	return nil
}
