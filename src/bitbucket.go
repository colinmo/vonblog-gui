package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
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
	IsImage    bool
}

func (b *BitBucket) GetFileContents(path string) (string, error) {
	fullUrl, _ := url.JoinPath(baseurl, `/repositories/`, workspacekey, `/`, reposslug, `/src/HEAD/`, path)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte("")),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+b.AccessToken)
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
	fullUrl, _ := url.JoinPath(baseurl, `/repositories/`, workspacekey, `/`, reposslug, `/src/HEAD/`, path)
	for len(fullUrl) > 0 {
		fmt.Printf("URL: %s\n", fullUrl)
		request, _ := http.NewRequest(
			"GET",
			fullUrl,
			bytes.NewBuffer([]byte("")),
		)
		request.Header.Set("Content-type", "application/x-www-form-urlencoded")
		request.Header.Set("Authorization", "Bearer "+b.AccessToken)
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
		fullUrl = j.Next
	}

	return toReturn, nil
}

// @todo: Pagination
func (b *BitBucket) GetProjects() ([]string, error) {
	fullUrl, _ := url.JoinPath(baseurl, `repositories/`, workspacekey, `/`, reposslug, `/src/HEAD/`)
	fmt.Printf("Url: %s\n", fullUrl)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte("")),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+b.AccessToken)
	resp, err := Client.Do(request)
	if err != nil {
		return []string{}, err
	}
	//fmt.Printf("Response: %v\n", resp)
	defer resp.Body.Close()
	var j interface{}
	err = json.NewDecoder(resp.Body).Decode(&j)
	fmt.Printf("Code: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", j)
	fmt.Printf("Error: %s\n", err)
	return []string{}, err
}

func (b *BitBucket) GetCurrentUser() (string, error) {
	fullUrl, _ := url.JoinPath(baseurl, `user`)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte("")),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+b.AccessToken)
	resp, err := Client.Do(request)
	if err != nil {
		return "", err
	}
	//fmt.Printf("Response: %v\n", resp)
	defer resp.Body.Close()
	var j interface{}
	err = json.NewDecoder(resp.Body).Decode(&j)
	fmt.Printf("Code: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", j)
	fmt.Printf("Error: %s\n", err)
	return "", err

}

func (b *BitBucket) GetUserWorkspaces() {
	fullUrl, _ := url.JoinPath(baseurl, `user/permissions/workspaces`)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte("")),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+b.AccessToken)
	resp, err := Client.Do(request)
	if err != nil {
		log.Fatal("Bah")
	}
	//fmt.Printf("Response: %v\n", resp)
	defer resp.Body.Close()
	var j interface{}
	err = json.NewDecoder(resp.Body).Decode(&j)
	fmt.Printf("Code: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", j)
	fmt.Printf("Error: %s\n", err)
	log.Fatal("Bah")
}

func (b *BitBucket) GetRepositories() {
	fullUrl, _ := url.JoinPath(baseurl, `repositories/`+workspacekey)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte("")),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+b.AccessToken)
	resp, err := Client.Do(request)
	if err != nil {
		log.Fatal("Bah")
	}
	//fmt.Printf("Response: %v\n", resp)
	defer resp.Body.Close()
	var j interface{}
	err = json.NewDecoder(resp.Body).Decode(&j)
	fmt.Printf("Code: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", j)
	fmt.Printf("Error: %s\n", err)
	log.Fatal("Bah")
}

func (b *BitBucket) UploadPost() {

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
		req.SetBasicAuth(clientkey, clientsecret)
		resp, err := Client.Do(req)

		if err != nil {
			log.Fatalf("Login failed %s\n", err)
		} else {
			defer resp.Body.Close()
			err := json.NewDecoder(resp.Body).Decode(&OToken)
			if err != nil {
				log.Fatalf("Failed MS %s\n", err)
			}
			b.RefreshToken = OToken.RefreshToken
			seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", OToken.ExpiresIn-10))
			b.Expiration = time.Now().Add(seconds)
			w.Header().Add("Content-type", "text/html")
			fmt.Fprintf(w, "<html><head></head><body><H1>Authenticated<p>You are authenticated, you may close this window.</body></html>")
			b.AccessToken = OToken.AccessToken
			/*
				b.GetFiles("/")
				x, _ := b.GetFileContents("/posts/article/NearlyThere.md")
				fmt.Printf("FileContents: %s\n", x)
				fmt.Printf("\nAT: %s\n", b.AccessToken)
			*/
		}
	}
}

func (b *BitBucket) Login() {
	browser.OpenURL(
		fmt.Sprintf(`https://bitbucket.org/site/oauth2/authorize?client_id=%s&response_type=code`,
			clientkey,
		),
	)
}

var AuthWebServer *http.Server

func startLocalServers() {
	http.HandleFunc("/bitbucket/", func(w http.ResponseWriter, r *http.Request) {
		bitbucket.Authenticate(w, r)
	})
	go func() {
		AuthWebServer = &http.Server{Addr: ":85", Handler: nil}
		if err := AuthWebServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}
