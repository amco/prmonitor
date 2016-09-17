package prmonitor

import (
	"fmt"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/go-github/github"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"testing/quick"
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
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   4,
			Title:    "test pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-1 * time.Hour),
			ClosedAt: now,
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   5,
			Title:    "yellow zone pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-25 * time.Hour),
			ClosedAt: now,
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   6,
			Title:    "red zone pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-73 * time.Hour),
			ClosedAt: now,
		},
		{
			Owner:    "brentdrich",
			Repo:     "prmonitor",
			Number:   7,
			Title:    "boundary value pr",
			Author:   "brentdrich",
			OpenedAt: now.Add(-1000 * time.Hour),
			ClosedAt: now,
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

// Filter Tests - github PRs have lots and lots of pointers, which
// makes unchecked nil pointer very likely. This test feeds in a bunch
// of randomized PRs to shake out any unchecked nil pointers.
func TestFilter_DetectNilPanics(t *testing.T) {
	if err := quick.Check(func(pr *github.PullRequest) bool {
		in := make(chan pipelinePR)
		out := make(chan SummarizedPullRequest)
		go func() {
			// catch nil panics and dump PRs
			defer func() {
				if r := recover(); r != nil {
					close(out)
					t.Logf("recovered %s: %+v", r, pr)
					t.Fail()
				}
			}()
			Filter(in, out, time.Now())
		}()
		in <- pipelinePR{
			Owner: "brentdrich",
			Repo:  "prmonitor",
			PR:    pr,
		}
		close(in)
		<-out
		return true
	}, &quick.Config{MaxCount: 500, Values: prValues}); err != nil {
		t.Logf("err: %s", err.Error())
		t.Fail()
	}
}

// --Quick Check Hack--
// This is actually kind of interesting. When generating large structs, quickcheck
// allocates a "size" among all the struct fields. The size is hardcoded to 50 in
// the testing/quick library. For large structs like a github pull request, the
// size isn't great enough to actually set any fields! You just end up with nil
// structs.
//
// The code below copies a bunch of the quick check code into a Value function
// and uses an arbitrary size value in order to generate rich structs.
//
// NOTE: the code below probably falls under https://golang.org/LICENSE. My plan
// is to break this out into another repository at some point.

func prValues(args []reflect.Value, rand *rand.Rand) {
	for j := 0; j < len(args); j++ {
		var ok bool
		args[j], ok = sizedValue(reflect.TypeOf(&github.PullRequest{}), rand, 500)
		if !ok {
			panic("not okay")
		}
	}
}

func sizedValue(t reflect.Type, rand *rand.Rand, size int) (value reflect.Value, ok bool) {
	if m, ok := reflect.Zero(t).Interface().(quick.Generator); ok {
		return m.Generate(rand, size), true
	}

	v := reflect.New(t).Elem()
	switch concrete := t; concrete.Kind() {
	case reflect.Bool:
		v.SetBool(rand.Int()&1 == 0)
	case reflect.Float32:
		v.SetFloat(float64(randFloat32(rand)))
	case reflect.Float64:
		v.SetFloat(randFloat64(rand))
	case reflect.Complex64:
		v.SetComplex(complex(float64(randFloat32(rand)), float64(randFloat32(rand))))
	case reflect.Complex128:
		v.SetComplex(complex(randFloat64(rand), randFloat64(rand)))
	case reflect.Int16:
		v.SetInt(randInt64(rand))
	case reflect.Int32:
		v.SetInt(randInt64(rand))
	case reflect.Int64:
		v.SetInt(randInt64(rand))
	case reflect.Int8:
		v.SetInt(randInt64(rand))
	case reflect.Int:
		v.SetInt(randInt64(rand))
	case reflect.Uint16:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Uint32:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Uint64:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Uint8:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Uint:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Uintptr:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Map:
		numElems := rand.Intn(size)
		v.Set(reflect.MakeMap(concrete))
		for i := 0; i < numElems; i++ {
			key, ok1 := sizedValue(concrete.Key(), rand, size)
			value, ok2 := sizedValue(concrete.Elem(), rand, size)
			if !ok1 || !ok2 {
				return reflect.Value{}, false
			}
			v.SetMapIndex(key, value)
		}
	case reflect.Ptr:
		if rand.Intn(size) == 0 {
			v.Set(reflect.Zero(concrete)) // Generate nil pointer.
		} else {
			elem, ok := sizedValue(concrete.Elem(), rand, size)
			if !ok {
				return reflect.Value{}, false
			}
			v.Set(reflect.New(concrete.Elem()))
			v.Elem().Set(elem)
		}
	case reflect.Slice:
		numElems := rand.Intn(size)
		sizeLeft := size - numElems
		v.Set(reflect.MakeSlice(concrete, numElems, numElems))
		for i := 0; i < numElems; i++ {
			elem, ok := sizedValue(concrete.Elem(), rand, sizeLeft)
			if !ok {
				return reflect.Value{}, false
			}
			v.Index(i).Set(elem)
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			elem, ok := sizedValue(concrete.Elem(), rand, size)
			if !ok {
				return reflect.Value{}, false
			}
			v.Index(i).Set(elem)
		}
	case reflect.String:
		numChars := rand.Intn(size)
		codePoints := make([]rune, numChars)
		for i := 0; i < numChars; i++ {
			codePoints[i] = rune(rand.Intn(0x10ffff))
		}
		v.SetString(string(codePoints))
	case reflect.Struct:
		n := v.NumField()
		// Divide sizeLeft evenly among the struct fields.
		sizeLeft := size
		if n > sizeLeft {
			sizeLeft = 1
		} else if n > 0 {
			sizeLeft /= n
		}
		for i := 0; i < n; i++ {
			elem, ok := sizedValue(concrete.Field(i).Type, rand, sizeLeft)
			if !ok {
				return reflect.Value{}, false
			}
			if v.Field(i).CanSet() {
				v.Field(i).Set(elem)
			}
		}
	default:
		return reflect.Value{}, false
	}

	return v, true
}

// randFloat32 generates a random float taking the full range of a float32.
func randFloat32(rand *rand.Rand) float32 {
	f := rand.Float64() * math.MaxFloat32
	if rand.Int()&1 == 1 {
		f = -f
	}
	return float32(f)
}

// randFloat64 generates a random float taking the full range of a float64.
func randFloat64(rand *rand.Rand) float64 {
	f := rand.Float64() * math.MaxFloat64
	if rand.Int()&1 == 1 {
		f = -f
	}
	return f
}

// randInt64 returns a random integer taking half the range of an int64.
func randInt64(rand *rand.Rand) int64 { return rand.Int63() - 1<<62 }
