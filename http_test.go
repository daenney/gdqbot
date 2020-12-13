package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"maunium.net/go/mautrix/id"
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

func TestNewMatrixClient(t *testing.T) {
	t.Run("empty homeserver URL", func(t *testing.T) {
		_, err := newMatrixClient("", "", "")
		assertNotNil(t, err)
	})
	t.Run("wonky homeserver URL", func(t *testing.T) {
		_, err := newMatrixClient("\x00", "", "")
		assertNotNil(t, err)
	})
	t.Run("homeserver URL without protocol", func(t *testing.T) {
		c, err := newMatrixClient("example.com", "a", "b")
		assertEqual(t, nil, err)
		assertEqual(t, "https://example.com", c.HomeserverURL.String())
		assertEqual(t, id.UserID("a"), c.UserID)
		assertEqual(t, "b", c.AccessToken)
	})
}
