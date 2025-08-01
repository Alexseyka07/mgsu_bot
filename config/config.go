package config

import "os"

var (
	TG_TOKEN = getEnv("TG_TOKEN", "1")
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
