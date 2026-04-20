package config

import "os"

type Config struct {
	Port           string
	DBUrl          string
	JWTSecret      string
	ClientAPIKey   string
	GoogleClientID string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DBUrl:          getEnv("DB", ""),
		JWTSecret:      getEnv("SECRET", ""),
		ClientAPIKey:   getEnv("CLIENT_API_KEY", ""),
		GoogleClientID: getEnv("CLIENGOOGLE_CLIENT_IDT_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
