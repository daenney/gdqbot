package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetUserAgent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header["User-Agent"][0]; h != userAgent {
			t.Errorf("Expected a User-Agent header of: %s, got: %s", userAgent, h)
		}
		fmt.Fprintln(w, "")
	}))
	defer ts.Close()

	_, _ = safeClient.Get(ts.URL)
}
