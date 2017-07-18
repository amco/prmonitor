package prmonitor

import (
	"fmt"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/go-github/github"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Basic Auth Tests
func TestBasicAuthFailure(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Basic 490hf")

	BasicAuth("user", "pass", func(w http.ResponseWriter, r *http.Request) {

	})(w, r)

	if w.Code != 401 {
		t.Logf("ERROR: http code '%d' expected, but got '%d'", 401, w.Code)
		t.Fail()
		return
	}

	if w.Header().Get("WWW-Authenticate") != "Basic" {
		t.Logf("ERROR: expected WWW-Authenticate header")
		t.Fail()
		return
	}
}

func TestBasicAuthSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	BasicAuth("user", "pass", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299)
		return
	})(w, r)

	if w.Code != 299 {
		t.Logf("ERROR: http code '%d' expected, but got '%d'", 299, w.Code)
		t.Fail()
		return
	}

	if w.Header().Get("WWW-Authenticate") != "" {
		t.Logf("ERROR: unexpected WWW-Authenticate header on successful auth")
		t.Fail()
		return
	}
}

func TestBasicAuthSuccess2(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Basic Zm9vOmJhcg==")

	BasicAuth("foo", "bar", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299)
		return
	})(w, r)

	if w.Code != 299 {
		t.Logf("ERROR: http code '%d' expected, but got '%d'", 299, w.Code)
		t.Fail()
		return
	}

	if w.Header().Get("WWW-Authenticate") != "" {
		t.Logf("ERROR: unexpected WWW-Authenticate header on successful auth")
		t.Fail()
		return
	}
}

// SSL Required Tests
func TestSSLRequiredRedirects(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.org/unsecure", nil)
	r.Header.Set("X-Forwarded-Proto", "http")

	SSLRequired("https://example.org/secure", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299)
		return
	})(w, r)

	if w.Code != 301 {
		t.Logf("ERROR: http code '%d' expected, but got '%d'", 301, w.Code)
		t.Fail()
		return
	}

	if w.Header().Get("Location") != "https://example.org/secure" {
		t.Logf("ERROR: location '%s' expected, but got '%s'", "https://example.org/secure", w.Header().Get("Location"))
		t.Fail()
		return
	}
}

func TestSSLRequiredRedirects2(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://other.example.org/1", nil)
	r.Header.Set("X-Forwarded-Proto", "http")

	SSLRequired("https://other.example.org/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299)
		return
	})(w, r)

	if w.Code != 301 {
		t.Logf("ERROR: http code '%d' expected, but got '%d'", 301, w.Code)
		t.Fail()
		return
	}

	if w.Header().Get("Location") != "https://other.example.org/1" {
		t.Logf("ERROR: location '%s' expected, but got '%s'", "https://other.example.org/1", w.Header().Get("Location"))
		t.Fail()
		return
	}
}

func TestSSLNotRedirectedIfAlreadyHTTPS(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://other.example.org/1", nil)
	r.Header.Set("X-Forwarded-Proto", "https")

	SSLRequired("http://other.example.org/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299)
		return
	})(w, r)

	if w.Code != 299 {
		t.Logf("ERROR: http code '%d' expected, but got '%d'", 299, w.Code)
		t.Fail()
		return
	}

	if w.Header().Get("Location") != "" {
		t.Logf("ERROR: no location expected but got '%s'", w.Header().Get("Location"))
		t.Fail()
		return
	}
}

// Timestamp Test
func TestTimestamp(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://other.example.org/1", nil)
	r.Header.Set("X-Forwarded-Proto", "https")

	basetime := time.Now()
	Timestamp(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299)
		return
	})(w, r)

	if w.Code != 299 {
		t.Logf("expected code %d, got %d", 299, w.Code)
		t.Fail()
	}

	reqtime, err := time.Parse(time.RFC3339, r.Header.Get("X-Timestamp"))
	if err != nil {
		t.Logf("couldn't parse time %s, got error: %s", w.Header().Get("X-Timestamp"), err)
		t.Fail()
	}

	if reqtime.Sub(basetime) > 1*time.Second {
		t.Logf("timestamps differ too much")
		t.Fail()
	}
}

