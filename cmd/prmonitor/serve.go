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

var client *github.Client

func main() {
	t := prmonitor.Config{}
	err := json.Unmarshal([]byte(os.Getenv("CONFIG")), &t)
	if err != nil {
		panic(err)
	}

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(t.GithubUser),
		Password: strings.TrimSpace(t.GithubPass),
	}
	client = github.NewClient(tp.Client())

	http.HandleFunc("/", prmonitor.SSLRequired(os.Getenv("SSLHOST"), prmonitor.BasicAuth(t.DashboardUser, t.DashboardPass, prmonitor.Dashboard(t, client))))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
