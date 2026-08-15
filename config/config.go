package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config agrupa todas as configurações da aplicação carregadas do ambiente.
type Config struct {
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load lê o arquivo .env (se existir) e monta a struct de configuração,
// aplicando valores padrão quando a variável não estiver definida.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "ecommerce"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// getEnv retorna o valor da variável de ambiente ou um fallback padrão.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
