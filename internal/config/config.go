package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     mustEnv("DB_HOST"),
			Port:     mustEnv("DB_PORT"),
			User:     mustEnv("DB_USER"),
			Password: mustEnv("DB_PASSWORD"),
			Name:     mustEnv("DB_NAME"),
		},
		Redis: RedisConfig{
			Addr:     mustEnv("REDIS_ADDR"),
			Password: mustEnv("REDIS_PASSWORD"),
			DB:       mustEnvInt("REDIS_DB"),
		},
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("ENV variable %s is required but not set", key)
	}
	return val
}

func mustEnvInt(key string) int {
	valStr := mustEnv(key)
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("ENV variable %s must be an integer, got: %s", key, valStr)
	}
	return val
}
