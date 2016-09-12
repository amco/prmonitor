package main

import (
	"encoding/json"
	"fmt"
	"github.com/brentdrich/prmonitor"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"os"
	"strings"
)

// Config contains deserialized configuration file information
// that tells the prmonitor which repos to monitor and which
// credentials to use when accessing github.
type Config struct {
	// Dashboard user
	DashboardUser string `json:"dashboard_user"`
	DashboardPass string `json:"dashboard_pass"`

	// Github API user
	GithubUser string `json:"github_user"`
	GithubPass string `json:"github_pass"`

	// Repos to pull onto dashboard
	Repos []Repo
}

// Repo is a single repository that should be monitored and the
// specific settings that should work for the repository. For
// example, high churn repos may need to be filtered by author
// and fetch more open PRs (since this only grabs the first page).
type Repo struct {
	Owner string
	Repo  string

	// number of open PRs to look through - can be tuned for each repo.
	Depth int

	// optional list of authors - if included, will only display open PRs
	// by those authors.
	Authors *[]string
}

var t Config

var client *github.Client

func main() {
	t = Config{}
	err := json.Unmarshal([]byte(os.Getenv("CONFIG")), &t)
	if err != nil {
		panic(err)
	}

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(t.GithubUser),
		Password: strings.TrimSpace(t.GithubPass),
	}
	client = github.NewClient(tp.Client())

	http.HandleFunc("/", prmonitor.SSLRequired(os.Getenv("SSLHOST"), prmonitor.BasicAuth(t.DashboardUser, t.DashboardPass, dashboard)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	for _, r := range t.Repos {
		// get 30 latest open pull requests.
		op := &github.PullRequestListOptions{}
		op.State = "open"
		op.Sort = "created"
		op.Direction = "desc"
		op.PerPage = r.Depth
		op.Page = 0
		op.Base = "master"
		oprs, _, err := client.PullRequests.List(r.Owner, r.Repo, op)
		if err != nil {
			panic(err)
		}

		// load up the prs
		var prs []prmonitor.SummarizedPullRequest{}
		for _, v := range oprs {
			start := *v.CreatedAt
			user := *v.User
			if r.Authors != nil {
				// match just some authors
				for _, a := range *r.Authors {
					if *user.Login == a {
						// print for single author
						prs = append(prs, prmonitor.SummarizedPullRequest{
							Owner: r.Owner,
							Repo: r.Repo,
							Number: *v.Number,
							Title: *v.Title,
							Author: *user.Login,
							OpenedAt: start,
						})
					}
				}
			} else {
				// match all
				prs = append(prs, prmonitor.SummarizedPullRequest{
					Owner: r.Owner,
					Repo: r.Repo,
					Number: *v.Number,
					Title: *v.Title,
					Author: *user.Login,
					OpenedAt: start,
				})
			}
		}

		// render the dashboard
		prmonitor.Render(w, prs)
	}
}