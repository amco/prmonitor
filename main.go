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
		oprs, _, err := client.PullRequests.List(r.Owner, r.Repo, op)
		if err != nil {
			panic(err)
		}

		// make display a bit more readable
		fmt.Fprintf(w, "<b>%s/%s</b><ul>", r.Owner, r.Repo)
		for _, v := range oprs {
			start := *v.CreatedAt
			user := *v.User
			if r.Authors != nil {
				// match just some authors
				for _, a := range *r.Authors {
					if *user.Login == a {
						// print for single author
						fmt.Fprintf(w, "<li>#%d %s by %s @ %s</li>", *v.Number, *v.Title, *user.Login, start.Format(time.RFC3339))
					}
				}
			} else {
				// match all
				fmt.Fprintf(w, "<li>#%d %s by %s @ %s</li>", *v.Number, *v.Title, *user.Login, start.Format(time.RFC3339))
			}
		}
		fmt.Fprintf(w, "</ul>")
	}

}