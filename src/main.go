package main

import (
	"bytes"
	"os"
	"time"
	icon "vonbloggui/icon"

	fyne "fyne.io/fyne/v2"
	app "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pkg/browser"
	"github.com/yuin/goldmark"
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

func setup() {
	os.Setenv("TZ", "Australia/Brisbane")
	AppStatus = AppStatusStruct{
		TaskCount: 0,
	}
}
func main() {
	setup()
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
	thisPost = BlogPost{
		Frontmatter: FrontMatter{
			Title: binding.NewString(),
			Tags:  binding.NewStringList(),
		},
		Contents: binding.NewString(),
	}
	menu := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentPrintIcon(), func() {
			parsedOut := markdownInput.Text
			md := goldmark.New(
				goldmark.WithExtensions(extension.GFM),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
				goldmark.WithRendererOptions(
					html.WithHardWraps(),
					html.WithXHTML(),
				),
			)
			var buf bytes.Buffer
			if err := md.Convert([]byte(parsedOut), &buf); err != nil {
				panic(err)
			}
			tmpFile, _ := os.CreateTemp(os.TempDir(), "markdownpreview-*.html")
			defer os.Remove(tmpFile.Name())
			tmpFile.Write([]byte(markdownHTMLHeader))
			tmpFile.Write(buf.Bytes())
			tmpFile.Write([]byte(markdownHTMLFooter))
			tmpFile.Close()
			browser.OpenFile(tmpFile.Name())
			time.Sleep(time.Second * 2)
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			// submit to bitbucket
		}),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			// Popup window for title etc.
			oldTitle, _ := thisPost.Frontmatter.Title.Get()
			dialog.ShowForm(
				"Frontmatter",
				"OK",
				"Cancel",
				[]*widget.FormItem{
					{Text: "Title", Widget: widget.NewEntryWithData(thisPost.Frontmatter.Title)},
					{Text: "Tags", Widget: widget.NewListWithData(thisPost.Frontmatter.Tags, func() fyne.CanvasObject { return fyne.NewContainer() }, func(y binding.DataItem, c fyne.CanvasObject) {})},
				},
				func(x bool) {
					if x {

					} else {
						thisPost.Frontmatter.Title.Set(oldTitle)
					}
				},
				mainWindow,
			)
		}),
	)
	content := container.NewBorder(
		container.NewHBox(menu),
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
