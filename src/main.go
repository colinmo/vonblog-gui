package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	icon "vonbloggui/icon"

	fyne "fyne.io/fyne/v2"
	app "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	html2 "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/pkg/browser"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type AppStatusStruct struct {
	TaskCount int
}

var thisApp fyne.App
var mainWindow fyne.Window
var preferencesWindow fyne.Window
var AppStatus AppStatusStruct
var markdownInput *widget.Entry
var thisPost BlogPost
var dateFormatString = "2006-01-02 15:04:05"
var blogTimezone = "Australia/Brisbane"
var md goldmark.Markdown
var formEntries = map[string]*widget.Entry{}
var formSelect = map[string]*widget.Select{}
var recentUploads []struct {
	Path  string
	Thumb string
}

func setup() {
	os.Setenv("TZ", blogTimezone)
	AppStatus = AppStatusStruct{
		TaskCount: 0,
	}
	// Default Markdown parser
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			meta.New(meta.WithStoresInDocument()),
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					html2.WithLineNumbers(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	Client = &http.Client{}
	startLocalServers()
}
func main() {
	setup()
	bitbucket.Login()
	thisApp = app.NewWithID("com.vonexplaino.vonblog")
	thisApp.SetIcon(fyne.NewStaticResource("Systray", icon.Data))
	preferencesWindow = thisApp.NewWindow("Preferences")
	preferencesWindowSetup()
	mainWindow = thisApp.NewWindow("Post")
	mainWindowSetup()
	if desk, ok := thisApp.(desktop.App); ok {
		m := fyne.NewMenu("VonBlog",
			fyne.NewMenuItem("Post", func() {
				mainWindow.Show()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Preferences", func() {
				preferencesWindowSetup()
				preferencesWindow.Show()
			}),
		)
		desk.SetSystemTrayMenu(m)
	}
	thisApp.Run()
}

func preferencesWindowSetup() {
	// Bitbucket URL and Keys
}

func mainWindowSetup() {
	mainWindow.Resize(fyne.NewSize(800, 800))
	mainWindow.SetMaster()
	mainWindow.Hide()
	markdownInput = widget.NewEntry()
	markdownInput.MultiLine = true
	markdownInput.Wrapping = fyne.TextWrapWord
	thisPost = BlogPost{}
	formEntries = map[string]*widget.Entry{
		"Title":        MakeEntryWithText(thisPost.Frontmatter.Title),
		"Tags":         MakeEntryWithText(strings.Join(thisPost.Frontmatter.Tags, ",")),
		"Created":      MakeEntryWithText(time.Now().Format(dateFormatString)),
		"Updated":      MakeEntryWithText(""),
		"Synopsis":     MakeEntryWithText(thisPost.Frontmatter.Synopsis),
		"FeatureImage": MakeEntryWithText(thisPost.Frontmatter.FeatureImage),

		"InReplyTo":  MakeEntryWithText(thisPost.Frontmatter.InReplyTo),
		"BookmarkOf": MakeEntryWithText(thisPost.Frontmatter.BookmarkOf),
		"FavoriteOf": MakeEntryWithText(thisPost.Frontmatter.FavoriteOf),
		"RepostOf":   MakeEntryWithText(thisPost.Frontmatter.RepostOf),
		"LikeOf":     MakeEntryWithText(thisPost.Frontmatter.LikeOf),
	}
	/*
		formMedia := []struct {
			URL  string
			File image.NRGBA
		}{}
	*/
	/*
		AttachedMedia    []string
		SyndicationLinks SyndicationLinksS
		Event            Event
		Resume           Resume
		Item             ItemS
	*/
	formSelect = map[string]*widget.Select{
		"Type":   MakeSelectWithOptions([]string{"article", "reply", "indieweb", "tweet", "resume", "event", "page", "review"}, thisPost.Frontmatter.Type),
		"Status": MakeSelectWithOptions([]string{"draft", "live", "retired"}, thisPost.Frontmatter.Status),
	}
	menu := widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			// @todo: Prompt to save first.
			ShowBitbucketNavigator()
		}),
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			// Validate/ parse fields as required
			frontMatterDefaults(&thisPost.Frontmatter)
			errors := frontMatterValidate(&thisPost.Frontmatter)
			UpdateAllFields(formEntries, formSelect)
			if len(errors) > 0 {
				fmt.Printf("Failed: %v\n", errors)
			} else {
				fmt.Printf("Continue upload")
				// Get the media together in a media submission
				// Convert the fields into the Markdown post
				// Submit to bitbucket
				// Handle response
			}
		}),
		widget.NewToolbarAction(theme.DocumentPrintIcon(), func() {
			parsedOut := markdownInput.Text
			tmpFile, _ := os.CreateTemp(os.TempDir(), "markdownpreview-*.html")
			defer os.Remove(tmpFile.Name())
			tmpFile.Write([]byte(markdownHTMLHeader))
			tmpFile.Write([]byte(markdownToHtml(parsedOut)))
			tmpFile.Write([]byte(markdownHTMLFooter))
			tmpFile.Close()
			browser.OpenFile(tmpFile.Name())
			time.Sleep(time.Second * 2)
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			// Popup window for title etc.
			// @todo: Pull entries from a Bitbucket file if editing
			dialog.ShowForm(
				"Frontmatter",
				"OK",
				"Cancel",
				[]*widget.FormItem{
					{Text: "Created", Widget: formEntries["Created"]},
					{Text: "Updated", Widget: formEntries["Updated"]},
					/*{Text: "FeatureImage", Widget: formEntries["FeatureImage"]},*/
					{Text: "", Widget: widget.NewLabel("Indieweb")},
					{Text: "InReplyTo", Widget: formEntries["InReplyTo"]},
					{Text: "BookmarkOf", Widget: formEntries["BookmarkOf"]},
					{Text: "FavoriteOf", Widget: formEntries["FavoriteOf"]},
					{Text: "RepostOf", Widget: formEntries["RepostOf"]},
					{Text: "LikeOf", Widget: formEntries["LikeOf"]},
					{Text: "Extended", Widget: container.NewVBox(widget.NewButton("Event", func() {}), widget.NewButton("Resume", func() {}), widget.NewButton("Review", func() {}))},
				},
				func(x bool) {
					if x {
						thisPost.Frontmatter.Title = formEntries["Title"].Text
						thisPost.Frontmatter.Tags = strings.Split(formEntries["Tags"].Text, ",")
					}
				},
				mainWindow,
			)
		}),
		widget.NewToolbarAction(theme.UploadIcon(), func() {
			// @todo when uploading images, remember the name and location
			// so when clicking the gallery button, they can be suggested
			LocalFileSelectorWindow()
			// File selector prompt
			// If an upload is an image, create the thumbnail
			// Upload
			// Handle response
			// Remember uploads
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.MediaPhotoIcon(), func() {}),
		widget.NewToolbarAction(theme.NavigateNextIcon(), func() {}),
	)
	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(menu),
			container.New(
				layout.NewFormLayout(),
				widget.NewLabel("Title"),
				formEntries["Title"],
			),
			container.NewGridWithColumns(
				2,
				container.NewGridWithColumns(
					2,
					container.New(
						layout.NewFormLayout(),
						widget.NewLabel("Type"),
						formSelect["Type"],
					),
					container.New(
						layout.NewFormLayout(),
						widget.NewLabel("Status"),
						formSelect["Status"],
					),
				),
				container.New(
					layout.NewFormLayout(),
					widget.NewLabel("Tags"),
					formEntries["Tags"],
				),
			),
			container.New(
				layout.NewFormLayout(),
				widget.NewLabel("Synopsis"),
				formEntries["Synopsis"],
			),
		),
		nil,
		nil,
		nil,
		container.NewMax(markdownInput),
	)
	mainWindow.SetContent(content)
	mainWindow.SetCloseIntercept(func() {
		mainWindow.Hide()
	})
}

