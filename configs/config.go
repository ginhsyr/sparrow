package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

type databaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
}
type serverConfig struct {
	Port string
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

	ServerConfig.Port = os.Getenv("PORT")
	jwtSigningKey := strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY"))
	if len(jwtSigningKey) < 32 {
		log.Fatal("JWT_SIGNING_KEY must be at least 32 characters")
	}
	JWTSigningKey = []byte(jwtSigningKey)
	LogLevel = os.Getenv("LOG_LEVEL")
}
