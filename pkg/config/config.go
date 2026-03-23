package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Server ServerConfig
	DB     DatabaseConfig
	JWT    JWTConfig
	App    AppConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret string
}

type AppConfig struct {
	Environment string
}

// Load carga las variables de entorno desde .env
func Load() (*Config, error) {
	// Cargar .env (opcional)
	if err := godotenv.Load(); err != nil {
		log.Println("[CONFIG] No se encontró .env, usando variables del sistema")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "coviar_user"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "coviar_db"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your_jwt_secret_key_here"),
		},
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
		},
	}

	if cfg.DB.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD es requerida")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
