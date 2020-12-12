package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/daenney/gdq"
)

func formatMetadata(event *gdq.Event) (runners, hosts, estimate string) {
	runners = "unknown"
	if len(event.Runners) > 0 {
		runners = formatHandles(event.Runners)
	}
	hosts = "unknown"
	if len(event.Hosts) > 0 {
		hosts = formatHandles(event.Hosts)
	}
	estimate = "unknown amount of time"
	if event.Estimate.Duration != 0 {
		estimate = event.Estimate.String()
	}
	return runners, hosts, estimate
}

func formatHandles(elems []string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	case 2:
		return fmt.Sprintf("%s and %s", elems[0], elems[1])
	default:
		return strings.Join(elems[0:len(elems)-1], ", ") + " and " + elems[len(elems)-1]
	}
}

func formatDate(t time.Time) string {
	var postfix string
	switch t.Day() {
	case 1, 21, 31:
		postfix = "st"
	case 2, 22:
		postfix = "nd"
	case 3, 23:
		postfix = "rd"
	default:
		postfix = "th"
	}
	return fmt.Sprintf(dateFormat,
		t.Weekday().String(),
		fmt.Sprintf("%d%s", t.Day(), postfix),
		t.Month().String(),
		t.Hour(),
		t.Minute(),
		t.Year())
}
