package bot

import (
	"strings"
	"testing"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/daenney/gdq/v2"
	"maunium.net/go/mautrix/event"
)

func assertContains(t *testing.T, a string, b string) {
	t.Helper()
	if strings.Contains(a, b) {
		return
	}
	t.Errorf("Received '%s', expected to contain '%s'", a, b)
}

func assertNotContains(t *testing.T, a string, b string) {
	t.Helper()
	if !strings.Contains(a, b) {
		return
	}
	t.Errorf("Received '%s', expected to not contain '%s'", a, b)
}
func TestHTMLMessage(t *testing.T) {
	e := &gdq.Run{
		Title: "first game",
		Runners: gdq.Runners{
			gdq.Runner{Handle: "first runner"},
		},
		Hosts:    []string{"first host"},
		Start:    time.Date(2020, 12, 1, 13, 37, 0, 0, time.UTC),
		Estimate: gdq.Duration{Duration: 20 * time.Minute},
	}
	f := htmlEvent(e)
	assertEqual(t, f, "<b>first game</b> on Tuesday, the 1st of December at 13:37 UTC (2020) run by <i>first runner</i> and hosted by <i>first host</i> lasting 20 minutes")
}

func TestPlainMessage(t *testing.T) {
	e := &gdq.Run{
		Title: "first game",
		Runners: gdq.Runners{
			gdq.Runner{Handle: "first runner"},
		},
		Hosts:    []string{"first host"},
		Start:    time.Date(2020, 12, 1, 13, 37, 0, 0, time.UTC),
		Estimate: gdq.Duration{Duration: 20 * time.Minute},
	}
	f := plainEvent(e)
	assertEqual(t, f, "first game on Tuesday, the 1st of December at 13:37 UTC (2020) run by first runner and hosted by first host lasting 20 minutes")
}

func TestMessageSchedule(t *testing.T) {
	t.Run("no events", func(t *testing.T) {
		s := &gdq.Schedule{}
		msg := msgSchedule(s)
		assertContains(t, msg.Body, "no events")
	})
	t.Run("one event", func(t *testing.T) {
		s := gdq.NewScheduleFrom([]*gdq.Run{{
			Title: "first game",
			Runners: gdq.Runners{
				gdq.Runner{Handle: "first runner"},
			},
			Hosts:    []string{"first host"},
			Start:    time.Date(2020, 12, 1, 13, 37, 0, 0, time.UTC),
			Estimate: gdq.Duration{Duration: 20 * time.Minute},
		}})
		msg := msgSchedule(s)
		assertEqual(t, event.FormatHTML, msg.Format)

		t.Run("plaintext", func(t *testing.T) {
			content := msg.Body
			assertContains(t, content, "one event")
			assertContains(t, content, "your query:\n")
			assertContains(t, content, "* first")
			assertContains(t, content, "20 minutes\n")
		})
		t.Run("HTML", func(t *testing.T) {
			content := msg.FormattedBody
			assertContains(t, content, "one event")
			assertContains(t, content, "your query:<br><ul>")
			assertContains(t, content, "<li><b>first")
			assertContains(t, content, "20 minutes</li></ul>")
		})
	})
	t.Run("multiple events", func(t *testing.T) {
		s := gdq.NewScheduleFrom([]*gdq.Run{
			{
				Title: "first game",
				Runners: gdq.Runners{
					gdq.Runner{Handle: "first runner"},
				},
				Hosts:    []string{"first host"},
				Start:    time.Date(2020, 12, 1, 13, 37, 0, 0, time.UTC),
				Estimate: gdq.Duration{Duration: 20 * time.Minute},
			},
			{
				Title: "second game",
				Runners: gdq.Runners{
					gdq.Runner{Handle: "second runner"},
				},
				Hosts:    []string{"second host"},
				Start:    time.Date(2020, 12, 1, 14, 37, 0, 0, time.UTC),
				Estimate: gdq.Duration{Duration: 1*time.Hour + 5*time.Minute},
			},
		})
		msg := msgSchedule(s)
		assertEqual(t, event.FormatHTML, msg.Format)

		t.Run("plaintext", func(t *testing.T) {
			content := msg.Body
			assertContains(t, content, "multiple events")
			assertContains(t, content, "your query:\n")
			assertContains(t, content, "* first")
			assertContains(t, content, "20 minutes\n")
			assertContains(t, content, "* second")
			assertContains(t, content, "1 hour and 5 minutes\n")
		})
		t.Run("HTML", func(t *testing.T) {
			content := msg.FormattedBody
			assertContains(t, content, "multiple events")
			assertContains(t, content, "your query:<br><ul>")
			assertContains(t, content, "<li><b>first")
			assertContains(t, content, "20 minutes</li>")
			assertContains(t, content, "<li><b>second")
			assertContains(t, content, "1 hour and 5 minutes</li></ul>")

		})
	})
}

