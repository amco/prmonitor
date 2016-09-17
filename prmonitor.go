package prmonitor

import (
	"encoding/base64"
	"fmt"
	"github.com/google/go-github/github"
	"io"
	"net/http"
	"time"
)

// Data Structures

type pipelinePR struct {
	Owner string
	Repo  string
	PR    *github.PullRequest
}

// SummarizedPullRequest contains information necessary to
// render a PR.
type SummarizedPullRequest struct {
	// the organization that owns the repository.
	Owner string

	// the repository name that the PR is located in.
	Repo string

	// the visible github PR #
	Number int

	// the title of the PR
	Title string

	// the username of the author of the PR.
	Author string

	// the time the PR was opened.
	OpenedAt time.Time
}

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

// Middlewares

// BasicAuth forces use of browser basic auth. If auth credentials aren't
// presented or are invalid, it sends back a WWW-Authenticate header to
// get the browser to prompt the user to enter credentials.
func BasicAuth(username string, password string, next http.HandlerFunc) http.HandlerFunc {
	match := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != match {
			w.Header().Set("WWW-Authenticate", "Basic")
			w.WriteHeader(401)
			return
		}
		next(w, r)
	}
}

// SSLRequired redirects to an SSL host if the incoming request was
// not made with HTTPS. When placed before the BasicAuth middleware,
// it ensures the basic auth handshake doesn't occur outside of https
// so credentials aren't sent as plaintext.
func SSLRequired(sslhost string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-Proto") != "https" {
			w.Header().Set("Location", sslhost)
			w.WriteHeader(301)
			return
		}
		next(w, r)
	}
}

// Timestamp adds a request header that includes an RFC3339 timestamp.
// The purpose of this middleware is to eliminate the need to use time.Now
// in the rest of the application, placing time sensitive calculations
// under better control.
func Timestamp(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Timestamp", time.Now().Format(time.RFC3339))
		next(w, r)
	}
}

// Dashboard responds to an http request with a dashboard displaying
// the configured pull requests by pulling information down from
// github.
func Dashboard(t Config, client *github.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now, err := time.Parse(time.RFC3339, r.Header.Get("X-Timestamp"))
		if err != nil {
			panic(err)
		}

		re := make(chan Repo)
		p := make(chan pipelinePR)
		f := make(chan SummarizedPullRequest)
		done := make(chan bool)

		// serial pipeline
		go Retrieve(re, p, client)
		go Filter(p, f)
		go Display(f, done, w, now)
		for _, repo := range t.Repos {
			re <- repo
		}
		close(re)

		// wait to respond until dashboard is rendered
		<-done
	}
}

// Data processing pipeline...

// Retrieve pulls in a repository and fetches pull requests that
// are passed to the next stage in the pipeline.
func Retrieve(in chan Repo, out chan pipelinePR, client *github.Client) {
	for {
		r, more := <-in
		if more {
			op := &github.PullRequestListOptions{}
			op.State = "open"
			op.Sort = "created"
			op.Direction = "desc"
			op.Page = 0
			op.Base = "master"
			op.PerPage = r.Depth
			oprs, _, err := client.PullRequests.List(r.Owner, r.Repo, op)
			if err != nil {
				return
			}
			for _, v := range oprs {
				out <- pipelinePR{
					Owner: r.Owner,
					Repo:  r.Repo,
					PR:    v,
				}
			}
		} else {
			close(out)
			return
		}
	}
}

// Filter reads PRs coming in from github and writes out summaries
// that can be aggregated and used by Render. I see this function
// handling author filtering, nil pointers, etc.
func Filter(in chan pipelinePR, out chan SummarizedPullRequest) {
	for {
		v, more := <-in
		if more {
			// TODO add back author processing
			start := *v.PR.CreatedAt
			user := *v.PR.User
			out <- SummarizedPullRequest{
				Owner:    v.Owner,
				Repo:     v.Repo,
				Number:   *v.PR.Number,
				Title:    *v.PR.Title,
				Author:   *user.Login,
				OpenedAt: start,
			}
		} else {
			close(out)
			return
		}
	}
}

// Display formats pull requests onto a html page as they
// come in from the rest of the pipeline.
func Display(in chan SummarizedPullRequest, done chan bool, w io.Writer, now time.Time) {
	fmt.Fprintf(w, "<html><head><meta http-equiv='refresh' content='86400'></head><body style='background: #333; color: #fff; width: 50%; margin: 0 auto;'>")
	fmt.Fprintf(w, "<h1 style='color: #FFF; padding: 0; margin: 0;'>Outstanding Pull Requests</h1><small style='color: #FFF'>last refreshed at %s</small><hr>", now.Format("2006-01-02 15:04:05"))
	for {
		pr, more := <-in
		if more {
			hours := now.Sub(pr.OpenedAt)
			n := hours.Hours() / (240 * time.Hour).Hours()
			if n > 1 {
				n = 1
			}
			stopyellow := (24 * time.Hour).Hours()
			stopred := (72 * time.Hour).Hours()
			color := "#777"
			if hours.Hours() > stopred {
				color = "#FF4500"
			} else if hours.Hours() > stopyellow {
				color = "#FFA500"
			}
			style := fmt.Sprintf(`margin: 3px; padding: 8px; background: linear-gradient( 90deg, %s %d%%, #333 %d%%);`, color, int(n*100), int(n*100))
			fmt.Fprintf(w, "<div style='%s'><b>%s/%s</b> #%d %s by %s @ %d days or %d hours</div>", style, pr.Owner, pr.Repo, pr.Number, pr.Title, pr.Author, hours/(24*time.Hour), hours/time.Hour)
		} else {
			fmt.Fprintf(w, "</body></html>")
			done <- true
			return
		}
	}
}
