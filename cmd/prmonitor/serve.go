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

	defaultCustom := prmonitor.GetCustomizations()
	if t.Customization.PassiveColor == "" {
		t.Customization.PassiveColor = defaultCustom.PassiveColor
	}
	if t.Customization.WarningColor == "" {
		t.Customization.WarningColor = defaultCustom.WarningColor
	}
	if t.Customization.AlertColor == "" {
		t.Customization.AlertColor = defaultCustom.AlertColor
	}
	if t.Customization.PassiveTime == 0 {
		t.Customization.PassiveTime = defaultCustom.PassiveTime
	}
	if t.Customization.WarningTime == 0 {
		t.Customization.WarningTime = defaultCustom.WarningTime
	}

	http.HandleFunc("/", prmonitor.BasicAuth(t.DashboardUser, t.DashboardPass, prmonitor.Timestamp(prmonitor.Dashboard(t, client))))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
