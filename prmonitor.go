package prmonitor

import (
	"encoding/base64"
	"fmt"
	"github.com/google/go-github/github"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Data Structures

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

	// the time the PR was closed (or the current time).
	ClosedAt time.Time
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

	// optional list of authors - if included, will only display open PRs
	// by those authors. Useful for filtering large codebases by team.
	Authors *[]string
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

		opened := make(chan Repo)
		closed := make(chan Repo)

		// construct pipeline
		done := Display(
			FilterByAuthor(
				FilterByDate(
					merge(
						Retrieve(opened, client, now, "open", "created"),
						Retrieve(closed, client, now, "closed", "updated"),
					), now),
				t.Authors),
			w, now)

		for _, repo := range t.Repos {
			opened <- repo
			closed <- repo
		}
		close(opened)
		close(closed)

		<-done
	}
}

// Data processing pipeline...

// Transform converts a github pull request into a pointer-free summary
// that can be used by the rest of the pipeline.
func Transform(v *github.PullRequest, now time.Time) (SummarizedPullRequest, error) {
	closedAt := now
	if v.ClosedAt != nil {
		closedAt = *v.ClosedAt
	}
	return SummarizedPullRequest{
		Owner:    *v.Base.Repo.Owner.Login,
		Repo:     *v.Base.Repo.Name,
		Number:   *v.Number,
		Title:    *v.Title,
		Author:   *v.User.Login,
		OpenedAt: *v.CreatedAt,
		ClosedAt: closedAt,
	}, nil
}

func merge(cs ...<-chan SummarizedPullRequest) <-chan SummarizedPullRequest {
	var wg sync.WaitGroup
	out := make(chan SummarizedPullRequest)
	output := func(c <-chan SummarizedPullRequest) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Retrieve pulls in a repository and fetches pull requests that
// are passed to the next stage in the pipeline.
func Retrieve(in chan Repo, client *github.Client, now time.Time, state string, sort string) <-chan SummarizedPullRequest {
	out := make(chan SummarizedPullRequest)
	go func() {
		for r := range in {
			op := &github.PullRequestListOptions{}
			op.State = state
			op.Sort = sort
			op.Direction = "desc"
			op.Page = 0
			op.Base = "master"
			op.PerPage = r.Depth
			oprs, _, err := client.PullRequests.List(r.Owner, r.Repo, op)
			if err != nil {
				return
			}
			for _, v := range oprs {
				if p, err := Transform(v, now); err == nil {
					out <- p
				}
			}
		}
		close(out)
	}()
	return out
}

// FilterByDate drops Summarized Pull Requests that were closed more
// than 10 days ago.
func FilterByDate(in <-chan SummarizedPullRequest, now time.Time) <-chan SummarizedPullRequest {
	out := make(chan SummarizedPullRequest)
	go func() {
		for v := range in {
			if now.Sub(v.ClosedAt) < 240*time.Hour {
				out <- v
			}
		}
		close(out)

	}()
	return out
}

// FilterByAuthor drops SummarizedPullRequests that don't belong to
// a team member (provided the array exists)
func FilterByAuthor(in <-chan SummarizedPullRequest, authors *[]string) <-chan SummarizedPullRequest {
	out := make(chan SummarizedPullRequest)
	go func() {
		for v := range in {
			if authors != nil {
				for _, a := range *authors {
					if a == v.Author {
						out <- v
						continue
					}
				}
			} else {
				out <- v
			}

		}
		close(out)
	}()
	return out
}

// Display formats pull requests onto a html page as they
// come in from the rest of the pipeline.
func Display(in <-chan SummarizedPullRequest, w io.Writer, now time.Time) <-chan bool {
	out := make(chan bool)
	go func() {
		fmt.Fprintf(w, "<html><head><meta http-equiv='refresh' content='86400'></head><body style='background: #333; color: #fff; width: 50%; margin: 0 auto;'>")
		fmt.Fprintf(w, "<h1 style='color: #FFF; padding: 0; margin: 0;'>Outstanding Pull Requests</h1><small style='color: #FFF'>last refreshed at %s</small><hr>", now.Format("2006-01-02 15:04:05"))
		total := (240 * time.Hour).Hours()
		var prs []SummarizedPullRequest
		for pr := range in {
			prs = append(prs, pr)
		}
		sort.Sort(ByDate(prs))
		for _, pr := range prs {
			start := (total - now.Sub(pr.OpenedAt).Hours()) / total
			end := (total - now.Sub(pr.ClosedAt).Hours()) / total
			color := "#00cc66"
			style := fmt.Sprintf(`margin: 2px; background: linear-gradient( 90deg, #333 0%%, #333 %.6f%%, %s %.6f%%, %s %.6f%%, #333 %.6f%%);`, start*100, color, start*100, color, end*100, end*100)
			fmt.Fprintf(w, "<div style='%s'><b>%s/%s</b> #%d %s by %s</div>", style, pr.Owner, pr.Repo, pr.Number, pr.Title, pr.Author)
		}
		fmt.Fprintf(w, "</body></html>")
		out <- true
		close(out)
	}()
	return out
}

// ByDate sorts summarized pull requests by date.
type ByDate []SummarizedPullRequest

// Len returns length of the ByDate array
func (a ByDate) Len() int { return len(a) }

// Swap exchanges elements in the ByDate array
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less compares two pull request indices
func (a ByDate) Less(i, j int) bool {
	if a[j].ClosedAt.Equal(a[i].ClosedAt) {
		return a[j].OpenedAt.Before(a[i].OpenedAt)
	}
	return a[j].ClosedAt.Before(a[i].ClosedAt)
}
