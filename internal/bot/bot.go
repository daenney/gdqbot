package bot

import (
	"bufio"
	"context"
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/daenney/gdq/v2"
	"go.uber.org/zap"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const eventID = "io.github.daenney.gdqbot.internal"

var filter = mautrix.Filter{
	AccountData: mautrix.FilterPart{
		Limit: 20,
		NotTypes: []event.Type{
			event.NewEventType(eventID),
		},
	},
	Room: mautrix.RoomFilter{
		Timeline: mautrix.FilterPart{
			Limit: 20,
			Types: []event.Type{
				event.EventMessage,
				event.StateMember,
			},
		},
		Ephemeral: mautrix.FilterPart{
			Limit: 20,
			NotTypes: []event.Type{
				event.EphemeralEventTyping,
				event.EphemeralEventReceipt,
			},
		},
	},
	EventFields: []string{
		"type",
		"event_id",
		"room_id",
		"state_key",
		"content.body",
		"content.membership",
	},
	Presence: mautrix.FilterPart{
		Limit: 20,
		NotTypes: []event.Type{
			event.EphemeralEventPresence,
		},
	},
}

// bot represents our bot
type bot struct {
	Client     *mautrix.Client
	cache      *ttlcache.Cache
	timerReset chan struct{}
	log        *zap.Logger
	gdqClient  *gdq.Client
	event      *gdq.Event
}

func New(homeserverURL, userID, accessToken string, log *zap.Logger) (b *bot, err error) {
	uid := id.UserID(userID)
	client, err := newMatrixClient(homeserverURL, uid, accessToken)
	if err != nil {
		return nil, err
	}

	b = &bot{
		Client:    client,
		cache:     ttlcache.NewCache(),
		log:       log.Named("bot"),
		gdqClient: gdq.New(context.Background(), safeClient),
	}

	ev, err := b.gdqClient.Latest()
	if err != nil {
		b.log.Fatal("unable to detect the latest GDQ event")
	}

	b.event = ev

	b.cache.SkipTTLExtensionOnHit(true)
	b.cache.SetTTL(10 * time.Minute)
	b.cache.SetLoaderFunction(func(key string) (data interface{}, ttl time.Duration, err error) {
		s, err := b.gdqClient.Schedule(b.event)
		if err != nil {
			b.log.Named("cache").Error("failed to load schedule into cache", zap.Error(err))
		}
		if err == nil {
			b.resetTimer()
		}
		return s, 10 * time.Minute, err
	})
	b.primeCache()

	fID, err := b.Client.CreateFilter(&filter)
	if err != nil {
		return nil, err
	}
	b.Client.Store.SaveFilterID(uid, fID.FilterID)

	syncer := b.Client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, b.handleMessage)
	syncer.OnEventType(event.StateMember, b.handleMembership)

	b.timerReset = make(chan struct{})

	return b, nil
}

func (b *bot) primeCache() {
	l := b.log.Named("cache")
	s, err := b.gdqClient.Schedule(b.event)
	if err != nil {
		l.Error("failed to prime cache with schedule", zap.Error(err))
		return
	}
	b.cache.SetWithTTL("sched", s, 10*time.Minute)
	l.Info("primed cache with schedule")
	return
}

func (b *bot) handleMessage(ms mautrix.EventSource, ev *event.Event) {
	l := b.log.Named("message")
	body := ev.Content.AsMessage().Body

	r := strings.NewReader(body)
	scanner := bufio.NewScanner(r)

	content := ""
	// Get the first line. Any additional lines are ignored as garbage
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "!gdq") {
			return
		}
		content = line
		break
	}

	var msg *event.MessageEventContent
	var err error

	fields := strings.Split(content, " ")
	if len(fields) < 2 {
		msg, err = b.msgHelp()
	} else {
		switch fields[1] {
		case "event", "title":
			msg, err = b.msgScheduleForEvent(strings.Join(fields[2:], " "))
		case "runner":
			msg, err = b.msgScheduleForRunner(strings.Join(fields[2:], " "))
		case "host":
			msg, err = b.msgScheduleForHost(strings.Join(fields[2:], " "))
		case "next":
			msg, err = b.msgScheduleNext()
		case "help":
			msg, err = b.msgHelp()
		default:
			msg, err = b.msgScheduleForEvent(strings.Join(fields[1:], ""))
		}
	}

	if err != nil {
		l.Error("failed to get and filter schedule", zap.Error(err))
		msg = &event.MessageEventContent{
			Body: `Sorry, something went wrong handling your request. This usually means 
			the GDQ schedule couldn't be retrieved. Please try again in a minute.`,
			MsgType: event.MsgNotice,
		}
	}

	msg.SetReply(ev)
	_, err = b.Client.SendMessageEvent(ev.RoomID, event.EventMessage, msg)
	if err != nil {
		l.Error("failed to send message", zap.Error(err))
	}
}

func (b *bot) handleMembership(_ mautrix.EventSource, ev *event.Event) {
	e := ev.Content.AsMember()
	if e.Membership != event.MembershipInvite {
		// Ignore it if it's not an invite
		return
	}

	if *ev.StateKey != b.Client.UserID.String() {
		// Ignore it if it's not meant for us
		return
	}

	l := b.log.Named("membership")

	l.Info("attempting to join room", zap.String("room", ev.RoomID.String()))

	time.Sleep(1 * time.Second)
	_, err := b.Client.JoinRoom(ev.RoomID.String(), "", struct{}{})
	if err != nil {
		l.Error("failed to join room", zap.String("room", ev.RoomID.String()), zap.Error(err))
	}

	return
}
