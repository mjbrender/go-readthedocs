package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const keyword = "snap-plugin"

var (
	personalAccessToken string
	org                 string
)

// TokenSource is an encapsulation of the AccessToken string
type TokenSource struct {
	AccessToken string
}

// RepositoryContentGetOptions represents an optional ref parameter
type RepositoryContentGetOptions struct {
	Ref string `url:"ref,omitempty"`
}

// Token authenticates via oauth
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func main() {
	org = os.Getenv("GH_ORG")
	personalAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")

	if len(personalAccessToken) == 0 {
		log.Fatal("Before you can use this you must set the GITHUB_ACCESS_TOKEN environment variable.")
	}
	if len(org) < 1 {
		log.Fatal("You need to have a single organization name set to GH_ORG environmental variable.")
	}

	tokenSource := &TokenSource{
		AccessToken: personalAccessToken,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := github.NewClient(oauthClient) // authenticated to GitHub here

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	// get all pages of results
	var allRepos []github.Repository

	// initialize the map of all Readmes
	readmeLibrary := make(map[string]string)
	for {
		repos, resp, err := client.Repositories.ListByOrg(org, opt)
		if err != nil {
			log.Fatal(err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	for _, rp := range allRepos {
		repo := *rp.Name

		// Going through a list of all repos for GH_ORG
		if strings.Contains(repo, keyword) {

			// error handling on retrieval of README.md.
			encodedText, _, err := client.Repositories.GetReadme(org, repo, &github.RepositoryContentGetOptions{})
			if err != nil {
				log.Printf("Repositories.GetReadme returned error: %v\n", err)
			}
			if encodedText == nil {
				log.Printf("The returned text from %v is nil. Are you sure it exists?\n", repo)
			}

			// encoding could have some issues. Let's catch them here.
			text, err := encodedText.Decode()
			if err != nil {
				log.Printf("Decoding failed: %v\n", err)
			}

			// conversion of the decoded file to string
			readme := string(text)
			readmeLibrary[repo] = readme
			fmt.Printf("Found a readme for %v\n", repo)

			// have decoded README.md here.

			// for testing. Just mess with one file for now.
			break
		}
	}
	// work out here will occur after we have all the readmes.
	fmt.Println(parseReadme(&readmeLibrary))
}
