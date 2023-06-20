package frontbuild

import (
	"log"

	"github.com/ducksouplab/mastok/helpers"
	"gopkg.in/yaml.v2"
)

type config struct {
	Version string
}

var version string

func init() {
	// read version from config file
	f, err := helpers.Open("front/config.yml")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	var c config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&c)
	if err != nil {
		log.Fatalln(err)
	}
	version = c.Version
}
