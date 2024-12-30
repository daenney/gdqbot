package bot

import (
	"testing"

	"maunium.net/go/mautrix/id"
)

func TestNewMatrixClient(t *testing.T) {
	t.Run("empty homeserver URL", func(t *testing.T) {
		_, err := newMatrixClient("", "", "", "")
		assertNotNil(t, err)
	})
	t.Run("wonky homeserver URL", func(t *testing.T) {
		_, err := newMatrixClient("", "\x00", "", "")
		assertNotNil(t, err)
	})
	t.Run("homeserver URL without protocol", func(t *testing.T) {
		c, err := newMatrixClient("", "example.com", "a", "b")
		assertEqual(t, nil, err)
		assertEqual(t, "https://example.com", c.HomeserverURL.String())
		assertEqual(t, id.UserID("a"), c.UserID)
		assertEqual(t, "b", c.AccessToken)
	})
}
