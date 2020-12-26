package bot

import (
	"context"
	"log"
	"time"

	"github.com/daenney/gdq"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *bot) Announce(ctx context.Context) {
	var last string
	for {
		select {
		case <-b.Announcer.C:
			s, err := b.cache.Get("sched")
			if err != nil {
				// Retry a bit later if we can't load the schedule
				b.Announcer.Reset(10 * time.Second)
				continue
			}

			ev := s.(*gdq.Schedule).NextEvent()
			dur := ev.Start.Sub(time.Now().UTC())

			if dur < 0 {
				// At this point there's no known future events so lets take a
				// long nap and check again later
				last = ""
				b.Announcer.Reset(1 * time.Hour)
				continue
			}

			if last == ev.Title {
				log.Printf("already announced: %s, skipping", ev.Title)
				b.Announcer.Reset(dur)
				continue
			}

			if dur > 10*time.Minute {
				// We don't want to announce events more than 10min
				// before the start time
				log.Printf("not announcing: %s, event is too far in the future: %s\n", ev.Title, dur)
				b.Announcer.Reset(dur - 10*time.Minute)
				continue
			}

			// It's time to announce something!
			log.Printf("announcing: %s, duration is: %s\n", ev.Title, dur)
			rooms, err := b.Client.JoinedRooms()
			if err != nil {
				// Assume some temporary issue occurred, retry in a bit
				b.Announcer.Reset(5 * time.Second)
			}

			sendTo := []id.RoomID{}
			for _, room := range rooms.JoinedRooms {
				members, err := b.Client.JoinedMembers(room)
				if err != nil {
					log.Printf("failed to retrieve memberships for room: %s", room)
					// Skip rooms we can't figure out the members for
					continue
				}
				if len(members.Joined) > 2 {
					sendTo = append(sendTo, room)
				}
			}
			msg := b.msgAnnounce(ev)

			for _, room := range sendTo {
				_, err := b.Client.SendMessageEvent(room, event.EventMessage, msg)
				if err != nil {
					log.Printf("failed to announce event: %s to room: %s\n", ev.Title, room)
				}
			}

			last = ev.Title
			// Reset the timer to fire once the event we just announced has started
			b.Announcer.Reset(dur)
		case <-ctx.Done():
			if !b.Announcer.Stop() {
				<-b.Announcer.C
			}
			return
		}
	}

}
