package prmonitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
