package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/daenney/gdq"
	"maunium.net/go/mautrix/event"
)

const (
	htmlEventMsg  = "<b>%s</b> on %s run by <i>%s</i> with commentary from <i>%s</i> lasting %s"
	plainEventMsg = "%s on %s run by %s with commentary from %s lasting %s"
	singleMatch   = "There is one event matching your query:"
	multiMatch    = "There are multiple events matching your query:"
	dateFormat    = "%s, the %s of %s at %02d:%02d UTC (%04d)"
)

type filteredHandler func(filter string) (*event.MessageEventContent, error)

func (b *bot) msgScheduleForEvent(filter string) (*event.MessageEventContent, error) {
	s, err := b.cache.Get("sched")
	if err != nil {
		return nil, err
	}
	return msgSchedule(s.(*gdq.Schedule).ForTitle(filter)), nil
}

func (b *bot) msgScheduleForRunner(filter string) (*event.MessageEventContent, error) {
	s, err := b.cache.Get("sched")
	if err != nil {
		return nil, err
	}
	return msgSchedule(s.(*gdq.Schedule).ForRunner(filter)), nil
}

func (b *bot) msgScheduleForHost(filter string) (*event.MessageEventContent, error) {
	s, err := b.cache.Get("sched")
	if err != nil {
		return nil, err
	}
	return msgSchedule(s.(*gdq.Schedule).ForHost(filter)), nil
}

func (b *bot) msgScheduleNext() (*event.MessageEventContent, error) {
	s, err := b.cache.Get("sched")
	if err != nil {
		return nil, err
	}

	e := s.(*gdq.Schedule).NextEvent()
	if e == nil {
		return &event.MessageEventContent{
			Body:    "There are currently no further events scheduled.",
			MsgType: event.MsgNotice,
		}, nil
	}

	return &event.MessageEventContent{
		Body:          fmt.Sprintf("The next event is: %s", plainEvent(e)),
		MsgType:       event.MsgNotice,
		Format:        event.FormatHTML,
		FormattedBody: fmt.Sprintf("The next event is: %s", htmlEvent(e)),
	}, nil
}

func (b *bot) msgAnnounce(e *gdq.Event) *event.MessageEventContent {
	d := gdq.Duration{Duration: e.Start.Sub(time.Now().UTC())}
	return &event.MessageEventContent{
		Body:          fmt.Sprintf("An event is starting in approximately %s: %s", d, plainEvent(e)),
		MsgType:       event.MsgNotice,
		Format:        event.FormatHTML,
		FormattedBody: fmt.Sprintf("An event is starting in approximately %s: %s", d, htmlEvent(e)),
	}
}

func (b *bot) msgHelp() (*event.MessageEventContent, error) {
	return &event.MessageEventContent{
		Body: `Supported commands are 'help', 'next', 'event|title', 'host' and 'runner'. The 'event|title', 
		'host' and 'runner' commands let you filter the schedule matching either the name of a game, the runner 
		or the host. It's not possible to stack filters. The supplied filter does not have to be an exact match. 
		Diacritics, capitalisation and punctuation are ignored when checking for matches. If the command doesn't 
		match, it's interpreted as 'event <rest>' The 'next' command returns the next/upcoming run.`,
		MsgType: event.MsgNotice,
	}, nil
}

func msgSchedule(s *gdq.Schedule) *event.MessageEventContent {
	if len(s.Events) == 0 {
		return &event.MessageEventContent{
			Body:    "There are no events matching your query.",
			MsgType: event.MsgNotice,
		}
	}

	plainBuilder := strings.Builder{}
	htmlBuilder := strings.Builder{}
	if len(s.Events) == 1 {
		htmlBuilder.WriteString(singleMatch + "<br><ul>")
		plainBuilder.WriteString(singleMatch + "\n")
	}
	if len(s.Events) > 1 {
		htmlBuilder.WriteString(multiMatch + "<br><ul>")
		plainBuilder.WriteString(multiMatch + "\n")
	}
	num := len(s.Events)
	for i, event := range s.Events {
		htmlBuilder.WriteString("<li>" + htmlEvent(event) + "</li>")
		plainBuilder.WriteString("* " + plainEvent(event) + "\n")
		if i == num-1 {
			htmlBuilder.WriteString("</ul>")
		}
	}

	return &event.MessageEventContent{
		Body:          plainBuilder.String(),
		MsgType:       event.MsgNotice,
		Format:        event.FormatHTML,
		FormattedBody: htmlBuilder.String(),
	}
}

func plainEvent(event *gdq.Event) string {
	runners, hosts, estimate := formatMetadata(event)
	return fmt.Sprintf(plainEventMsg, event.Title, formatDate(event.Start), runners, hosts, estimate)
}

func htmlEvent(event *gdq.Event) string {
	runners, hosts, estimate := formatMetadata(event)
	return fmt.Sprintf(htmlEventMsg, event.Title, formatDate(event.Start), runners, hosts, estimate)
}
