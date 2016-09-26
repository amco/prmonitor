package main

import (
	"encoding/json"
	"fmt"
	"github.com/brentdrich/prmonitor"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strings"
)

var client *github.Client

func main() {
	t := prmonitor.Config{
		DashboardUser: strings.TrimSpace(os.Getenv("DASHBOARD_USER")),
		DashboardPass: strings.TrimSpace(os.Getenv("DASHBOARD_PASSWORD")),
		GithubUser:    strings.TrimSpace(os.Getenv("GITHUB_USER")),
		GithubPass:    strings.TrimSpace(os.Getenv("GITHUB_PASSWORD")),
		GithubToken:   strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
	}
	err := json.Unmarshal([]byte(os.Getenv("CONFIG")), &t)
	if err != nil {
		panic(err)
	}

	var client *github.Client
	if t.GithubToken == "" {
		tp := github.BasicAuthTransport{
			Username: t.GithubUser,
			Password: t.GithubPass,
		}
		client = github.NewClient(tp.Client())
	} else {
		tc := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: t.GithubToken},
		))
		client = github.NewClient(tc)
	}

	http.HandleFunc("/", prmonitor.SSLRequired(os.Getenv("SSLHOST"), prmonitor.BasicAuth(t.DashboardUser, t.DashboardPass, prmonitor.Timestamp(prmonitor.Dashboard(t, client)))))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
