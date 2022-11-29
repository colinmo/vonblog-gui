package main

import (
	"time"

	"fyne.io/fyne/v2/data/binding"
)

type Event struct {
	Start    binding.String `yaml:"StartDate"`
	End      binding.String `yaml:"EndDate"`
	Status   binding.String `yaml:"Status"`
	Location binding.String `yaml:"Location"`
}

type Contact struct {
	Name      binding.String `yaml:"name"`
	Honorific binding.String `yaml:"honorific"`
	Email     binding.String `yaml:"email"`
	Photo     binding.String `yaml:"u-photo"`
	URL       binding.String `yaml:"u-url"`
	Key       binding.String `yaml:"u-key"`
	LinkedIn  binding.String `yaml:"linkedin"`
	Logo      binding.String `yaml:"u-logo"`
	Title     binding.String `yaml:"p-job-title"`
}

type Education struct {
	Name     binding.String `yaml:"p-name"`
	Start    binding.String `yaml:"dt-start"`
	End      binding.String `yaml:"dt-end"`
	URL      binding.String `yaml:"u-url"`
	Category binding.String `yaml:"p-category"`
	Location binding.String `yaml:"p-location"`
}

type Experience struct {
	Name        binding.String `yaml:"p-name"`
	Summary     binding.String `yaml:"p-summary"`
	Start       binding.String `yaml:"dt-start"`
	Description binding.String `yaml:"p-description"`
	URL         binding.String `yaml:"u-url"`
	Location    binding.String `yaml:"p-location"`
	Category    binding.String `yaml:"p-category"`
	Published   binding.String `yaml:"dt-published"`
	Author      binding.String `yaml:"p-author"`
}

type TimedExperience struct {
	FivePlus  binding.StringList `yaml:"5+ years"`
	OneToFive binding.StringList `yaml:"1-5 years"`
	New       binding.StringList `yaml:"<1 year"`
}

type SkillGroup struct {
	Name    binding.String     `yaml:"name"`
	Members binding.StringList `yaml:"members"`
}

type Skill struct {
	SeniorDev      []SkillGroup       `yaml:"seniordev"`
	Developer      []SkillGroup       `yaml:"developer"`
	Intern         []SkillGroup       `yaml:"intern"`
	HobbyPro       []SkillGroup       `yaml:"hobbypro"`
	Hobbiest       []SkillGroup       `yaml:"hobbiest"`
	Dabbler        []SkillGroup       `yaml:"dabbler"`
	Programming    TimedExperience    `yaml:"Programming languages"`
	Libraries      TimedExperience    `yaml:"Libraries/ services/ technologies"`
	Accreditations binding.StringList `yaml:"Principal methodology accreditations"`
}
type Resume struct {
	Contact     Contact            `yaml:"Contact"`
	Education   []Education        `yaml:"Education"`
	Experience  []Experience       `yaml:"Experience"`
	Skill       Skill              `yaml:"Skill"`
	Affiliation binding.StringList `yaml:"Affiliation"`
}

type SyndicationLinksS struct {
	Twitter   binding.String `yaml:"Twitter"`
	Instagram binding.String `yaml:"Instagram"`
	Mastodon  binding.String `yaml:"Mastodon"`
}

type ItemS struct {
	URL    binding.String
	Image  binding.String
	Name   binding.String
	Type   binding.String
	Rating float32 `yaml:"rating"`
}

type FrontMatter struct {
	ID               string
	Title            binding.String
	Tags             binding.StringList
	Created          binding.String
	Updated          binding.String
	Type             binding.String
	Status           binding.String
	Synopsis         binding.String
	Author           binding.String
	FeatureImage     binding.String
	AttachedMedia    binding.StringList
	SyndicationLinks SyndicationLinksS
	Slug             binding.String
	Event            Event
	Resume           Resume
	Link             binding.String
	InReplyTo        binding.String
	BookmarkOf       binding.String
	FavoriteOf       binding.String
	RepostOf         binding.String
	LikeOf           binding.String
	Item             ItemS
	RelativeLink     binding.String
	CreatedDate      time.Time
	UpdatedDate      time.Time
}

type BlogPost struct {
	Frontmatter FrontMatter
	Contents    binding.String
}