func TestHelpMessage(t *testing.T) {
	b := &bot{}
	m, err := b.msgHelp()
	assertEqual(t, err, nil)
	assertContains(t, m.Body, "Supported commands")
}

func TestNextEventMessage(t *testing.T) {
	t.Run("with loaded cache", func(t *testing.T) {
		t.Run("no future events", func(t *testing.T) {
			b := &bot{
				cache: ttlcache.NewCache(),
			}
			b.cache.Set("sched", gdq.NewScheduleFrom([]*gdq.Run{{
				Title: "first game",
				Runners: gdq.Runners{
					gdq.Runner{Handle: "first runner"},
				},
				Hosts:    []string{"first host"},
				Start:    time.Date(1900, 12, 1, 13, 37, 0, 0, time.UTC),
				Estimate: gdq.Duration{Duration: 20 * time.Minute},
			}}))

			msg, err := b.msgScheduleNext()
			assertEqual(t, nil, err)
			assertContains(t, msg.Body, "no further")
		})
		t.Run("one future event", func(t *testing.T) {
			b := &bot{
				cache: ttlcache.NewCache(),
			}
			b.cache.Set("sched", gdq.NewScheduleFrom([]*gdq.Run{{
				Title: "first game",
				Runners: gdq.Runners{
					gdq.Runner{Handle: "first runner"},
				},
				Hosts:    []string{"first host"},
				Start:    time.Date(2100, 12, 1, 13, 37, 0, 0, time.UTC),
				Estimate: gdq.Duration{Duration: 20 * time.Minute},
			}}))

			msg, err := b.msgScheduleNext()
			assertEqual(t, nil, err)
			assertEqual(t, event.FormatHTML, msg.Format)
			assertContains(t, msg.Body, "The next event is: first game")
			assertContains(t, msg.FormattedBody, "The next event is: <b>first game</b>")
		})
		t.Run("multiple future events", func(t *testing.T) {
			b := &bot{
				cache: ttlcache.NewCache(),
			}
			b.cache.Set("sched", gdq.NewScheduleFrom([]*gdq.Run{
				{
					Title: "first game",
					Runners: gdq.Runners{
						gdq.Runner{Handle: "first runner"},
					},
					Hosts:    []string{"first host"},
					Start:    time.Date(2100, 12, 1, 13, 37, 0, 0, time.UTC),
					Estimate: gdq.Duration{Duration: 20 * time.Minute},
				},
				{
					Title: "second game",
					Runners: gdq.Runners{
						gdq.Runner{Handle: "second runner"},
					},
					Hosts:    []string{"second host"},
					Start:    time.Date(2100, 12, 1, 14, 37, 0, 0, time.UTC),
					Estimate: gdq.Duration{Duration: 1*time.Hour + 5*time.Minute},
				},
			}))

			msg, err := b.msgScheduleNext()
			assertEqual(t, nil, err)
			assertEqual(t, event.FormatHTML, msg.Format)
			assertContains(t, msg.Body, "The next event is: first game")
			assertContains(t, msg.FormattedBody, "The next event is: <b>first game</b>")
			assertNotContains(t, msg.Body, "The next event is: second game")
			assertNotContains(t, msg.FormattedBody, "The next event is: <b>second game</b>")
		})
	})
	t.Run("with unavailable cache", func(t *testing.T) {
		b := &bot{
			cache: ttlcache.NewCache(),
		}

		_, err := b.msgScheduleNext()
		assertNotNil(t, err)
	})
}

