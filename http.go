package main

import (
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

const userAgent = "gdqbot (+https://github.com/daenney/gdq)"

var defaultTrasnport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   runtime.NumCPU() + 1,
}

type transport struct{}

func (*transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", userAgent)
	return defaultTrasnport.RoundTrip(req)
}

var safeClient = &http.Client{Transport: &transport{}}

func newMatrixClient(homeserverURL string, userID id.UserID, accessToken string) (*mautrix.Client, error) {
	hsURL, err := url.Parse(homeserverURL)
	if err != nil {
		return nil, err
	}
	if hsURL.Scheme == "" {
		hsURL.Scheme = "https"
	}
	return &mautrix.Client{
		AccessToken:   accessToken,
		HomeserverURL: hsURL,
		UserID:        userID,
		Client:        safeClient,
		Prefix:        mautrix.URLPath{"_matrix", "client", "r0"},
		Syncer:        mautrix.NewDefaultSyncer(),
		Store:         mautrix.NewInMemoryStore(),
	}, nil
}
