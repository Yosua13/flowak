package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port         string
	JWTSecret    string
	GeminiAPIKey string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
}

var ActiveConfig Config

// LoadEnv reads the .env file if it exists and sets environment variables.
func LoadEnv(envPath string) {
	file, err := os.Open(envPath)
	if err != nil {
		return // File does not exist, fallback to system env
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Strip quotes if present
			if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
				(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
				val = val[1 : len(val)-1]
			}
			os.Setenv(key, val)
		}
	}
}

// InitConfig initializes the global ActiveConfig.
func InitConfig() {
	LoadEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_flowak_secret_key"
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	// PostgreSQL Env Vars with default fallbacks
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "flowak"
	}

	ActiveConfig = Config{
		Port:         port,
		JWTSecret:    jwtSecret,
		GeminiAPIKey: geminiAPIKey,
		DBHost:       dbHost,
		DBPort:       dbPort,
		DBUser:       dbUser,
		DBPassword:   dbPassword,
		DBName:       dbName,
	}
}
