package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/blang/semver"
)

// Tag is the individual tag from gitlab, with addition of Version.
type Tag struct {
	Version semver.Version `json:"-"`
	Name    string         `json:"name"`
	Message string         `json:"message"`
}

// Tags is the array of gitlab tags.
type Tags []Tag

func (a Tags) Len() int      { return len(a) }
func (a Tags) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// This is a reverse sort - most recent first, hence > instead of <.
func (a Tags) Less(i, j int) bool { return a[i].Name > a[j].Name }

var (
	baseURL    string
	token      string
	org        string
	repo       string
	namePrefix string
	insecure   bool
	sortSemver bool
	since      string
)

func init() {
	flag.StringVar(&baseURL, "url", "", "Base GitLab URL formatted as https://gitlab.example.com/")
	flag.StringVar(&token, "token", "", "Personal access token (create one in your GitLab instance at '/profile/personal_access_tokens'; be sure to check 'Api: Access your API')")
	flag.StringVar(&org, "org", "", "Organization name")
	flag.StringVar(&repo, "repo", "", "Repository name")
	flag.StringVar(&namePrefix, "version-prefix", "", "Text to put before the version name (e.g. '#' for markdown header)")
	flag.BoolVar(&insecure, "insecure", false, "Do not check the server's certificate")
	flag.BoolVar(&sortSemver, "sort-semver", true, "Sort by tag name according to semantic versioning from most recent to oldest")
	flag.StringVar(&since, "since-tag", "0.0.0", "Print tags that are greater than or equal to the specified semantic version (e.g. 1.0.0 will show all tags/messages since 1.0.0)")
}

func main() {

	flag.Parse()

	if baseURL == "" || org == "" || repo == "" {
		log.Fatal("Please define the url, token, org, and repo.")
	}

	sinceVers, err := semver.Parse(since)
	if err != nil {
		log.Fatalf("unable to parse since version %s: %s", since, err)
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	url, err := url.Parse(baseURL + "api/v4/projects/" + org + "%2F" + repo + "/repository/tags")
	if err != nil {
		log.Fatalf("error parsing url %s: %s", baseURL, err)
	}

	tr := &http.Transport{}
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Fatalf("error creating request for url %s: %s", url.String(), err)
	}
	req.Header.Add("PRIVATE-TOKEN", token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error getting url %s: %s", url.String(), err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body for url %s: %s", url.String(), err)
	}

	// Check that the response is valid JSON array.
	if !bytes.HasPrefix(body, []byte("[")) || !bytes.HasSuffix(body, []byte("]")) {
		log.Fatalf("response was not valid; if this is a private repo, did you specify a token?\nResponse: %s", string(body))
	}

	var jsonResp []Tag
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		log.Fatalf("error decoding json for url %s: %s", url.String(), err)
	}

	var errors string
	var tags = make(Tags, len(jsonResp))
	for i, tag := range jsonResp {
		t := Tag{
			Name:    tag.Name,
			Message: tag.Message,
		}
		if sortSemver {
			n := strings.Replace(tag.Name, "v", "", 1)
			vers, err := semver.Make(n)
			if err != nil {
				errors += fmt.Sprintf("error parsing tag %s: %s\n\n", tag.Name, err)
			} else {
				t.Version = vers
			}
		}
		tags[i] = t
	}

	if sortSemver {
		sort.Sort(tags)
	}

	for _, tag := range tags {
		if sortSemver {
			if tag.Version.GTE(sinceVers) {
				fmt.Printf("%s %s\n%s\n\n", namePrefix, tag.Name, tag.Message)
			}
		} else {
			fmt.Printf("%s %s\n%s\n\n", namePrefix, tag.Name, tag.Message)
		}
	}

	if errors != "" {
		fmt.Fprintf(os.Stderr, "\n\nErrors parsing semver tags:\n%s", errors)
	}

}
