package config

import "os"

var (
	TG_TOKEN = getEnv("TG_TOKEN", "8276926837:AAHiD4nlbYcJokRpotch27-uhmhgHyKsz90")
	DB_HOST  = getEnv("DB_HOST", "localhost")
	DB_PORT  = getEnv("DB_PORT", "5432")
	DB_USER  = getEnv("DB_USER", "postgres")
	DB_NAME  = getEnv("DB_NAME", "fisher")
	DB_PASS  = getEnv("DB_PASS", "pappIw-4cabqa-qynnat")
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
