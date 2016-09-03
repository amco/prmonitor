package main

import (
	"github.com/google/go-github/github"
	"fmt"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"time"
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

func main() {

	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	t := Config{}
	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		panic(err)
	}

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(t.Username),
		Password: strings.TrimSpace(t.Password),
	}

	client := github.NewClient(tp.Client())

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
		fmt.Printf("%s/%s\n", r.Owner, r.Repo)
		for _, v := range oprs {
			start := *v.CreatedAt
			user := *v.User
			if r.Authors != nil {
				// match just some authors
				for _, a := range *r.Authors {
					if *user.Login == a {
						// print for single author
						fmt.Printf("  #%d %s by %s @ %s\n", *v.Number, *v.Title, *user.Login, start.Format(time.RFC3339))
					}
				}
			} else {
				// match all
				fmt.Printf("  #%d %s by %s @ %s\n", *v.Number, *v.Title, *user.Login, start.Format(time.RFC3339))
			}
		}
	}
}
