package bot

import (
	"bufio"
	"log"
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/daenney/gdq"
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
	Client    *mautrix.Client
	cache     *ttlcache.Cache
	Announcer *time.Timer
}

func New(homeserverURL, userID, accessToken string) (b *bot, err error) {
	uid := id.UserID(userID)
	client, err := newMatrixClient(homeserverURL, uid, accessToken)
	if err != nil {
		return nil, err
	}

	b = &bot{
		Client: client,
		cache:  ttlcache.NewCache(),
	}

	b.cache.SkipTTLExtensionOnHit(true)
	b.cache.SetTTL(10 * time.Minute)
	b.cache.SetLoaderFunction(func(key string) (data interface{}, ttl time.Duration, err error) {
		s, err := gdq.GetSchedule(gdq.Latest, safeClient)
		if err != nil {
			log.Printf("loader: failed to load schedule into cache: %s\n", err)
		}
		// reset the announcer timer when the cache gets reloaded to ensure
		// we notice schedule changes
		b.Announcer.Reset(1 * time.Second)
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

	b.Announcer = time.NewTimer(5 * time.Second)

	return b, nil
}

func (b *bot) primeCache() {
	s, err := gdq.GetSchedule(gdq.Latest, safeClient)
	if err != nil {
		log.Print("primer: failed to load cache with schedule")
		return
	}
	b.cache.SetWithTTL("sched", s, 10*time.Minute)
	log.Print("primer: loaded cache with schedule")
	return
}

func (b *bot) handleMessage(ms mautrix.EventSource, ev *event.Event) {
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
		log.Printf("failed to get and filter schedule: %s", err)
		msg = &event.MessageEventContent{
			Body: `Sorry, something went wrong handling your request. This usually means 
			the GDQ schedule couldn't be retrieved. Please try again in a minute.`,
			MsgType: event.MsgNotice,
		}
	}

	msg.SetReply(ev)
	_, err = b.Client.SendMessageEvent(ev.RoomID, event.EventMessage, msg)
	if err != nil {
		log.Printf("failed to send message: %s", err)
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

	log.Print("attempting to join room: ", ev.RoomID)

	time.Sleep(1 * time.Second)
	_, err := b.Client.JoinRoom(ev.RoomID.String(), "", struct{}{})
	if err != nil {
		log.Printf("failed to join room: %s, error: %s\n", ev.RoomID, err.Error())
	}

	return
}