func TestMessageForRunnerHostEvent(t *testing.T) {
	t.Run("with entries available in cache", func(t *testing.T) {

		b := &bot{
			cache: ttlcache.NewCache(),
		}
		b.cache.Set("sched", gdq.NewScheduleFrom([]*gdq.Run{
			{
				Title: "first game",
				Runners: gdq.Runners{
					gdq.Runner{Handle: "first runner"},
				},
				Hosts:    []string{"first host"},
				Start:    time.Date(2100, 12, 1, 13, 37, 0, 0, time.UTC),
				Estimate: gdq.Duration{Duration: 20 * time.Minute},
			},
			{
				Title: "second game",
				Runners: gdq.Runners{
					gdq.Runner{Handle: "second runner"},
				},
				Hosts:    []string{"second host"},
				Start:    time.Date(2100, 12, 1, 14, 37, 0, 0, time.UTC),
				Estimate: gdq.Duration{Duration: 1*time.Hour + 5*time.Minute},
			},
			{
				Title: "third game",
				Runners: gdq.Runners{
					gdq.Runner{Handle: "second runner"},
				},
				Hosts:    []string{"second host"},
				Start:    time.Date(2100, 12, 1, 14, 37, 0, 0, time.UTC),
				Estimate: gdq.Duration{Duration: 1*time.Hour + 5*time.Minute},
			},
		}))

		var tests = []struct {
			name         string
			f            filteredHandler
			filter       string
			bodyMust     []string
			bodyMustNot  []string
			fbodyMust    []string
			fbodyMustNot []string
			fbodyEmpty   bool
		}{
			{
				name: "no matching game", f: b.msgScheduleForEvent, filter: "x",
				bodyMust:    []string{"are no events"},
				bodyMustNot: []string{"on", "run by", "hosted by", "lasting"},
				fbodyEmpty:  true,
			},
			{
				name: "no matching host", f: b.msgScheduleForHost, filter: "x",
				bodyMust:    []string{"are no events"},
				bodyMustNot: []string{"on", "run by", "hosteby by", "lasting"},
				fbodyEmpty:  true,
			},
			{
				name: "no matching runner", f: b.msgScheduleForRunner, filter: "x",
				bodyMust:    []string{"are no events"},
				bodyMustNot: []string{"on", "run by", "hosted by", "lasting"},
				fbodyEmpty:  true,
			},
			{
				name: "single matching game", f: b.msgScheduleForEvent, filter: "fir",
				bodyMust:     []string{"one event", "* first game", "first runner", "first host"},
				bodyMustNot:  []string{"* second", "second runner", "second host", "* third", "third runner", "third host"},
				fbodyMust:    []string{"one event", "<li><b>first game</b>", "<i>first runner</i>", "<i>first host</i>"},
				fbodyMustNot: []string{"multiple events:<br>", "<li><b>second game</b>", "<i>second runnner</i>", "<i>second host</i>", "<li><b>third game</b>", "<i>third runnner</i>", "<i>third host</i>"},
			},
			{
				name: "single matching host", f: b.msgScheduleForHost, filter: "first h",
				bodyMust:     []string{"one event", "* first game", "first runner", "first host"},
				bodyMustNot:  []string{"multiple events", "* second", "second runner", "second host", "* third", "third runner", "third host"},
				fbodyMust:    []string{"one event", "<li><b>first game</b>", "<i>first runner</i>", "<i>first host</i>"},
				fbodyMustNot: []string{"multiple events:<br>", "<li><b>second game</b>", "<i>second runnner</i>", "<i>second host</i>", "<li><b>third game</b>", "<i>third runnner</i>", "<i>third host</i>"},
			},
			{
				name: "single matching runner", f: b.msgScheduleForRunner, filter: "first r",
				bodyMust:     []string{"one event", "* first game", "first runner", "first host"},
				bodyMustNot:  []string{"multiple events", "* second", "second runner", "second host", "* third", "third runner", "third host"},
				fbodyMust:    []string{"one event", "<li><b>first game</b>", "<i>first runner</i>", "<i>first host</i>"},
				fbodyMustNot: []string{"multiple events:<br>", "<li><b>second game</b>", "<i>second runnner</i>", "<i>second host</i>", "<li><b>third game</b>", "<i>third runnner</i>", "<i>third host</i>"},
			},
			// here here
			{
				name: "multiple matching games", f: b.msgScheduleForEvent, filter: "game",
				bodyMust:     []string{"multiple event", "* first game", "* second game", "* third game"},
				bodyMustNot:  []string{"no events"},
				fbodyMust:    []string{"multiple events", "<li><b>first game</b>", "<li><b>second game</b>", "<li><b>third game</b>"},
				fbodyMustNot: []string{"no events"},
			},
			{
				name: "multiple matching host", f: b.msgScheduleForHost, filter: "second",
				bodyMust:     []string{"multiple events", "second host", "second game", "third game"},
				bodyMustNot:  []string{"no events"},
				fbodyMust:    []string{"multiple events", "<i>second host</i>", "<li><b>second game</b>", "<li><b>third game</b>"},
				fbodyMustNot: []string{"no events"},
			},
			{
				name: "multiple matching runner", f: b.msgScheduleForRunner, filter: "second",
				bodyMust:     []string{"multiple events", "second host", "second game", "third game"},
				bodyMustNot:  []string{"no events"},
				fbodyMust:    []string{"multiple events", "<i>second host</i>", "<li><b>second game</b>", "<li><b>third game</b>"},
				fbodyMustNot: []string{"no events"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				msg, err := tt.f(tt.filter)
				assertEqual(t, nil, err)
				for _, m := range tt.bodyMust {
					assertContains(t, msg.Body, m)
				}
				for _, m := range tt.bodyMustNot {
					assertNotContains(t, msg.Body, m)
				}
				for _, m := range tt.fbodyMust {
					assertContains(t, msg.FormattedBody, m)
				}
				for _, m := range tt.fbodyMustNot {
					assertNotContains(t, msg.FormattedBody, m)
				}
				if tt.fbodyEmpty {
					assertEqual(t, msg.FormattedBody, "")
				}
			})
		}
	})
	t.Run("with unavailable cache", func(t *testing.T) {
		b := &bot{
			cache: ttlcache.NewCache(),
		}
		var tests = []struct {
			name   string
			f      filteredHandler
			filter string
		}{
			{name: "forEvent", f: b.msgScheduleForEvent, filter: "r"},
			{name: "forHost", f: b.msgScheduleForHost, filter: "r"},
			{name: "forRunner", f: b.msgScheduleForRunner, filter: "r"},
		}
		for _, tt := range tests {
			_, err := tt.f(tt.filter)
			assertNotNil(t, err)
		}
	})
}

func TestAnnounceEventMessage(t *testing.T) {
	b := &bot{
		cache: ttlcache.NewCache(),
	}

	msg := b.msgAnnounce(&gdq.Run{
		Title: "first game",
		Runners: gdq.Runners{
			gdq.Runner{Handle: "first runner"},
		},
		Hosts:    []string{"first host"},
		Start:    time.Date(1900, 12, 1, 13, 37, 0, 0, time.UTC),
		Estimate: gdq.Duration{Duration: 20 * time.Minute},
	})
	assertContains(t, msg.Body, "is starting in approximately")
}
