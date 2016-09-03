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
	Repos []Repo
}

type Repo struct {
	Owner string
	Repo string
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
		op.PerPage = 30
		op.Page = 0
		oprs, _, err := client.PullRequests.List(r.Owner, r.Repo, op)
		if err != nil {
			panic(err)
		}

		// make display a bit more readable
		fmt.Printf("%s/%s\n", r.Owner, r.Repo)
		for _, v := range oprs {
			var start, end time.Time
			if v.CreatedAt == nil {
				panic(fmt.Errorf("no createdat date"))
			} else {
				start = *v.CreatedAt
			}
			fmt.Printf("  #%d %s [%s to %s]\n", *v.Number, *v.Title, start.Format(time.RFC3339), end.Format(time.RFC3339))
		}
	}
}
