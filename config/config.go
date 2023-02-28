package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var OwnEnv, OwnPort, OwnOrigin, OwnWebPrefix, OwnRoot, OwnBasicLogin, OwnBasicPassword, DatabaseURL, OTreeURL, OTreeKey string

// CAUTION: other init functions in "config" package may be called before this
func init() {
	OwnEnv = GetEnvOr("MASTOK_ENV", "PROD")
	if OwnEnv == "DEV" {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal(err)
		}
	}
	OwnPort = GetEnvOr("MASTOK_PORT", "8190")
	OwnOrigin = GetEnvOr("MASTOK_ORIGIN", "http://localhost:8190")
	OwnWebPrefix = GetEnvOr("MASTOK_WEB_PREFIX", "/")
	OwnRoot = GetEnvOr("MASTOK_PROJECT_ROOT", ".") + "/"
	OwnBasicLogin = GetEnvOr("MASTOK_LOGIN", "mastok")
	OwnBasicPassword = GetEnvOr("MASTOK_PASSWORD", "mastok")
	// no default value provided
	DatabaseURL = os.Getenv("MASTOK_DATABASE_URL")
	OTreeURL = os.Getenv("MASTOK_OTREE_URL")
	OTreeKey = os.Getenv("MASTOK_OTREE_REST_KEY")
}