// Render Tests - attempt to render the page to a tmp file so it
// can be visually inspected using hand-crafted data
func TestRender(t *testing.T) {
	now := time.Now()
	prs := []SummarizedPullRequest{
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   2,
			Title:    "closed pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-72 * time.Hour),
			ClosedAt: now.Add(-24 * time.Hour),
			State:    "closed",
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   4,
			Title:    "test pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-1 * time.Hour),
			ClosedAt: now,
			State:    "open",
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   5,
			Title:    "yellow zone pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-25 * time.Hour),
			ClosedAt: now,
			State:    "open",
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   6,
			Title:    "red zone pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-73 * time.Hour),
			ClosedAt: now,
			State:    "open",
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   7,
			Title:    "boundary value pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-1000 * time.Hour),
			ClosedAt: now,
			State:    "open",
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   8,
			Title:    "green zone pr closed days ago",
			Author:   "brentdrich",
			OpenedAt: now.Add(-72 * time.Hour),
			ClosedAt: now.Add(-64 * time.Hour),
			State:    "closed",
		},
	}
	f, err := os.Create("tmp.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	c := make(chan SummarizedPullRequest)
	d := Display(c, f, now, SortBy("date"), Config{Customization: GetCustomizations()})
	for _, pr := range prs {
		c <- pr
	}
	close(c)
	<-d
}

// Make sure that getColor is comparing time properly
func TestGetColor(t *testing.T) {
	now := time.Now()
	onePr := SummarizedPullRequest{
		Owner:    "brentdrich",
		Repo:     "prmonitor",
		Number:   2,
		Title:    "closed pr",
		Author:   "brentdrich",
		OpenedAt: now.Add(-72 * time.Hour),
		ClosedAt: now.Add(-24 * time.Hour),
	}
	openedFor := now.Sub(onePr.OpenedAt).Hours()
	res := getColor(Config{Customization: GetCustomizations()}, openedFor, "opened")
	if res != "#cc0000" {
		fmt.Sprintf("Expected to get #cc0000, but got %s", res)
		return
	}

	twoPr := SummarizedPullRequest{
		Owner:    "brentdrich",
		Repo:     "prmonitor",
		Number:   2,
		Title:    "closed pr",
		Author:   "brentdrich",
		OpenedAt: now.Add(-10 * time.Hour),
		ClosedAt: now.Add(-24 * time.Hour),
	}

	openedFor = now.Sub(twoPr.OpenedAt).Hours()
	res = getColor(Config{Customization: GetCustomizations()}, openedFor, "opened")
	if res != "#00cc66" {
		fmt.Sprintf("Expected to get #00cc66, but got %s", res)
		return
	}

	// test that gray is returned when pr is closed
	res = getColor(Config{Customization: GetCustomizations()}, openedFor, "closed")
	if res != "#00cc66" {
		fmt.Sprintf("Expected to get #00cc66, but got %s", res)
		return
	}
}

// Dashboard Tests - outputs a rendered dashboard from cached or real data.
func TestDashboard(t *testing.T) {
	r, err := recorder.New("github")
	if err != nil {
		panic(err)
	}
	defer r.Stop()
	h := &http.Client{
		Timeout:   1 * time.Second,
		Transport: r.Transport,
	}
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-Timestamp", time.Now().Format(time.RFC3339))
	Dashboard(
		Config{
			Repos: []Repo{
				{
					Owner: "Docker",
					Repo:  "swarmkit",
					Depth: 15,
				},
			},
			Authors: &[]string{
				"aaronlehmann",
				"LK4D4",
			},
		},
		github.NewClient(
			h,
		),
	)(w, req)
	f, err := os.Create("e2e.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprint(f, w.Body.String())
}
