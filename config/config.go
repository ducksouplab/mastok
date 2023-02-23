package config

import (
	"log"

	"github.com/joho/godotenv"
)

var OwnEnv, OwnPort, DatabaseURL, OTreeURL, OTreeKey, ProjectRoot, AuthBasicLogin, AuthBasicPassword string

// CAUTION: other init functions in "config" package may be called before this
func init() {
	OwnEnv = GetEnvOr("MASTOK_ENV", "PROD")
	if OwnEnv == "DEV" {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal(err)
		}
	}
	OwnPort = GetEnvOr("MASTOK_PORT", "8190")
	DatabaseURL = GetEnvOr("MASTOK_DATABASE_URL", "")
	OTreeURL = GetEnvOr("MASTOK_OTREE_URL", "http://localhost:8180")
	OTreeKey = GetEnvOr("MASTOK_OTREE_REST_KEY", "key")
	ProjectRoot = GetEnvOr("MASTOK_PROJECT_ROOT", ".") + "/"
	AuthBasicLogin = GetEnvOr("MASTOK_LOGIN", "admin")
	AuthBasicPassword = GetEnvOr("MASTOK_PASSWORD", "admin")
}
