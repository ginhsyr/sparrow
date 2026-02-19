package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type databaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
}
type serverConfig struct {
	Port              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
	RateLimitPerMin   int
	RateLimitBurst    int
	CORSAllowOrigins  []string
}

var DBConfig databaseConfig
var ServerConfig serverConfig
var JWTSigningKey []byte
var LogLevel string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to read configuration file: %v", err)
	}

	DBConfig = databaseConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Port:     os.Getenv("DB_PORT"),
	}

	ServerConfig.Port = getEnv("PORT", "8025")
	ServerConfig.ReadTimeout = time.Duration(getPositiveInt("HTTP_READ_TIMEOUT_SEC", 10)) * time.Second
	ServerConfig.WriteTimeout = time.Duration(getPositiveInt("HTTP_WRITE_TIMEOUT_SEC", 15)) * time.Second
	ServerConfig.IdleTimeout = time.Duration(getPositiveInt("HTTP_IDLE_TIMEOUT_SEC", 60)) * time.Second
	ServerConfig.ReadHeaderTimeout = time.Duration(getPositiveInt("HTTP_READ_HEADER_TIMEOUT_SEC", 5)) * time.Second
	ServerConfig.ShutdownTimeout = time.Duration(getPositiveInt("SHUTDOWN_TIMEOUT_SEC", 10)) * time.Second
	ServerConfig.RateLimitPerMin = getPositiveInt("RATE_LIMIT_PER_MIN", 120)
	ServerConfig.RateLimitBurst = getPositiveInt("RATE_LIMIT_BURST", 20)
	ServerConfig.CORSAllowOrigins = parseCSVEnv("CORS_ALLOW_ORIGINS", "*")
	jwtSigningKey := strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY"))
	if len(jwtSigningKey) < 32 {
		log.Fatal("JWT_SIGNING_KEY must be at least 32 characters")
	}
	JWTSigningKey = []byte(jwtSigningKey)
	LogLevel = os.Getenv("LOG_LEVEL")
}

func getEnv(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func getPositiveInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		log.Printf("Invalid value for %s, fallback to default %d", key, defaultValue)
		return defaultValue
	}
	return parsed
}

func parseCSVEnv(key, defaultValue string) []string {
	raw := getEnv(key, defaultValue)
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item != "" {
			origins = append(origins, item)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}
