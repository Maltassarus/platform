package config

import (
    "fmt"
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    DBHost               string
    DBPort               string
    DBUser               string
    DBPassword           string
    DBName               string
    DBSSLMode            string
    JWTSecret            string
    ServerPort           string
    PublishCheckInterval int
    WorkerPoolSize       int
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    return &Config{
        DBHost:               getEnv("DB_HOST", "localhost"),
        DBPort:               getEnv("DB_PORT", "5432"),
        DBUser:               getEnv("DB_USER", "postgres"),
        DBPassword:           getEnv("DB_PASSWORD", "postgres"),
        DBName:               getEnv("DB_NAME", "blog"),
        DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
        JWTSecret:            getEnv("JWT_SECRET", "your-secret-key"),
        ServerPort:           getEnv("SERVER_PORT", "8080"),
        PublishCheckInterval: getEnvAsInt("PUBLISH_CHECK_INTERVAL", 10),
        WorkerPoolSize:       getEnvAsInt("WORKER_POOL_SIZE", 5),
    }
}

func (c *Config) GetDBConnString() string {
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}