package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// parseString parses the passed string and returns the html conversion and yaml frontmatter
func parseString(body string) (string, FrontMatter, error) {
	var markdown string
	var err error
	var frontMatter FrontMatter

	// Parse the frontmatter at the start of the file
	split := strings.SplitN(body[3:], "---", 2)
	if len(split) != 2 {
		return markdown, frontMatter, err
	}
	frontMatter, err = parseFrontMatter(split[0])
	if err != nil {
		return markdown, frontMatter, err
	}
	markdown = strings.Join(split[1:], "---")
	return markdown, frontMatter, err
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

var gallery_index = 0

func markdownToHtml(markdown string) string {
	// Convert the Gallery tags
	var buf2 bytes.Buffer
	re := regexp.MustCompile(`<section [^>]*gallery[^>]* markdown="1"[^>]*>(?sm)(.*?)</section>`)
	bodybyte := re.ReplaceAllFunc(
		[]byte(markdown),
		convertGallery,
	)
	// Convert the markdown=1 tags
	re = regexp.MustCompile(`<[^>]* markdown="1"[^>]*>(?sm)(.*?)</[^>]*>`)
	bodybyte = re.ReplaceAllFunc(
		bodybyte,
		convertMarkdownHtml,
	)
	// Now convert what's left to Markdown
	md.Convert(bodybyte, &buf2)
	return buf2.String()

}

func convertGallery(mep []byte) []byte {
	re := regexp.MustCompile(`(<section [^>]*gallery[^>]*) markdown="1"([^>]*>)(?sm)(.*?)(</[^>]*>)`)
	mep1 := re.FindAllStringSubmatch(string(mep), -1)
	// Convert the image collection from Markdown to HTML
	var buf2 bytes.Buffer
	md.Convert([]byte(mep1[0][3]), &buf2)
	// Convert the individual images into the converted Gallery version
	re2 := regexp.MustCompile(`<a href="([^"]*)"><img src="([^"]*)" alt="([^"]*)" title="([^"]*)"></a>`)
	mep2 := re2.FindAllStringSubmatch(buf2.String(), -1)
	stringOut := ""
	for i := 0; i < len(mep2); i++ {
		stringOut += fmt.Sprintf(`<input type="radio" name="gallery-2020-4-%d" id="gallery-2020-4-%d-%d" tabindex="-1" />
				<label for="gallery-2020-4-%d-%d">
					<img src="%s" />
				</label>
				<figure>
					<img loading="lazy" src="%s" alt="%s" />
					<figcaption>
						<em>%s</em>
					</figcaption>
				</figure>
		`,
			gallery_index,
			gallery_index,
			i+1,
			gallery_index,
			i,
			mep2[i][2],
			mep2[i][1],
			mep2[i][3],
			mep2[i][4],
		)
	}
	// Print out the converted HTML, removing the "markdown="1"""
	temp := []byte(fmt.Sprintf(
		`%s%s<input type="radio" name="gallery-2020-4-%d" id="gallery-2020-4-%d-0" />
				<label></label><figure></figure>%s<input type="radio" name="gallery-2020-4-%d" id="gallery-2020-4-%d-close" />
				<label for="gallery-2020-4-%d-close">X</label>%s`,
		mep1[0][1],
		mep1[0][2],
		gallery_index,
		gallery_index,
		stringOut,
		gallery_index,
		gallery_index,
		gallery_index,
		mep1[0][4],
	))
	gallery_index++
	return temp
}

func convertMarkdownHtml(mep []byte) []byte {
	re := regexp.MustCompile(`(<[^>]*) markdown="1"([^>]*>)(?sm)((.|\r|\n)*?)(</[^>]*>)`)
	mep1 := re.FindAllStringSubmatch(string(mep), -1)
	var buf2 bytes.Buffer
	md.Convert([]byte(mep1[0][3]), &buf2)
	html := buf2.String()
	if mep1[0][3][0:1] != "\n" {
		html = html[3 : len(html)-5]
	}

	return []byte(fmt.Sprintf(
		`%s%s%s%s`,
		mep1[0][1],
		mep1[0][2],
		html,
		mep1[0][5],
	))
}
