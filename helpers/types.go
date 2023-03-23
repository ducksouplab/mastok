package helpers

import (
	"bufio"
	"log"
	"strconv"
	"strings"
)

type Group struct {
	Label string
	Size  uint
}

type Grouping struct {
	Question string
	Groups   []Group
}

func ParseGroup(text string) (grouping Grouping, ok bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	scanner := bufio.NewScanner(strings.NewReader(text))
	var question string
	var groups []Group
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
		if len(question) == 0 {
			question = line
		} else {
			splits := strings.Split(line, ":")
			log.Println(splits)
			size, err := strconv.ParseUint(splits[1], 10, 64)
			log.Println(size, err)
			if err != nil {
				return
			}
			groups = append(groups, Group{splits[0], uint(size)})
		}
	}
	return Grouping{question, groups}, true
}
