//go:generate fyne bundle -o bundled.go icon/quotes.png
//go:generate fyne bundle -o bundled.go -append icon/picture.png
//go:generate fyne bundle -o bundled.go -append icon/0star.png
//go:generate fyne bundle -o bundled.go -append icon/0.5star.png
//go:generate fyne bundle -o bundled.go -append icon/1star.png
//go:generate fyne bundle -o bundled.go -append icon/1.5star.png
//go:generate fyne bundle -o bundled.go -append icon/2star.png
//go:generate fyne bundle -o bundled.go -append icon/2.5star.png
//go:generate fyne bundle -o bundled.go -append icon/3star.png
//go:generate fyne bundle -o bundled.go -append icon/3.5star.png
//go:generate fyne bundle -o bundled.go -append icon/4star.png
//go:generate fyne bundle -o bundled.go -append icon/4.5star.png
//go:generate fyne bundle -o bundled.go -append icon/5star.png

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	icon "vonbloggui/icon"

	fyne "fyne.io/fyne/v2"
	app "fyne.io/fyne/v2/app"
	canvas "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	storage "fyne.io/fyne/v2/storage"
	repository "fyne.io/fyne/v2/storage/repository"
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

type GalleryStruct struct {
	Alt     binding.String
	Caption binding.String
	Image   binding.String
}

var thisApp fyne.App
var mainWindow fyne.Window
var preferencesWindow fyne.Window
var galleryWindow fyne.Window
var AppStatus AppStatusStruct
var markdownInput *widget.Entry
var thisPost BlogPost
var dateFormatString = "2006-01-02T15:04:05-0700"
var blogTimezone = "Australia/Brisbane"
var md goldmark.Markdown

var formEntries = map[string]*widget.Entry{}
var formSelect = map[string]*widget.Select{}
var formCheckbox = map[string]*widget.Check{}
var formSlider = map[string]*widget.Slider{}

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
	thisApp = app.NewWithID("com.vonexplaino.vonblog")
	thisApp.SetIcon(fyne.NewStaticResource("Systray", icon.Data))
	hideMeIfNotWindows()
	preferencesWindow = thisApp.NewWindow("Preferences")
	preferencesWindowSetup()
	clientkey := binding.BindPreferenceString("clientkey", thisApp.Preferences())
	clientsecret := binding.BindPreferenceString("clientsecret", thisApp.Preferences())
	ck, _ := clientkey.Get()
	cs, _ := clientsecret.Get()
	if len(ck) == 0 || len(cs) == 0 {
		preferencesWindow.Show()
	} else if len(thisApp.Preferences().String("accesstoken")) == 0 ||
		len(thisApp.Preferences().String("refreshtoken")) == 0 {
		bitbucket.Login()
	} else {
		bitbucket.RefreshIfRequired()
	}
	mainWindow = thisApp.NewWindow("Post")
	mainWindowSetup()

	galleryWindow = thisApp.NewWindow("Gallery")
	galleryWindow.Resize(fyne.NewSize(800, 800))
	galleryWindow.Hide()
}

