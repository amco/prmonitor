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

	// get 30 latest closed pull requests.
	op := &github.PullRequestListOptions{}
	op.State = "open"
	op.Sort = "created"
	op.Direction = "desc"
	op.PerPage = 30
	op.Page = 0
	oprs, _, err := client.PullRequests.List(t.Owner, t.Repo, op)
	if err != nil {
		panic(err)
	}

	// get 30 latest closed pull requests.
	cp := &github.PullRequestListOptions{}
	cp.State = "closed"
	cp.Sort = "created"
	cp.Direction = "desc"
	cp.PerPage = 30
	cp.Page = 0
	cprs, _, err := client.PullRequests.List(t.Owner, t.Repo, cp)

	// make display a bit more readable
	fmt.Println("open prs:")
	for _, v := range oprs {
		var start, end time.Time
		if v.CreatedAt == nil {
			panic(fmt.Errorf("no createdat date"))
		} else {
			start = *v.CreatedAt
		}
		if v.ClosedAt == nil {
			end = time.Now()
		} else {
			end = *v.ClosedAt
		}
		fmt.Printf("  #%d %s [%s to %s]\n", *v.Number, *v.Title, start.Format(time.RFC3339), end.Format(time.RFC3339))
	}
	fmt.Println("closed prs:")
	for _, v := range cprs {
		var start, end time.Time
		if v.CreatedAt == nil {
			panic(fmt.Errorf("no createdat date"))
		} else {
			start = *v.CreatedAt
		}
		if v.ClosedAt == nil {
			end = time.Now()
		} else {
			end = *v.ClosedAt
		}
		fmt.Printf("  #%d %s [%s to %s]\n", *v.Number, *v.Title, start.Format(time.RFC3339), end.Format(time.RFC3339))
	}
}
