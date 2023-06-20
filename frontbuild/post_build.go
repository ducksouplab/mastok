package frontbuild

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var jsLineRegex, jsUpdateRegex, cssLineRegex, cssUpdateRegex *regexp.Regexp

func init() {
	jsLineRegex = regexp.MustCompile("WebPrefix.*/assets/(.*)/js.*js")
	jsUpdateRegex = regexp.MustCompile("assets/.*?/js")
	cssLineRegex = regexp.MustCompile("WebPrefix.*/assets/(.*)/css.*css")
	cssUpdateRegex = regexp.MustCompile("assets/.*?/css")
}

func cleanUpAssets() {
	path := "front/static/assets/"
	infos, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, info := range infos {
		log.Println(info.Name())
		if info.IsDir() && info.Name() != version {
			os.RemoveAll(path + info.Name())
		}
	}
}

func updateTemplates() {
	filepath.Walk("templates/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() {
			replaceIncludes(path)
		}
		return nil
	})
}

func replaceIncludes(path string) {
	input, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	var js, css bool
	for i, line := range lines {
		if jsLineRegex.MatchString(line) {
			lines[i] = jsUpdateRegex.ReplaceAllString(line, "assets/"+version+"/js")
			js = true
		} else if cssLineRegex.MatchString(line) {
			lines[i] = cssUpdateRegex.ReplaceAllString(line, "assets/"+version+"/css")
			css = true
		}
	}
	// log once
	if js {
		log.Printf("[Template] %v CSS prefixed with version %v\n", path, version)
	}
	if css {
		log.Printf("[Template] %v  JS prefixed with version %v\n", path, version)
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(path, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
