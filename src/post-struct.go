package main

import "time"

type Event struct {
	Start     string `yaml:"StartDate"`
	End       string `yaml:"EndDate"`
	StartDate time.Time
	EndDate   time.Time
	Status    string `yaml:"Status"`
	Location  string `yaml:"Location"`
}

type Contact struct {
	Name      string `yaml:"name"`
	Honorific string `yaml:"honorific"`
	Email     string `yaml:"email"`
	Photo     string `yaml:"u-photo"`
	URL       string `yaml:"u-url"`
	Key       string `yaml:"u-key"`
	LinkedIn  string `yaml:"linkedin"`
	Logo      string `yaml:"u-logo"`
	Title     string `yaml:"p-job-title"`
}

type Education struct {
	Name      string `yaml:"p-name"`
	Start     string `yaml:"dt-start"`
	End       string `yaml:"dt-end"`
	StartDate time.Time
	EndDate   time.Time
	URL       string `yaml:"u-url"`
	Category  string `yaml:"p-category"`
	Location  string `yaml:"p-location"`
}

type Experience struct {
	Name          string `yaml:"p-name"`
	Summary       string `yaml:"p-summary"`
	Start         string `yaml:"dt-start"`
	StartDate     time.Time
	Description   string `yaml:"p-description"`
	URL           string `yaml:"u-url"`
	Location      string `yaml:"p-location"`
	Category      string `yaml:"p-category"`
	Published     string `yaml:"dt-published"`
	PublishedDate time.Time
	Author        string `yaml:"p-author"`
}

type TimedExperience struct {
	FivePlus  []string `yaml:"5+ years"`
	OneToFive []string `yaml:"1-5 years"`
	New       []string `yaml:"<1 year"`
}

type SkillGroup struct {
	Name    string   `yaml:"name"`
	Members []string `yaml:"members"`
}

type Skill struct {
	SeniorDev      []SkillGroup    `yaml:"seniordev"`
	Developer      []SkillGroup    `yaml:"developer"`
	Intern         []SkillGroup    `yaml:"intern"`
	HobbyPro       []SkillGroup    `yaml:"hobbypro"`
	Hobbiest       []SkillGroup    `yaml:"hobbiest"`
	Dabbler        []SkillGroup    `yaml:"dabbler"`
	Programming    TimedExperience `yaml:"Programming languages"`
	Libraries      TimedExperience `yaml:"Libraries/ services/ technologies"`
	Accreditations []string        `yaml:"Principal methodology accreditations"`
}
type Resume struct {
	Contact     Contact      `yaml:"Contact"`
	Education   []Education  `yaml:"Education"`
	Experience  []Experience `yaml:"Experience"`
	Skill       Skill        `yaml:"Skill"`
	Affiliation []string     `yaml:"Affiliation"`
}

type SyndicationLinksS struct {
	Twitter   string `yaml:"Twitter"`
	Instagram string `yaml:"Instagram"`
	Mastodon  string `yaml:"Mastodon"`
}

type ItemS struct {
	URL    string  `yaml:"url"`
	Image  string  `yaml:"image"`
	Name   string  `yaml:"name"`
	Type   string  `yaml:"type"`
	Rating float32 `yaml:"rating"`
}

type FrontMatter struct {
	ID               string            `yaml:"Id"`
	Title            string            `yaml:"Title"`
	Tags             []string          `yaml:"Tags"`
	Created          string            `yaml:"Created"`
	Updated          string            `yaml:"Updated"`
	Type             string            `yaml:"Type"`
	Status           string            `yaml:"Status"`
	Synopsis         string            `yaml:"Synopsis"`
	Author           string            `yaml:"Author"`
	FeatureImage     string            `yaml:"FeatureImage"`
	AttachedMedia    []string          `yaml:"AttachedMedia"`
	SyndicationLinks SyndicationLinksS `yaml:"Syndication"`
	Slug             string            `yaml:"Slug"`
	Event            Event             `yaml:"Event"`
	Resume           Resume            `yaml:"Resume"`
	Link             string            `yaml:"Link"`
	InReplyTo        string            `yaml:"in-reply-to"`
	BookmarkOf       string            `yaml:"bookmark-of"`
	FavoriteOf       string            `yaml:"favorite-of"`
	RepostOf         string            `yaml:"repost-of"`
	LikeOf           string            `yaml:"like-of"`
	Item             ItemS             `yaml:"Item"`
	RelativeLink     string
	CreatedDate      time.Time
	UpdatedDate      time.Time
}

type BlogPost struct {
	Frontmatter FrontMatter
	Contents    string
}
