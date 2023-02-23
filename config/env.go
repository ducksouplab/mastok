package config

import (
	"os"
)

func GetEnvOr(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
