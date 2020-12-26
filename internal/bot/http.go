package bot

import (
	"fmt"
	"net/http"
	"net/url"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

const userAgent = "gdqbot (+https://github.com/daenney/gdq)"

type transport struct{}

func (*transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", userAgent)
	return http.DefaultTransport.RoundTrip(req)
}

var safeClient = &http.Client{Transport: &transport{}}

func newMatrixClient(homeserverURL string, userID id.UserID, accessToken string) (*mautrix.Client, error) {
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
		Client:        safeClient,
		Prefix:        mautrix.URLPath{"_matrix", "client", "r0"},
		Syncer:        mautrix.NewDefaultSyncer(),
	}
	store := mautrix.NewAccountDataStore(eventID, c)
	c.Store = store

	return c, nil
}
