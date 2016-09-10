package prmonitor

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

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