func main() {
	setup()
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
	preferencesWindow.Resize(fyne.NewSize(500, 500))
	preferencesWindow.Hide()
	baseurl := binding.BindPreferenceString("baseurl", thisApp.Preferences())
	workspacekey := binding.BindPreferenceString("workspacekey", thisApp.Preferences())
	reposslug := binding.BindPreferenceString("reposslug", thisApp.Preferences())
	clientkey := binding.BindPreferenceString("clientkey", thisApp.Preferences())
	clientsecret := binding.BindPreferenceString("clientsecret", thisApp.Preferences())
	accesstoken := binding.BindPreferenceString("accesstoken", thisApp.Preferences())
	refreshtoken := binding.BindPreferenceString("refreshtoken", thisApp.Preferences())
	expiration := binding.BindPreferenceString("expiration", thisApp.Preferences())
	startingfolder := binding.BindPreferenceString("startingfolder", thisApp.Preferences())

	oldclientkey, _ := clientkey.Get()
	oldclientsecret, _ := clientsecret.Get()

	preferencesWindow.SetContent(
		container.New(
			layout.NewFormLayout(),
			widget.NewLabel("Base URL"),
			widget.NewEntryWithData(baseurl),
			widget.NewLabel("Workspace Key"),
			widget.NewEntryWithData(workspacekey),
			widget.NewLabel("Repository Slug"),
			widget.NewEntryWithData(reposslug),
			widget.NewLabel("Client Key"),
			widget.NewEntryWithData(clientkey),
			widget.NewLabel("Client Secret"),
			widget.NewEntryWithData(clientsecret),
			widget.NewLabel("AccessToken"),
			widget.NewEntryWithData(accesstoken),
			widget.NewLabel("Refresh Token"),
			widget.NewEntryWithData(refreshtoken),
			widget.NewLabel("Expiration ("+dateFormatString+")"),
			widget.NewEntryWithData(expiration),
			widget.NewLabel("Starting folder"),
			widget.NewEntryWithData(startingfolder),
		),
	)
	preferencesWindow.SetCloseIntercept(func() {
		preferencesWindow.Hide()
		newclientkey, _ := clientkey.Get()
		newclientsecret, _ := clientsecret.Get()
		if oldclientkey != newclientkey || oldclientsecret != newclientsecret {
			bitbucket.Login()
		}
	})
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
		"Created":      MakeEntryWithText(""),
		"Updated":      MakeEntryWithText(""),
		"Synopsis":     MakeEntryWithText(thisPost.Frontmatter.Synopsis),
		"FeatureImage": MakeEntryWithText(thisPost.Frontmatter.FeatureImage),

		"Mastodon": MakeEntryWithText(thisPost.Frontmatter.SyndicationLinks.Mastodon),

		"InReplyTo":  MakeEntryWithText(thisPost.Frontmatter.InReplyTo),
		"BookmarkOf": MakeEntryWithText(thisPost.Frontmatter.BookmarkOf),
		"FavoriteOf": MakeEntryWithText(thisPost.Frontmatter.FavoriteOf),
		"RepostOf":   MakeEntryWithText(thisPost.Frontmatter.RepostOf),
		"LikeOf":     MakeEntryWithText(thisPost.Frontmatter.LikeOf),

		"Item.Name":  MakeEntryWithText(thisPost.Frontmatter.Item.Name),
		"Item.URL":   MakeEntryWithText(thisPost.Frontmatter.Item.URL),
		"Item.Image": MakeEntryWithText(thisPost.Frontmatter.Item.Image),

		"Event.Start":    MakeEntryWithText(thisPost.Frontmatter.Event.Start),
		"Event.End":      MakeEntryWithText(thisPost.Frontmatter.Event.End),
		"Event.Location": MakeEntryWithText(thisPost.Frontmatter.Event.Location),
	}
	formEntries["Created"].SetPlaceHolder("YYYY-MM-DDTHH:MI:SS+1000")
	formEntries["Updated"].SetPlaceHolder("YYYY-MM-DDTHH:MI:SS+1000")
	formSlider = map[string]*widget.Slider{
		"Item.Rating": MakeSliderWithValue(thisPost.Frontmatter.Item.Rating),
	}
	formSelect = map[string]*widget.Select{
		"Type":         MakeSelectWithOptions([]string{"article", "reply", "indieweb", "tweet", "resume", "event", "page", "review"}, thisPost.Frontmatter.Type),
		"Status":       MakeSelectWithOptions([]string{"draft", "live", "retired"}, thisPost.Frontmatter.Status),
		"Event.Status": MakeSelectWithOptions([]string{"proposed", "open", "cancelled", "done"}, thisPost.Frontmatter.Event.Status),
	}
	formCheckbox = map[string]*widget.Check{
		"Mastodon": widget.NewCheck("M", func(b bool) {}),
	}
	var content *fyne.Container
	menu := widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			// @todo: Prompt to save first.
			ShowBitbucketNavigator()
		}),
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			confirm := dialog.NewConfirm(
				"Are you sure?",
				"Delete the current entry and start over?",
				func(ok bool) {
					if ok {
						thisPost = BlogPost{
							Frontmatter: FrontMatter{Created: time.Now().Format(dateFormatString)},
						}
						UpdateAllFields(formEntries, formSelect)
						markdownInput.Text = ""
						markdownInput.Refresh()
						thisPost.Filename = ""
					}
				},
				mainWindow,
			)
			confirm.Show()
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			// Validate/ parse fields as required
			FieldsToPost(formEntries, formSelect)
			frontMatterDefaults(&thisPost.Frontmatter)
			errors := frontMatterValidate(&thisPost.Frontmatter)
			thisPost.Contents = markdownInput.Text
			if len(errors) > 0 {
				fmt.Printf("Failed: %v\n", errors)
			} else {
				err := bitbucket.UploadPost()
				if err != nil {
					fmt.Printf("FAILED: %s\n", err)
					widget.ShowModalPopUp(
						widget.NewLabel(fmt.Sprintf("ERROR: %s", err)),
						mainWindow.Canvas(),
					)
				}
				// Handle response
				formEntries["Mastodon"].SetText(thisPost.Frontmatter.SyndicationLinks.Mastodon)
			}
		}),
		widget.NewToolbarAction(
			resourcePreviewSvg,
			func() {
				targetFolder := filepath.Join(os.TempDir(), "vonblog")
				_, err := os.Stat(targetFolder)
				if !os.IsNotExist(err) {
					os.RemoveAll(targetFolder)
				}
				os.Mkdir(targetFolder, 0770)
				tmpFile, _ := os.CreateTemp(targetFolder, "markdownpreview-*.html")
				// @todo: copy any uploaded images
				tmpFile.Write([]byte(markdownHTMLHeader))
				tmpFile.Write([]byte(markdownToHtml(markdownInput.Text)))
				tmpFile.Write([]byte(markdownHTMLFooter))
				tmpFile.Close()
				browser.OpenFile(tmpFile.Name())
				time.Sleep(time.Second * 2)
			}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			// Popup window for title etc.
			// @todo: Pull entries from a Bitbucket file if editing
			dialog.ShowCustomConfirm(
				"Frontmatter",
				"OK",
				"Cancel",
				container.NewVBox(
					func() *widget.Accordion {
						valueLabel := widget.NewIcon(sliderToResource(formSlider["Item.Rating"].Value))
						formSlider["Item.Rating"].OnChanged = func(f float64) {
							valueLabel.SetResource(sliderToResource(formSlider["Item.Rating"].Value))
						}
						rating := widget.NewFormItem(
							"Rating",
							container.NewBorder(
								nil,
								nil,
								valueLabel,
								nil,
								formSlider["Item.Rating"],
							),
						)
						dude := widget.NewAccordion(
							widget.NewAccordionItem(
								"Basics",
								widget.NewForm(
									[]*widget.FormItem{
										{Text: "Created", Widget: formEntries["Created"]},
										{Text: "Updated", Widget: formEntries["Updated"]},
										{Text: "", Widget: widget.NewLabel("Syndication [XPOST to make]")},
										{Text: "Mastodon", Widget: formEntries["Mastodon"]},
										/*{Text: "FeatureImage", Widget: formEntries["FeatureImage"]},*/
									}...,
								),
							),
							widget.NewAccordionItem(
								"Indieweb",
								widget.NewForm(
									[]*widget.FormItem{
										{Text: "InReplyTo", Widget: formEntries["InReplyTo"]},
										{Text: "BookmarkOf", Widget: formEntries["BookmarkOf"]},
										{Text: "FavoriteOf", Widget: formEntries["FavoriteOf"]},
										{Text: "RepostOf", Widget: formEntries["RepostOf"]},
										{Text: "LikeOf", Widget: formEntries["LikeOf"]},
									}...,
								),
							),
							widget.NewAccordionItem(
								"Event",
								widget.NewForm(
									[]*widget.FormItem{
										{Text: "Start", Widget: formEntries["Event.Start"]},
										{Text: "End", Widget: formEntries["Event.End"]},
										{Text: "Status", Widget: formSelect["Event.Status"]},
										{Text: "Location", Widget: formEntries["Event.Location"]},
									}...,
								),
							),
							widget.NewAccordionItem(
								"Review",
								container.NewVBox(
									widget.NewForm(
										[]*widget.FormItem{
											{Text: "URL", Widget: formEntries["Item.URL"]},
											{Text: "Name", Widget: formEntries["Item.Name"]},
											rating,
											{Text: "Image", Widget: formEntries["Item.Image"]},
										}...,
									),
								),
							),
						)
						dude.Open(0)
						return dude
					}(),
				),
				func(x bool) {
					if x {
						thisPost.Frontmatter.Title = formEntries["Title"].Text
						thisPost.Frontmatter.Tags = strings.Split(formEntries["Tags"].Text, ",")
					}
				},
				mainWindow,
			)
		}),
		widget.NewToolbarAction(theme.MediaPhotoIcon(), func() {
			LocalFileSelectorWindow()
		}),
		widget.NewToolbarSeparator(),
		// GALLERY
		widget.NewToolbarAction(resourcePicturePng, func() {
			inserts := []fyne.CanvasObject{}
			inputs := []GalleryStruct{}
			for _, bob := range toUpload {
				if bob.IsImage {
					fmt.Printf("Add: %s\n", bob.LocalFile)
					input := GalleryStruct{
						binding.NewString(),
						binding.NewString(),
						binding.NewString(),
					}
					input.Alt.Set("")
					input.Caption.Set("")
					input.Image.Set(bob.RemotePath)
					img := canvas.NewImageFromFile(bob.LocalFile)
					img.SetMinSize(fyne.NewSize(150, 150))
					img.FillMode = canvas.ImageFillContain
					inserts = append(inserts,
						container.New(
							layout.NewFormLayout(),
							img,
							container.New(
								layout.NewFormLayout(),
								widget.NewLabel("Alt Text"),
								widget.NewEntryWithData(input.Alt),
								widget.NewLabel("Caption"),
								widget.NewEntryWithData(input.Caption),
							)),
					)
					inputs = append(inputs, input)
				}
			}
			inserts = append(inserts, container.New(
				layout.NewGridLayout(2),
				widget.NewButton("Insert", func() {
					textToAdd := `<section class="gallery-2020-4" markdown="1">` + "\n"
					for _, bob := range inputs {
						fmt.Printf("Add [%v]", bob)
						x, _ := bob.Image.Get()
						y, _ := bob.Alt.Get()
						z, _ := bob.Caption.Get()
						textToAdd = textToAdd + fmt.Sprintf(
							`[![%s](%s "%s")](%s)`+"\n",
							y,
							"/blog"+getThumbnailFilename(x),
							z,
							"/blog"+x)
					}
					textToAdd = textToAdd + `</section>` + "\n"

					oldClipboard := mainWindow.Clipboard().Content()
					mainWindow.Clipboard().SetContent(string(textToAdd))
					s := &fyne.ShortcutPaste{Clipboard: mainWindow.Clipboard()}
					markdownInput.TypedShortcut(s)
					mainWindow.Clipboard().SetContent(oldClipboard)
					galleryWindow.Hide()
				}),
				widget.NewButton("Cancel", func() {
					galleryWindow.Hide()
				}),
			))
			galleryWindow.SetContent(
				container.New(
					layout.NewVBoxLayout(),
					inserts...,
				),
			)
			galleryWindow.Show()
		}),
		// BLOCKQUOTE
		widget.NewToolbarAction(resourceQuotesPng, func() {
			toChange := markdownInput.SelectedText()
			if len(toChange) == 0 {
				return
			}
			blockQuoteRegex := regexp.MustCompile("\n")
			textToAdd := blockQuoteRegex.ReplaceAll([]byte(toChange), []byte("\n> "))
			oldClipboard := mainWindow.Clipboard().Content()
			mainWindow.Clipboard().SetContent("> " + string(textToAdd))
			s := &fyne.ShortcutPaste{Clipboard: mainWindow.Clipboard()}
			markdownInput.TypedShortcut(s)
			mainWindow.Clipboard().SetContent(oldClipboard)
		}),
	)
	content = container.NewBorder(
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
			container.NewBorder(
				nil,
				nil,
				nil,
				formCheckbox["Mastodon"],
				container.New(
					layout.NewFormLayout(),
					widget.NewLabel("Synopsis"),
					formEntries["Synopsis"],
				),
			),
		),
		nil,
		nil,
		nil,
		container.NewStack(markdownInput),
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

