package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// @todo: Pagination
func getFiles(path string) ([]string, error) {
	data := url.Values{
		//		"status":     {message},
	}

	fullUrl, _ := url.JoinPath(baseurl, `/rest/api/latest/projects/`, projectkey, `/repos/`, reposslug, `/files`, path)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte(data.Encode())),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(username, password)
	resp, err := Client.Do(request)
	if err != nil {
		return []string{}, err
	}
	fmt.Printf("Response: %v\n", resp)
	return []string{}, nil
}

// @todo: Pagination
func getProjects() ([]string, error) {
	fullUrl, _ := url.JoinPath(baseurl, `repositories/`, projectkey, `/`, reposslug, `/src/HEAD/`)
	request, _ := http.NewRequest(
		"GET",
		fullUrl,
		bytes.NewBuffer([]byte("")),
	)
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(username, password)
	resp, err := Client.Do(request)
	if err != nil {
		return []string{}, err
	}
	//fmt.Printf("Response: %v\n", resp)
	defer resp.Body.Close()
	var j interface{}
	err = json.NewDecoder(resp.Body).Decode(&j)
	fmt.Printf("Response: %s\n", j)
	return []string{}, err
}
