package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/daenney/gdq/v3"
)

func formatMetadata(event *gdq.Run) (runners, hosts, estimate string) {
	runners = formatHandles(event.Runners)
	hosts = formatHandles(event.Hosts)
	estimate = "unknown amount of time"
	if event.Estimate.Duration != 0 {
		estimate = event.Estimate.String()
	}
	return runners, hosts, estimate
}

func formatHandles(elems []gdq.Talent) string {
	switch len(elems) {
	case 0:
		return "unknown"
	case 1:
		return elems[0].Name
	case 2:
		return fmt.Sprintf("%s and %s", elems[0].Name, elems[1].Name)
	default:
		batch := elems[0 : len(elems)-1]
		names := make([]string, 0, len(batch))
		for _, t := range batch {
			names = append(names, t.Name)
		}
		return strings.Join(names, ", ") + " and " + elems[len(elems)-1].Name
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