func MakeSliderWithValue(value float32) *widget.Slider {
	b := widget.NewSlider(0, 5)
	b.Step = 0.5
	b.SetValue(float64(value))
	return b
}

func UpdateAllFields(formEntries map[string]*widget.Entry, formSelect map[string]*widget.Select) {
	formEntries["Title"].SetText(thisPost.Frontmatter.Title)
	formEntries["Tags"].SetText(strings.Join(thisPost.Frontmatter.Tags, ","))
	formEntries["Created"].SetText(thisPost.Frontmatter.Created)
	formEntries["Updated"].SetText(thisPost.Frontmatter.Updated)
	formEntries["Synopsis"].SetText(thisPost.Frontmatter.Synopsis)
	formEntries["Mastodon"].SetText(thisPost.Frontmatter.SyndicationLinks.Mastodon)
	formEntries["FeatureImage"].SetText(thisPost.Frontmatter.FeatureImage)
	formEntries["InReplyTo"].SetText(thisPost.Frontmatter.InReplyTo)
	formEntries["BookmarkOf"].SetText(thisPost.Frontmatter.BookmarkOf)
	formEntries["FavoriteOf"].SetText(thisPost.Frontmatter.FavoriteOf)
	formEntries["RepostOf"].SetText(thisPost.Frontmatter.RepostOf)
	formEntries["LikeOf"].SetText(thisPost.Frontmatter.LikeOf)
	//
	if len(thisPost.Frontmatter.Item.Name) > 0 {
		formEntries["Item.URL"].SetText(thisPost.Frontmatter.Item.URL)
		formEntries["Item.Image"].SetText(thisPost.Frontmatter.Item.Image)
		formEntries["Item.Name"].SetText(thisPost.Frontmatter.Item.Name)
		formSlider["Item.Rating"].SetValue(float64(thisPost.Frontmatter.Item.Rating))
	}
	if len(thisPost.Frontmatter.Event.Start) > 0 {
		formEntries["Event.Start"].SetText(thisPost.Frontmatter.Event.Start)
		formEntries["Event.End"].SetText(thisPost.Frontmatter.Event.End)
		formEntries["Event.Location"].SetText(thisPost.Frontmatter.Event.Location)
		formSelect["Event.Status"].SetSelected(thisPost.Frontmatter.Event.Status)
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
	formSelect["Type"].SetSelected(thisPost.Frontmatter.Type)
	formSelect["Status"].SetSelected(thisPost.Frontmatter.Status)
	formCheckbox["Mastodon"].SetChecked(len(thisPost.Frontmatter.SyndicationLinks.Mastodon) > 0)
}

func FieldsToPost(formEntries map[string]*widget.Entry, formSelect map[string]*widget.Select) {
	thisPost.Frontmatter.Title = formEntries["Title"].Text
	thisPost.Frontmatter.Tags = strings.Split(formEntries["Tags"].Text, ",")
	thisPost.Frontmatter.Created = formEntries["Created"].Text
	if len(formEntries["Updated"].Text) == 0 {
		formEntries["Updated"].Text = thisPost.Frontmatter.Updated
	} else {
		thisPost.Frontmatter.Updated = formEntries["Updated"].Text
	}
	thisPost.Frontmatter.Synopsis = formEntries["Synopsis"].Text
	if formCheckbox["Mastodon"].Checked && formEntries["Mastodon"].Text == "" {
		thisPost.Frontmatter.SyndicationLinks.Mastodon = "XPOST"
	} else {
		thisPost.Frontmatter.SyndicationLinks.Mastodon = formEntries["Mastodon"].Text
	}
	thisPost.Frontmatter.FeatureImage = formEntries["FeatureImage"].Text
	thisPost.Frontmatter.InReplyTo = formEntries["InReplyTo"].Text
	thisPost.Frontmatter.BookmarkOf = formEntries["BookmarkOf"].Text
	thisPost.Frontmatter.FavoriteOf = formEntries["FavoriteOf"].Text
	thisPost.Frontmatter.RepostOf = formEntries["RepostOf"].Text
	thisPost.Frontmatter.LikeOf = formEntries["LikeOf"].Text
	if len(thisPost.Frontmatter.Slug) == 0 {
		thisPost.Frontmatter.Slug = cleanName(thisPost.Frontmatter.Title) + ".html"
	}

	// Review
	if len(formEntries["Item.Name"].Text) > 0 {
		thisPost.Frontmatter.Item.Name = formEntries["Item.Name"].Text
		thisPost.Frontmatter.Item.Image = formEntries["Item.Image"].Text
		thisPost.Frontmatter.Item.URL = formEntries["Item.URL"].Text
		thisPost.Frontmatter.Item.Type = "item"
		thisPost.Frontmatter.Item.Rating = float32(formSlider["Item.Rating"].Value)
	} else {
		thisPost.Frontmatter.Item = ItemS{}
	}
	if len(formEntries["Event.Start"].Text) > 0 {
		thisPost.Frontmatter.Event.Start = formEntries["Event.Start"].Text
		thisPost.Frontmatter.Event.End = formEntries["Event.End"].Text
		thisPost.Frontmatter.Event.Location = formEntries["Event.Location"].Text
		thisPost.Frontmatter.Event.Status = formSelect["Event.Status"].Selected
	} else {
		thisPost.Frontmatter.Event = Event{}
	}
	thisPost.Frontmatter.Type = formSelect["Type"].Selected
	thisPost.Frontmatter.Status = formSelect["Status"].Selected
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
			container.NewVScroll(container.NewGridWrap(fyne.NewSize(150, 40), fileFinderContent...)),
			mainWindow,
		)
		fileFinder.Resize(fyne.NewSize(500, 500))
		fileFinder.Show()
	} else {
		// When loading a file to edit, you have to store the sourceCommitId to save later
		contents, e := bitbucket.GetFileContents(thispath)
		if e != nil {
			log.Fatalf("Failed to get file from bitbucket %s", e)
		}
		x, y, e := parseString(contents)
		if e != nil {
			log.Fatalf("Failed to parse file %s\n", e)
		}
		thisPost.Contents = strings.Trim(x, "\n\r ")
		markdownInput.Text = strings.Trim(x, "\r\n ")
		markdownInput.Refresh()
		thisPost.Frontmatter = y
		thisPost.Filename = thispath
		UpdateAllFields(formEntries, formSelect)
	}
}

