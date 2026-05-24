package httphandler

import (
	"net/http"
	"net/url"
	"strings"
)

// NormalizePath trims a trailing slash so /v1/foo/ matches routes registered as /v1/foo.
func NormalizePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			r2 := *r
			r2.URL = cloneURL(r.URL)
			r2.URL.Path = strings.TrimSuffix(path, "/")
			next.ServeHTTP(w, &r2)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return &url.URL{}
	}
	u2 := *u
	return &u2
}
