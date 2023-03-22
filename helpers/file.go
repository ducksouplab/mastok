package helpers

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/ducksouplab/mastok/env"
)

// Open file relatively to project
func Open(name string) (*os.File, error) {
	path := fmt.Sprintf(env.ProjectRoot+"%s", name)
	return os.Open(path)
}

func ReadFile(name string) string {
	var output string
	path := fmt.Sprintf(env.ProjectRoot+"%s", name)
	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return output
}
