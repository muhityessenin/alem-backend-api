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
		Name         string
		Env          string
		Host         string
		Port         int
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
	}
	DB struct {
		URL             string
		MaxConns        int32
		MinConns        int32
		MaxConnLifetime time.Duration
		MaxConnIdleTime time.Duration
	}
	JWT struct {
		AccessSecret  string
		RefreshSecret string
		AccessTTL     time.Duration
		RefreshTTL    time.Duration
		Issuer        string
		Audience      string
	}
	OTP struct {
		Enabled bool
		TTL     time.Duration
		Length  int
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

	// App
	c.App.Name = getEnv("APP_NAME", "auth")
	c.App.Env = getEnv("APP_ENV", "dev")
	c.App.Host = getEnv("HOST", "0.0.0.0")
	c.App.Port = getEnvInt("PORT", 8080)
	c.App.ReadTimeout = getEnvDur("HTTP_READ_TIMEOUT", "10s")
	c.App.WriteTimeout = getEnvDur("HTTP_WRITE_TIMEOUT", "15s")
	c.App.IdleTimeout = getEnvDur("HTTP_IDLE_TIMEOUT", "60s")

	// DB
	c.DB.URL = getEnvStrict("DB_URL") // required
	c.DB.MaxConns = int32(getEnvInt("DB_MAX_CONNS", 20))
	c.DB.MinConns = int32(getEnvInt("DB_MIN_CONNS", 2))
	c.DB.MaxConnLifetime = getEnvDur("DB_MAX_CONN_LIFETIME", "30m")
	c.DB.MaxConnIdleTime = getEnvDur("DB_MAX_CONN_IDLE_TIME", "5m")

	// JWT
	c.JWT.AccessSecret = getEnvStrict("JWT_ACCESS_SECRET") // required
	c.JWT.RefreshSecret = getEnv("JWT_REFRESH_SECRET", c.JWT.AccessSecret)
	c.JWT.AccessTTL = getEnvDur("JWT_ACCESS_TTL", "15m")
	c.JWT.RefreshTTL = getEnvDur("JWT_REFRESH_TTL", "720h") // 30 days
	c.JWT.Issuer = getEnv("JWT_ISSUER", "alem-auth")
	c.JWT.Audience = getEnv("JWT_AUDIENCE", "alem-clients")

	// OTP
	c.OTP.Enabled = getEnvBool("OTP_ENABLED", true)
	c.OTP.TTL = getEnvDur("OTP_TTL", "2m")
	c.OTP.Length = getEnvInt("OTP_LENGTH", 6)

	// CORS
	c.CORS.AllowedOrigins = getEnvList("CORS_ALLOWED_ORIGINS", "*")
	c.CORS.AllowedMethods = getEnvList("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	c.CORS.AllowedHeaders = getEnvList("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")

	// sanity checks
	if c.App.Port <= 0 {
		log.Fatal("invalid PORT")
	}
	if len(c.CORS.AllowedOrigins) == 0 {
		log.Fatal("CORS_ALLOWED_ORIGINS must not be empty")
	}
	return c
}

func getEnvStrict(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("%s is not set", key)
	}
	return v
}
func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
func getEnvInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return def
}
func getEnvBool(key string, def bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return def
}
func getEnvDur(key, def string) time.Duration {
	s := getEnv(key, def)
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("invalid duration for %s: %s", key, s)
	}
	return d
}
func getEnvList(key, def string) []string {
	val := getEnv(key, def)
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
