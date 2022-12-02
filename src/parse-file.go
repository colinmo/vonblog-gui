package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

// parseString parses the passed string and returns the html conversion and yaml frontmatter
func parseString(body string) (string, FrontMatter, error) {
	var html2 string
	var err error
	var frontMatter FrontMatter

	// Parse the frontmatter at the start of the file
	split := strings.SplitN(body[3:], "---", 2)
	if len(split) != 2 {
		return html2, frontMatter, err
	}
	frontMatter, err = parseFrontMatter(split[0])
	if err != nil {
		return html2, frontMatter, err
	}

	html2 = strings.Join(split[1:], "---")

	return html2, frontMatter, err
}

func parseFrontMatter(inFrontMatter string) (FrontMatter, error) {
	var frontMatter FrontMatter
	err := yaml.Unmarshal([]byte(inFrontMatter), &frontMatter)
	if err != nil {
		fmt.Printf("Failed to parse frontmatter %s\n", inFrontMatter)
		log.Fatal(err)
		return frontMatter, err
	}
	frontMatterDefaults(&frontMatter)
	collectedErrors := frontMatterValidate(&frontMatter)

	if len(collectedErrors) > 0 {
		err = errors.New(strings.Join(collectedErrors, ", "))
	}
	return frontMatter, err
}
