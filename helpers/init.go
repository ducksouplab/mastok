package helpers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// CAUTION: other init functions in "helpers" package may be called before this
	if os.Getenv("MASTOK_ENV") == "DEV" {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal(err)
		}
	}
}
