package main

import (
	"github.com/google/go-github/github"
	"fmt"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
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
	prs, _, err := client.PullRequests.List(t.Owner, t.Repo, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", prs)
}
