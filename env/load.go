package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// own variables
var AsCommandLine bool
var BasicLogin, BasicPassword, Mode, Origin, Port, ProjectRoot, WebPrefix string
var AllowedOrigins []string

// connection to other services
var DatabaseURL, OTreeURL, OTreeKey string

// CAUTION: other init functions in "config" package may be called before this
func init() {
	Mode = GetEnvOr("MASTOK_MODE", "PROD")
	if Mode == "DEV" {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal(err)
		}
	}
	if Mode == "BUILD_FRONT" || Mode == "RESET_DEV" {
		AsCommandLine = true
	}
	Port = GetEnvOr("MASTOK_PORT", "8190")
	Origin = GetEnvOr("MASTOK_ORIGIN", "http://localhost:8190")
	WebPrefix = GetEnvOr("MASTOK_WEB_PREFIX", "")
	ProjectRoot = GetEnvOr("MASTOK_PROJECT_ROOT", ".") + "/"
	BasicLogin = GetEnvOr("MASTOK_LOGIN", "mastok")
	BasicPassword = GetEnvOr("MASTOK_PASSWORD", "mastok")
	// no default value provided
	DatabaseURL = os.Getenv("MASTOK_DATABASE_URL")
	OTreeURL = os.Getenv("MASTOK_OTREE_URL")
	OTreeKey = os.Getenv("MASTOK_OTREE_REST_KEY")
	// derived
	AllowedOrigins = []string{Origin}
	if Mode == "ENV" {
		AllowedOrigins = append(AllowedOrigins, "127.0.0.1")
	}
}
