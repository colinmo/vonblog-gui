package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/pkg/browser"
	"gopkg.in/yaml.v3"
)

type OAuthResponse struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresDate  time.Time
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type BitBucket struct {
	AccessToken  string
	Expiration   time.Time
	RefreshToken string
	ClientID     string
	ClientSecret string
}

var bitbucket = BitBucket{}

type BitBucketAPILinks struct {
	Self struct {
		Href string `json:"href"`
	} `json:"self"`
	Meta struct {
		Href string `json:"href"`
	} `json:"meta"`
	HTML struct {
		Href string `json:"href"`
	} `json:"html"`
}

type Attachment struct {
	LocalFile  string
	RemotePath string
	MimeType   string
	IsImage    bool
}

var toUpload = []Attachment{}

func (b *BitBucket) MakeRequestToTalkToEndpoint(method string, path []string, body *bytes.Reader) *http.Request {
	b.RefreshIfRequired()
	fullUrl, _ := url.JoinPath(
		thisApp.Preferences().String("baseurl"),
		path...,
	)
	request, _ := http.NewRequest(
		method,
		fullUrl,
		body)

	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+b.AccessToken)
	return request
}

func (b *BitBucket) GetFileContents(path string) (string, error) {
	request := b.MakeRequestToTalkToEndpoint(
		"GET",
		[]string{
			`repositories`,
			thisApp.Preferences().String("workspacekey"),
			`/`,
			thisApp.Preferences().String("reposslug"),
			`/src/HEAD/`,
			path,
		},
		bytes.NewReader([]byte("")),
	)
	resp, err := Client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("wrong status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (b *BitBucket) GetFiles(path string) (map[string]string, error) {
	type thisReturnFromAPI struct {
		Pagelen int32 `json:"pagelen"`
		Values  []struct {
			Path   string            `json:"path"`
			Type   string            `json:"type"`
			Links  BitBucketAPILinks `json:"links"`
			Commit struct {
				Type  string            `json:"type"`
				Hash  string            `json:"hash"`
				Links BitBucketAPILinks `json:"links"`
			} `json:"commit"`
		} `json:"values"`
		Page int32  `json:"page"`
		Size int32  `json:"size"`
		Next string `json:"next"`
	}
	toReturn := map[string]string{}
	if len(path) > 1 {
		toReturn[".."] = "x"
	}
	request := b.MakeRequestToTalkToEndpoint(
		"GET",
		[]string{
			`repositories`,
			thisApp.Preferences().String("workspacekey"),
			thisApp.Preferences().String("reposslug"),
			`src`,
			`HEAD`,
			path,
		},
		bytes.NewReader([]byte("")),
	)
	for {
		resp, err := Client.Do(request)
		if err != nil {
			return toReturn, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return toReturn, fmt.Errorf("wrong status code %d", resp.StatusCode)
		}
		var j thisReturnFromAPI
		err = json.NewDecoder(resp.Body).Decode(&j)
		if err != nil {
			return toReturn, fmt.Errorf("failed to parse %v", err)
		}
		for _, x := range j.Values {
			toReturn[x.Path] = x.Commit.Hash
		}
		if len(j.Next) == 0 {
			break
		}
		request.URL, err = url.Parse(j.Next)
		if err != nil {
			break
		}
	}

	return toReturn, nil
}

// file upload help: https://community.atlassian.com/t5/Bitbucket-questions/How-to-commit-multiple-files-from-memory-using-bitbucket-API/qaq-p/1845800
func (b *BitBucket) UploadPost() error {
	if len(thisPost.Filename) == 0 {
		if thisPost.Frontmatter.Type == "page" {
			thisPost.Filename = "posts/" + thisPost.Frontmatter.Type + cleanName(thisPost.Frontmatter.Title) + ".md"
		} else {
			thisPost.Filename = "posts/" + thisPost.Frontmatter.Type + "/" + time.Now().Format("2006/01/02/") + cleanName(thisPost.Frontmatter.Title) + ".md"
		}
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	content, _ := yaml.Marshal(thisPost.Frontmatter)
	writer.WriteField("message", "Post from exec - "+time.Now().Format(dateFormatString))
	writer.WriteField(thisPost.Filename, "---\n"+string(content)+"---\n"+thisPost.Contents)

	for _, z := range toUpload {
		part, _ := writer.CreateFormFile(z.RemotePath, z.RemotePath)
		b, _ := os.Open(z.LocalFile)
		defer b.Close()
		io.Copy(part, b)
		if z.IsImage {
			img, err := readImage(z.LocalFile)
			if err == nil {
				img = resize.Thumbnail(480, 480, img, resize.Lanczos3)
				thumbnailFile := getThumbnailFilename(z.RemotePath)
				part2, _ := writer.CreateFormFile(thumbnailFile, thumbnailFile)
				jpOp := jpeg.Options{
					Quality: 90,
				}
				jpeg.Encode(part2, img, &jpOp)
			}
		}
	}
	writer.Close()
	z := body.Bytes()
	body.Reset()
	body.Write(z)
	request := b.MakeRequestToTalkToEndpoint(
		"POST",
		[]string{
			`repositories`,
			thisApp.Preferences().String("workspacekey"),
			thisApp.Preferences().String("reposslug"),
			`src`,
		},
		bytes.NewReader(body.Bytes()),
	)
	request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	request.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	request.Header.Set("Accept", "application/json")
	resp, err := Client.Do(request)
	if err != nil {
		fmt.Printf("Failure %v\n", err)
	}

	defer resp.Body.Close()
	//var j interface{}
	//json.NewDecoder(resp.Body).Decode(&j)
	if resp.StatusCode != 201 {
		return fmt.Errorf(
			"did not save, %d\n%s\n%s",
			resp.StatusCode,
			resp.Request.URL.RequestURI(),
			z,
		)
	}
	// If the Syndication.Mastodon == 'XPOST' then poll until we get an updated commit
	for thisPost.Frontmatter.SyndicationLinks.Mastodon == "XPOST" {
		request := b.MakeRequestToTalkToEndpoint(
			"GET",
			[]string{
				`repositories`,
				thisApp.Preferences().String("workspacekey"),
				thisApp.Preferences().String("reposslug"),
				`commits`,
			},
			bytes.NewReader(body.Bytes()),
		)
		request.Header.Set("Accept", "application/json")
		resp, err := Client.Do(request)
		if err != nil {
			fmt.Printf("Failure %v\n", err)
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf(
				"failed to get response from XPOST change, %d\n%s\n%s",
				resp.StatusCode,
				resp.Request.URL.RequestURI(),
				z,
			)
		}
		var j struct {
			Values []struct {
				Message string `json:message`
			} `json:values`
		}
		json.NewDecoder(resp.Body).Decode(&j)
		if j.Values[0].Message[0:4] == "XPOST" {
			thisPost.Frontmatter.SyndicationLinks.Mastodon = j.Values[0].Message[7:]
			return nil
		}
	}
	return nil
}

func (b *BitBucket) Authenticate(w http.ResponseWriter, r *http.Request) {
	var OToken OAuthResponse
	query := r.URL.Query()
	if query.Get("code") != "" {
		payload := url.Values{
			"code":       {query.Get("code")},
			"grant_type": {"authorization_code"},
		}
		req, _ := http.NewRequest(
			http.MethodPost,
			`https://bitbucket.org/site/oauth2/access_token`,
			strings.NewReader(payload.Encode()),
		)
		req.Header.Set("Content-type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(thisApp.Preferences().String("clientkey"), thisApp.Preferences().String("clientsecret"))
		resp, err := Client.Do(req)

		if err != nil {
			log.Fatalf("Login failed %s\n", err)
		} else {
			defer resp.Body.Close()
			err := json.NewDecoder(resp.Body).Decode(&OToken)
			if err != nil {
				log.Fatalf("Failed MS %s\n", err)
			}
			seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", OToken.ExpiresIn-10))
			b.RefreshToken = OToken.RefreshToken
			b.Expiration = time.Now().Add(seconds)
			b.AccessToken = OToken.AccessToken
			thisApp.Preferences().SetString("refreshtoken", b.RefreshToken)
			thisApp.Preferences().SetString("accesstoken", b.AccessToken)
			thisApp.Preferences().SetString("expiration", b.Expiration.Format(dateFormatString))
			w.Header().Add("Content-type", "text/html")
			fmt.Fprintf(w, "<html><head></head><body><H1>Authenticated<p>You are authenticated, you may close this window.</body></html>")
		}
	}
}

func (b *BitBucket) RefreshIfRequired() {
	expiration := thisApp.Preferences().String("expiration")
	if len(expiration) == 0 {
		b.Refresh()
	} else {
		expirationDate, err := time.Parse(expiration, dateFormatString)
		if err != nil {
			b.Refresh()
		} else {
			if time.Now().After(expirationDate) {
				b.Refresh()
			}
		}
	}
}

func (b *BitBucket) Refresh() {
	var OToken OAuthResponse
	payload := url.Values{
		"refresh_token": {thisApp.Preferences().String("refreshtoken")},
		"grant_type":    {"refresh_token"},
	}
	req, _ := http.NewRequest(
		http.MethodPost,
		`https://bitbucket.org/site/oauth2/access_token`,
		strings.NewReader(payload.Encode()),
	)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(thisApp.Preferences().String("clientkey"), thisApp.Preferences().String("clientsecret"))
	resp, err := Client.Do(req)

	if err != nil {
		log.Fatalf("Login failed %s\n", err)
	} else {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&OToken)
		if err != nil {
			log.Fatalf("Failed MS %s\n", err)
		}
		seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", OToken.ExpiresIn-10))
		b.RefreshToken = OToken.RefreshToken
		b.Expiration = time.Now().Add(seconds)
		b.AccessToken = OToken.AccessToken
		thisApp.Preferences().SetString("refreshtoken", b.RefreshToken)
		thisApp.Preferences().SetString("accesstoken", b.AccessToken)
		thisApp.Preferences().SetString("expiration", b.Expiration.Format(dateFormatString))
	}
}

func (b *BitBucket) Login() {
	browser.OpenURL(
		fmt.Sprintf(`https://bitbucket.org/site/oauth2/authorize?client_id=%s&response_type=code`,
			thisApp.Preferences().String("clientkey"),
		),
	)
}

var AuthWebServer *http.Server

func startLocalServers() {
	http.HandleFunc("/bitbucket/", func(w http.ResponseWriter, r *http.Request) {
		bitbucket.Authenticate(w, r)
	})
	go func() {
		AuthWebServer = &http.Server{Addr: ":1025", Handler: nil}
		if err := AuthWebServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}

func readImage(name string) (image.Image, error) {
	fd, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	img, _, err := image.Decode(fd)
	if err != nil {
		return nil, err
	}
	return img, nil
}
