package bot

import (
	"reflect"
	"testing"
	"time"

	"github.com/daenney/gdq/v2"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a == b {
		return
	}
	t.Errorf("Received '%v' (type %v), expected '%v' (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func assertNotNil(t *testing.T, a interface{}) {
	t.Helper()
	if a != nil {
		return
	}
	t.Errorf("Received '%v' (type %v), nil", a, reflect.TypeOf(a))
}

func TestFormatMetadata(t *testing.T) {
	var tests = []struct {
		name     string
		event    *gdq.Run
		runners  string
		hosts    string
		estimate string
	}{
		{name: "empty metadata", event: &gdq.Run{}, runners: "unknown", hosts: "unknown", estimate: "unknown amount of time"},
		{name: "complete metadata",
			event: &gdq.Run{
				Runners: gdq.Runners{
					gdq.Runner{Handle: "runner 1"},
					gdq.Runner{Handle: "runner 2"},
					gdq.Runner{Handle: "runner 3"},
				},
				Hosts:    []string{"host 1", "host 2", "host 3"},
				Estimate: gdq.Duration{Duration: 2*time.Hour + 5*time.Minute},
			},
			runners:  "runner 1, runner 2 and runner 3",
			hosts:    "host 1, host 2 and host 3",
			estimate: "2 hours and 5 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runners, hosts, estimate := formatMetadata(tt.event)
			assertEqual(t, runners, tt.runners)
			assertEqual(t, hosts, tt.hosts)
			assertEqual(t, estimate, tt.estimate)
		})
	}

}

func TestFormatHandles(t *testing.T) {
	var tests = []struct {
		name    string
		handles []string
		res     string
	}{
		{name: "no handles", handles: []string{}, res: "unknown"},
		{name: "one handle", handles: []string{"runner 1"}, res: "runner 1"},
		{name: "two handles", handles: []string{"runner 1", "runner 2"}, res: "runner 1 and runner 2"},
		{name: "more than two handles", handles: []string{"runner 1", "runner 2", "runner 3"}, res: "runner 1, runner 2 and runner 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := formatHandles(tt.handles)
			assertEqual(t, res, tt.res)
		})
	}
}

func TestFormatDate(t *testing.T) {
	var tests = []struct {
		name string
		time time.Time
		res  string
	}{
		{name: "1st", time: time.Date(2020, 12, 1, 13, 37, 0, 0, time.UTC), res: "Tuesday, the 1st of December at 13:37 UTC (2020)"},
		{name: "2nd", time: time.Date(2020, 12, 2, 13, 37, 0, 0, time.UTC), res: "Wednesday, the 2nd of December at 13:37 UTC (2020)"},
		{name: "3rd", time: time.Date(2020, 12, 3, 13, 37, 0, 0, time.UTC), res: "Thursday, the 3rd of December at 13:37 UTC (2020)"},
		{name: "21st", time: time.Date(2020, 12, 21, 13, 37, 0, 0, time.UTC), res: "Monday, the 21st of December at 13:37 UTC (2020)"},
		{name: "22nd", time: time.Date(2020, 12, 22, 13, 37, 0, 0, time.UTC), res: "Tuesday, the 22nd of December at 13:37 UTC (2020)"},
		{name: "23rd", time: time.Date(2020, 12, 23, 13, 37, 0, 0, time.UTC), res: "Wednesday, the 23rd of December at 13:37 UTC (2020)"},
		{name: "31st", time: time.Date(2020, 12, 31, 13, 37, 0, 0, time.UTC), res: "Thursday, the 31st of December at 13:37 UTC (2020)"},
		{name: "nth", time: time.Date(2020, 12, 4, 13, 37, 0, 0, time.UTC), res: "Friday, the 4th of December at 13:37 UTC (2020)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := formatDate(tt.time)
			assertEqual(t, res, tt.res)
		})
	}
}
