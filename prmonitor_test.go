package prmonitor

import (
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
// can be visually inspected.
func TestRender(t *testing.T) {
	now := time.Now()
	prs := []SummarizedPullRequest{
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   4,
			Title:    "test pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-5 * time.Hour),
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   5,
			Title:    "yellow zone pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-25 * time.Hour),
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   6,
			Title:    "red zone pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-73 * time.Hour),
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   7,
			Title:    "boundary value pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-1000 * time.Hour),
		},
	}
	f, err := os.Create("tmp.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	c := make(chan SummarizedPullRequest)
	d := make(chan bool)
	go Display(c, d, f, now)
	for _, pr := range prs {
		c <- pr
	}
	close(c)
	<-d
}

// Dashboard Tests
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
	req.Header.Set("X-Timestamp", "2016-09-16T15:04:05Z")
	Dashboard(
		Config{
			Repos: []Repo{
				{
					Owner: "Docker",
					Repo:  "swarmkit",
					Depth: 15,
				},
			},
		},
		github.NewClient(
			h,
		),
	)(w, req)
	if w.Body.String() != "<html><head><meta http-equiv='refresh' content='86400'></head><body style='background: #333; color: #fff; width: 50%!;(MISSING) margin: 0 auto;'><h1 style='color: #FFF; padding: 0; margin: 0;'>Outstanding Pull Requests</h1><small style='color: #FFF'>last refreshed at 2016-09-16 15:04:05</small><hr><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #777 6%, #333 6%);'><b>Docker/swarmkit</b> #1544 Deliver secrets using assignments by cyli @ 0 days or 16 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #777 6%, #333 6%);'><b>Docker/swarmkit</b> #1543 raft: defer closing conns on applyRemoveNode by LK4D4 @ 0 days or 16 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FFA500 15%, #333 15%);'><b>Docker/swarmkit</b> #1540 Improve error message for removing pre-defined (e.g., `ingress`) network by yongtang @ 1 days or 36 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FFA500 22%, #333 22%);'><b>Docker/swarmkit</b> #1536 add portallocator test by allencloud @ 2 days or 53 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FFA500 28%, #333 28%);'><b>Docker/swarmkit</b> #1530 Make it even more clear where to file bugs by dperny @ 2 days or 69 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 70%, #333 70%);'><b>Docker/swarmkit</b> #1517 Re-enable plugin filter by nishanttotla @ 7 days or 168 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 85%, #333 85%);'><b>Docker/swarmkit</b> #1512 [WIP] Topology-aware scheduling by aaronlehmann @ 8 days or 204 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 86%, #333 86%);'><b>Docker/swarmkit</b> #1511 Secret-specific protos by cyli @ 8 days or 207 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 95%, #333 95%);'><b>Docker/swarmkit</b> #1500 Check for duplicate mount points in ServiceSpec by kunalkushwaha @ 9 days or 228 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 100%, #333 100%);'><b>Docker/swarmkit</b> #1488 judge *api.Task nil and add nodeinfo test by allencloud @ 14 days or 341 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 100%, #333 100%);'><b>Docker/swarmkit</b> #1483 raft: Migrate WAL files to new directory by aaronlehmann @ 14 days or 357 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 100%, #333 100%);'><b>Docker/swarmkit</b> #1476 Change Task Name to `<ServiceAnnotations.Name>.<NodeID>.<TaskID>` in global mode by yongtang @ 16 days or 398 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 100%, #333 100%);'><b>Docker/swarmkit</b> #1473 design: add topology-aware scheduling proposal by aaronlehmann @ 16 days or 404 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 100%, #333 100%);'><b>Docker/swarmkit</b> #1446 Implement HA scheduling by aaronlehmann @ 21 days or 516 hours</div><div style='margin: 3px; padding: 8px; background: linear-gradient( 90deg, #FF4500 100%, #333 100%);'><b>Docker/swarmkit</b> #1433 return all validation error in spec validation by allencloud @ 22 days or 550 hours</div></body></html>" {
		t.Logf("%s", w.Body.String())
		t.Fail()
	}
}
