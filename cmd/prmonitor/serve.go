package main

import (
	"encoding/json"
	"fmt"
	"github.com/brentdrich/prmonitor"
	"github.com/google/go-github/github"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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

	http.HandleFunc("/", prmonitor.SSLRequired(prmonitor.BasicAuth(t.DashboardUser, t.DashboardPass, dashboard)))
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

		// make display a bit more readable
		fmt.Fprintf(w, "<html><body style='background: #333; color: #fff;'>")
		for _, v := range oprs {
			start := *v.CreatedAt
			user := *v.User
			if r.Authors != nil {
				// match just some authors
				for _, a := range *r.Authors {
					if *user.Login == a {
						// print for single author
						render(w, r.Owner, r.Repo, *v.Number, *v.Title, *user.Login, time.Since(start))
					}
				}
			} else {
				// match all
				render(w, r.Owner, r.Repo, *v.Number, *v.Title, *user.Login, time.Since(start))
			}
		}
		fmt.Fprintf(w, "</body></html>")
	}
}

func render(w io.Writer, owner string, repo string, number int, title string, author string, hours time.Duration) {
	n := hours.Hours() / (240 * time.Hour).Hours()
	if n > 1 {
		n = 1
	}

	stopyellow := (24 * time.Hour).Hours()

	stopred := (24 * 3 * time.Hour).Hours()

	color := "#777"
	if hours.Hours() > stopred {
		color = "#FF4500"
	} else if hours.Hours() > stopyellow {
		color = "#FFA500"
	}

	style := fmt.Sprintf(`margin: 3px; padding: 8px; background: linear-gradient( 90deg, %s %d%%, #333 %d%%);`, color, int(n*100), int(n*100))
	fmt.Fprintf(w, "<div style='%s'><b>%s/%s</b> #%d %s by %s @ %d days or %d hours</div>", style, owner, repo, number, title, author, hours/(24*time.Hour), hours/time.Hour)

	return
}
