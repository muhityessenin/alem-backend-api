package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Host, Env                              string
		Port                                   int
		ReadTimeout, WriteTimeout, IdleTimeout time.Duration
	}
	DB struct {
		URL      string
		MaxConns int32
	}
	JWT struct {
		AccessSecret string
	}
	CORS struct {
		AllowedOrigins []string
		AllowedMethods []string
		AllowedHeaders []string
	}
}

func MustLoad() Config {
	_ = godotenv.Load()
	var c Config
	c.App.Host = env("HOST", "0.0.0.0")
	c.App.Env = env("APP_ENV", "dev")
	c.App.Port = envInt("PORT", 8082)
	c.App.ReadTimeout = envDur("HTTP_READ_TIMEOUT", "10s")
	c.App.WriteTimeout = envDur("HTTP_WRITE_TIMEOUT", "15s")
	c.App.IdleTimeout = envDur("HTTP_IDLE_TIMEOUT", "60s")

	c.DB.URL = must("DB_URL")
	c.DB.MaxConns = int32(envInt("DB_MAX_CONNS", 20))

	c.JWT.AccessSecret = must("JWT_SECRET")

	c.CORS.AllowedOrigins = envList("CORS_ALLOWED_ORIGINS", "*")
	c.CORS.AllowedMethods = envList("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	c.CORS.AllowedHeaders = envList("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")
	return c
}

func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
func must(k string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		log.Fatalf("%s missing", k)
	}
	return v
}
func envInt(k string, d int) int {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return d
}
func envDur(k, d string) time.Duration {
	s := env(k, d)
	dur, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("bad duration %s=%s", k, s)
	}
	return dur
}
func envList(k, d string) []string {
	s := env(k, d)
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
