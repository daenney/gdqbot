package bot

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type transport struct {
	userAgent string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.userAgent)
	return http.DefaultTransport.RoundTrip(req)
}

func newMatrixClient(userAgent string, homeserverURL string, userID id.UserID, accessToken string) (*mautrix.Client, error) {
	if homeserverURL == "" {
		return nil, fmt.Errorf("received empty homeserver URL")
	}
	hsURL, err := url.Parse(homeserverURL)
	if err != nil {
		return nil, err
	}
	if hsURL.Scheme == "" {
		hsURL.Scheme = "https"
	}
	c := &mautrix.Client{
		AccessToken:   accessToken,
		HomeserverURL: hsURL,
		UserID:        userID,
		Client: &http.Client{
			Transport: &transport{userAgent: userAgent},
			Timeout:   60 * time.Second,
		},
		UserAgent: userAgent,
		Prefix:    mautrix.URLPath{"_matrix", "client", "r0"},
		Syncer:    mautrix.NewDefaultSyncer(),
	}
	store := mautrix.NewAccountDataStore(eventID, c)
	c.Store = store

	return c, nil
}