func MakeEntryWithText(settext string) *widget.Entry {
	b := widget.NewEntry()
	b.SetText(settext)
	return b
}

func MakeSelectWithOptions(options []string, value string) *widget.Select {
	b := widget.NewSelect(options, func(cng string) {})
	b.SetSelected(value)
	return b
}

func MakeCheckGroupWithOptions(options []string, values []string) *widget.CheckGroup {
	b := widget.NewCheckGroup(options, func(cng []string) {})
	b.SetSelected(values)
	return b
}

func UpdateAllFields(formEntries map[string]*widget.Entry, formSelect map[string]*widget.Select) {
	formEntries["Title"].Text = thisPost.Frontmatter.Title
	formEntries["Tags"].Text = strings.Join(thisPost.Frontmatter.Tags, ",")
	formEntries["Created"].Text = thisPost.Frontmatter.Created
	formEntries["Updated"].Text = thisPost.Frontmatter.Updated
	formEntries["Synopsis"].Text = thisPost.Frontmatter.Synopsis
	formEntries["FeatureImage"].Text = thisPost.Frontmatter.FeatureImage
	formEntries["InReplyTo"].Text = thisPost.Frontmatter.InReplyTo
	formEntries["BookmarkOf"].Text = thisPost.Frontmatter.BookmarkOf
	formEntries["FavoriteOf"].Text = thisPost.Frontmatter.FavoriteOf
	formEntries["RepostOf"].Text = thisPost.Frontmatter.RepostOf
	formEntries["LikeOf"].Text = thisPost.Frontmatter.LikeOf
	formEntries["Title"].Refresh()
	formEntries["Tags"].Refresh()
	formEntries["Created"].Refresh()
	formEntries["Updated"].Refresh()
	formEntries["Synopsis"].Refresh()
	formEntries["FeatureImage"].Refresh()
	formEntries["InReplyTo"].Refresh()
	formEntries["BookmarkOf"].Refresh()
	formEntries["FavoriteOf"].Refresh()
	formEntries["RepostOf"].Refresh()
	formEntries["LikeOf"].Refresh()
	/*
		formMedia := []struct {
			URL  string
			File image.NRGBA
		}{}
	*/
	/*
		AttachedMedia    []string
		SyndicationLinks SyndicationLinksS
		Event            Event
		Resume           Resume
		Item             ItemS
	*/
	formSelect["Type"].Selected = thisPost.Frontmatter.Type
	formSelect["Status"].Selected = thisPost.Frontmatter.Status
	formSelect["Type"].Refresh()
	formSelect["Status"].Refresh()
}

