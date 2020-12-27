package bot

import (
	"context"
	"time"

	"github.com/daenney/gdq"
	"go.uber.org/zap"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *bot) resetTimer() {
	go func() {
		b.announcer <- struct{}{}
	}()
	b.log.Named("announcer").Debug("reset timers")
}

func (b *bot) Announce(ctx context.Context) {
	l := b.log.Named("announcer")
	l.Info("started routine")

	var last string

	t := time.NewTimer(1 * time.Second)

	for {
		select {
		case <-b.announcer:
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
		case <-ctx.Done():
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
			return
		case <-t.C:
		}

		s, err := b.cache.Get("sched")
		if err != nil {
			// Retry a bit later if we can't load the schedule
			t.Reset(10 * time.Second)
			continue
		}

		ev := s.(*gdq.Schedule).NextEvent()
		dur := ev.Start.Sub(time.Now().UTC())

		if dur < 0 {
			// At this point there's no known future events so lets take a
			// long nap and check again later
			last = ""
			t.Reset(1 * time.Hour)
			continue
		}

		if last == ev.Title {
			l.Debug("not announcing event",
				zap.String("reason", "already announced"),
				zap.String("event", ev.Title))
			t.Reset(dur)
			continue
		}

		if dur > 10*time.Minute {
			// We don't want to announce events more than 10min
			// before the start time
			l.Debug("not announcing event",
				zap.String("reason", "too far in the future"),
				zap.String("event", ev.Title),
				zap.String("duration", dur.String()),
				zap.Time("start", ev.Start))
			t.Reset(dur - 10*time.Minute)
			continue
		}

		// It's time to announce something!
		l.Debug("announcing event",
			zap.String("event", ev.Title),
			zap.String("duration", dur.String()),
			zap.Time("start", ev.Start))

		rooms, err := b.Client.JoinedRooms()
		if err != nil {
			// Assume some temporary issue occurred, retry in a bit
			t.Reset(5 * time.Second)
		}

		sendTo := []id.RoomID{}
		for _, room := range rooms.JoinedRooms {
			members, err := b.Client.JoinedMembers(room)
			if err != nil {
				l.Error("failed to retrieve memberships",
					zap.String("room", room.String()),
					zap.Error(err))
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
				l.Error("failed to announce event",
					zap.String("event", ev.Title),
					zap.String("room", room.String()),
					zap.Error(err))
			}
		}

		last = ev.Title
		// Reset the timer to fire once the event we just announced has started
		t.Reset(dur)
	}
}
