package main

import (
	"github.com/google/go-github/github"
	"fmt"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"time"
	"net/http"
	"log"
)

type Config struct {
	Username string
	Password string

	// Repos to pull onto dashboard
	Repos []Repo
}

type Repo struct {
	Owner string
	Repo string

	// number of open PRs to look through - can be tuned for each repo.
	Depth int

	// optional list of authors - if included, will only display open PRs
	// by those authors.
	Authors *[]string
}

var t Config

var client *github.Client

func main() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	t = Config{}
	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		panic(err)
	}

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(t.Username),
		Password: strings.TrimSpace(t.Password),
	}
	client = github.NewClient(tp.Client())

	http.HandleFunc("/", dashboard)
	log.Fatal(http.ListenAndServe(":8080", nil))
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

func render(w http.ResponseWriter, owner string, repo string, number int, title string, author string, hours time.Duration) {
	n := hours.Hours() / (240*time.Hour).Hours()
	if n > 1 {
		n = 1
	}

	stopyellow := (24*time.Hour).Hours()

	stopred := (24*3*time.Hour).Hours()

	color := "#777"
	if hours.Hours() > stopred {
		color = "#FF4500"
	} else if hours.Hours() > stopyellow {
		color = "#FFA500"
	}


	style := fmt.Sprintf(`margin: 3px; padding: 8px; background: linear-gradient( 90deg, %s %d%%, #333 %d%%);`, color, int(n*100), int(n*100))
	fmt.Fprintf(w, "<div style='%s'><b>%s/%s</b> #%d %s by %s @ %d days or %d hours</div>", style, owner, repo, number, title, author, hours/(24*time.Hour), hours/time.Hour)
}