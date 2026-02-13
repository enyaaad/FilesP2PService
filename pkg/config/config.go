package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Storage  StorageConfig
	WebRTC   WebRTCConfig
}

type ServerConfig struct {
	Port          string
	GRPCPort      string
	WebSocketPort string
	Environment   string
	JWTSecret     string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type StorageConfig struct {
	Provider   string
	LocalPath  string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
}

type WebRTCConfig struct {
	STUNServers  []string
	TURNServers  []string
	TURNPort     string // UDP порт для TURN сервера
	TURNUsername string // Username для TURN аутентификации
	TURNPassword string // Password для TURN аутентификации
	TURNRealm    string // Realm для TURN
	TURNPublicIP string // Публичный IP адрес TURN сервера (для relay)
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:          getEnv("SERVER_PORT", "8080"),
			GRPCPort:      getEnv("GRPC_PORT", "9090"),
			WebSocketPort: getEnv("WS_PORT", "8081"),
			Environment:   getEnv("ENV", "development"),
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Storage: StorageConfig{
			Provider:   getEnv("STORAGE_PROVIDER", "local"),
			LocalPath:  getEnv("STORAGE_LOCAL_PATH", "./storage"),
			Endpoint:   getEnv("STORAGE_ENDPOINT", ""),
			AccessKey:  getEnv("STORAGE_ACCESS_KEY", ""),
			SecretKey:  getEnv("STORAGE_SECRET_KEY", ""),
			BucketName: getEnv("STORAGE_BUCKET", "files"),
		},
		WebRTC: WebRTCConfig{
			STUNServers: []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
			},
			TURNServers:  []string{},
			TURNPort:     getEnv("TURN_PORT", "3478"),
			TURNUsername: getEnv("TURN_USERNAME", "user"),
			TURNPassword: getEnv("TURN_PASSWORD", "password"),
			TURNRealm:    getEnv("TURN_REALM", "local"),
			TURNPublicIP: getEnv("TURN_PUBLIC_IP", ""),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
