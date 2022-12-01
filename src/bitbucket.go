package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// @todo: Pagination
// @todo: run https://api.bitbucket.org/2.0/repositories/vonexplaino/blog/src first to get the commit number, then use the commit number provided by the 302 redirect
func (b *BitBucket) GetFiles(path string) ([]string, error) {
	data := url.Values{
		//		"status":     {message},
	}

	fullUrl, _ := url.JoinPath(baseurl, `/repositories/`, workspacekey, `/`, reposslug, `/src/HEAD/`, path)
	fmt.Printf("URL: %s\n", fullUrl)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte(data.Encode())),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+b.AccessToken)
	resp, err := Client.Do(request)
	if err != nil {
		return []string{}, err
	}
	fmt.Printf("Response: %v\n", resp)
	return []string{}, nil
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
			b.GetFiles("/")
			fmt.Printf("\nAT: %s\n", b.AccessToken)
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
