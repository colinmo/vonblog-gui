package main

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func frontMatterDefaults(frontMatter *FrontMatter) {
	created, err2 := parseUnknownDateFormat(frontMatter.Created)
	if err2 != nil {
		created = time.Now()
	}
	frontMatter.Created = created.Format("2006-01-02T15:04:05-0700")
	frontMatter.Updated = time.Now().Format("2006-01-02T15:04:05-0700")

	frontMatter.Slug = setEmptyStringDefault(frontMatter.Slug, textToSlug(frontMatter.Title))
	if len(frontMatter.Slug) < 5 || frontMatter.Slug[len(frontMatter.Slug)-5:] != ".html" {
		frontMatter.Slug = frontMatter.Slug + ".html"
	}
	frontMatter.Status = setEmptyStringDefault(frontMatter.Status, "live")

	if len(frontMatter.Tags) == 0 {
		frontMatter.Tags = []string{}
	}

	if len(frontMatter.Type) == 0 {
		frontMatter.Type = "article"
	}
}

func frontMatterValidateExperience(frontMatter *FrontMatter) {
	for i, x := range frontMatter.Resume.Experience {
		if len(x.Description) > 0 {
			var buf2 bytes.Buffer
			md.Convert([]byte(x.Description), &buf2)
			frontMatter.Resume.Experience[i].Description = buf2.String()
		}
		if len(x.Summary) > 0 {
			var buf2 bytes.Buffer
			md.Convert([]byte(x.Summary), &buf2)
			frontMatter.Resume.Experience[i].Summary = strings.Replace(strings.Replace(buf2.String(), "<p>", "", 1), "</p>", "", 1)
		}
		frontMatter.Resume.Experience[i].StartDate, _ = parseUnknownDateFormat(x.Start)
		frontMatter.Resume.Experience[i].PublishedDate, _ = parseUnknownDateFormat(x.Published)
	}
	for i, d := range frontMatter.Resume.Education {
		frontMatter.Resume.Education[i].StartDate, _ = parseUnknownDateFormat(d.Start)
		frontMatter.Resume.Education[i].EndDate, _ = parseUnknownDateFormat(d.End)
	}
}

func frontMatterValidate(frontMatter *FrontMatter) []string {
	var collectedErrors []string
	if len(frontMatter.Resume.Contact.Name) > 0 {
		frontMatterValidateExperience(frontMatter)
	}
	return collectedErrors
}
func parseUnknownTimezone(dateString string) *time.Location {
	// Timezones
	var l *time.Location
	if dateString[len(dateString)-1:] == "Z" {
		l, _ = time.LoadLocation("UTC")
		return l
	}
	re := regexp.MustCompile(`([+-]\d{4}|[+-]\d{2}:\d{2})`)
	matches := re.FindStringSubmatch(dateString)
	hrplus, minplus := 0, 0
	if matches != nil {
		if len(matches[1]) == 5 {
			hrplus, _ = strconv.Atoi(matches[1][0:3])
			minplus, _ = strconv.Atoi(matches[1][3:5])
		} else {
			hrplus, _ = strconv.Atoi(matches[1][0:3])
			minplus, _ = strconv.Atoi(matches[1][4:6])
		}
	} else {
		re = regexp.MustCompile(`([A-Z]{3}[+-]\d\d)`)
		matches = re.FindStringSubmatch(dateString)
		if matches != nil {
			hrplus, _ = strconv.Atoi(matches[1][4:])
		}
	}
	l = time.FixedZone("postzone", hrplus*3600+minplus*60)
	return l
}

func parseUnknownTime(dateString string, re *regexp.Regexp) (int, int, int) {
	hr := 0
	mi := 0
	se := 0
	// Time
	matches := re.FindStringSubmatch(dateString)
	if matches == nil {
		matches = re.FindStringSubmatch(dateString)
		if matches != nil {
			hr, _ = strconv.Atoi(matches[1])
			mi, _ = strconv.Atoi(matches[2])
			se = 0
			if strings.ToLower(matches[3]) == "pm" {
				hr += 12
			}
		}
	} else {
		hr, _ = strconv.Atoi(matches[1])
		mi, _ = strconv.Atoi(matches[2])
		se, _ = strconv.Atoi(matches[3])
	}
	return hr, mi, se
}

func parseUnknownDate(dateString string) (int, int, int, error) {
	yr, mn, dy := 0, 0, 0
	var newTime time.Time
	// Date
	re := regexp.MustCompile(`(\d{1,2})\s*(\w{3})\s*(\d{4})`)
	date := re.FindStringSubmatch(dateString)
	if date == nil {
		re = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
		date = re.FindStringSubmatch(dateString)
		if date == nil {
			re = regexp.MustCompile(`(\w{3})\w*\s+(\d{1,2})[,\s]+(\d{4})`)
			date = re.FindStringSubmatch(dateString)
			if date == nil {
				return yr, mn, dy, errors.New("could not parse date")
			} else {
				dy, _ = strconv.Atoi(date[2])
				yr, _ = strconv.Atoi(date[3])
				newTime, _ = time.Parse("Jan", date[1])
				mn = int(newTime.Month())
			}
		} else {
			dy, _ = strconv.Atoi(date[3])
			yr, _ = strconv.Atoi(date[1])
			mn, _ = strconv.Atoi(date[2])

		}
	} else {
		dy, _ = strconv.Atoi(date[1])
		yr, _ = strconv.Atoi(date[3])
		newTime, _ = time.Parse("Jan", date[2])
		mn = int(newTime.Month())
	}
	return yr, mn, dy, nil
}

func parseUnknownDateFormat(dateString string) (time.Time, error) {
	var newTime time.Time
	var err error
	var hr, mi, se, dy, yr int
	var l *time.Location
	var mn int

	if len(dateString) == 0 {
		return newTime, err
	}
	l = parseUnknownTimezone(dateString)
	re := regexp.MustCompile(`(\d{1,2}):(\d{1,2})[: ]((\d{1,2})|([ap]m))`)
	hr, mi, se = parseUnknownTime(dateString, re)
	dateString = re.ReplaceAllString(dateString, " ")
	yr, mn, dy, err = parseUnknownDate(dateString)
	// Create date in specified timezone
	newTime = time.Date(yr, time.Month(mn), dy, hr, mi, se, 0, l)
	// Convert to blog timezone
	loc, _ := time.LoadLocation(blogTimezone)
	newTime = newTime.In(loc)

	return newTime, err
}

func setEmptyStringDefault(value string, ifempty string) string {
	if len(value) == 0 {
		return ifempty
	}
	return value
}

func textToSlug(intext string) string {
	re := regexp.MustCompile("[^.a-zA-Z0-9-]")
	slug := strings.ToLower(re.ReplaceAllString(intext, "-"))
	re = regexp.MustCompile("-+")
	slug = re.ReplaceAllString(slug, "-")
	re = regexp.MustCompile("-*$")
	slug = re.ReplaceAllString(slug, "")
	re = regexp.MustCompile("^-*")
	slug = re.ReplaceAllString(slug, "")
	return slug
}