func LocalFileSelectorWindow() {
	open := dialog.NewFolderOpen(
		func(directory fyne.ListableURI, err error) {
			if directory == nil {
				return
			}
			// Get all files
			files, _ := os.ReadDir(directory.Path())
			checkGroup := widget.NewCheckGroup([]string{}, func(bob []string) {})
			for _, file := range files {
				if !file.IsDir() {
					checkGroup.Append(file.Name())
				}
			}
			uploadPrefix := time.Now().Format("/media/2006/01/02/")
			fileFinder := dialog.NewCustomConfirm(
				"Upload",
				"Upload",
				"Nevermind",
				container.NewVScroll(checkGroup),
				func(ok bool) {
					if ok {
						toUpload = []Attachment{}
						for _, selectedFile := range checkGroup.Selected {
							fullPath := filepath.Join(directory.Path(), selectedFile)
							cleanName := cleanName(selectedFile)
							mimeType, isImage := isFileImage(fullPath)
							toUpload = append(
								toUpload,
								Attachment{
									LocalFile:  fullPath,
									RemotePath: uploadPrefix + strings.ToLower(cleanName),
									MimeType:   mimeType,
									IsImage:    isImage,
								},
							)
						}
					}
				},
				mainWindow,
			)
			fileFinder.Resize(fyne.NewSize(300, 500))
			fileFinder.Show()
		},
		mainWindow,
	)
	x, y := storage.ListerForURI(repository.NewFileURI(thisApp.Preferences().String("startingfolder")))
	if y == nil {
		open.SetLocation(x)
	}

	open.Show()
}