func ShowBitbucketNavigator() {
	// Pull down browsable directory list
	// Provide navigations through list
	FileFinderWindow("/")
}

func FileFinderWindow(thispath string) {
	var fileFinder dialog.Dialog
	files, err := bitbucket.GetFiles(thispath)
	if err == nil {
		fileFinderContent := []fyne.CanvasObject{}
		// Sort files
		keys := make([]string, 0, len(files))
		for k := range files {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, path := range keys { // second param is hash
			lPath := path
			labelForPath := path
			if len(path) > len(thispath) && len(thispath) > 2 {
				labelForPath = path[len(thispath)+1:]
			}
			fileFinderContent = append(fileFinderContent, widget.NewButton(labelForPath, func() {
				// load file
				fileFinder.Hide()
				thepath := lPath
				if lPath == ".." {
					thepath = filepath.Dir(thispath)
					if len(thepath) < 2 {
						thepath = "/"
					}
				}
				// If it's a file, load the file. Otherwise, new file finder dialog
				FileFinderWindow(thepath)
			}))
		}
		fileFinder = dialog.NewCustom(
			fmt.Sprintf("Path: %s", thispath),
			"Nevermind",
			container.NewGridWithColumns(3, fileFinderContent...),
			mainWindow,
		)
		fileFinder.Show()
	} else {
		// When loading a file to edit, you have to store the sourceCommitId to save later
		contents, _ := bitbucket.GetFileContents(thispath)
		x, y, _ := parseString(contents)
		thisPost.Contents = x
		markdownInput.Text = x
		markdownInput.Refresh()
		thisPost.Frontmatter = y
		UpdateAllFields(formEntries, formSelect)
	}
}

func LoadDirectory(path string) {

}

func LoadFile(path string) {

}

func LocalFileSelectorWindow() {
	dialog.ShowFolderOpen(
		func(directory fyne.ListableURI, err error) {
			fmt.Printf("Processing %v\n", directory)
			// Get all files
			files, _ := os.ReadDir(directory.Path())
			checkGroup := binding.NewStringList()
			for _, file := range files {
				if !file.IsDir() {
					checkGroup.Append(file.Name())
				}
			}
			fileFinder := dialog.NewCustomConfirm(
				"Upload",
				"Upload",
				"Nevermind",
				widget.NewListWithData(
					checkGroup,
					func() fyne.CanvasObject {
						return widget.NewLabel("template")
					},
					func(i binding.DataItem, o fyne.CanvasObject) {
						o.(*widget.Label).Bind(i.(binding.String))
					},
				),
				func(ok bool) {},
				mainWindow,
			)
			fileFinder.Resize(fyne.NewSize(300, 300))
			fileFinder.Show()
			// Show them with checkboxes
			// Process all checkboxeds
		},
		mainWindow,
	)
}