func isFileImage(filename string) (string, bool) {
	clientFile, _ := os.Open(filename)
	defer clientFile.Close()
	buff := make([]byte, 512) // docs tell that it take only first 512 bytes into consideration
	if _, err := clientFile.Read(buff); err != nil {
		return "", false
	}
	mimetype := http.DetectContentType(buff)
	return mimetype, mimetype[:5] == "image"
}

var fileCleanRegexp = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func cleanName(filename string) string {
	slug := fileCleanRegexp.ReplaceAllString(filename, "-")
	slug = strings.Trim(slug, "-")
	return strings.ToLower(slug)
}

func getThumbnailFilename(filename string) string {
	return filename[0:strings.LastIndex(filename, ".")] + "-thumb.jpg"
}

func sliderToResource(value float64) *fyne.StaticResource {
	switch value {
	case 5.0:
		return resource5starPng
	case 4.5:
		return resource45starPng
	case 4:
		return resource4starPng
	case 3.5:
		return resource35starPng
	case 3:
		return resource3starPng
	case 2.5:
		return resource25starPng
	case 2:
		return resource2starPng
	case 1.5:
		return resource15starPng
	case 1:
		return resource1starPng
	case 0.5:
		return resource05starPng
	case 0:
		return resource0starPng
	default:
		return resource0starPng
	}
}
