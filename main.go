package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/daenney/gdqbot/internal/bot"

	"go.uber.org/zap"
)

func main() {
	hs := flag.String("homeserver", "", "Homeserver URL, for example 'matrix.org', or 'https://domain.tld:8843/'")
	token := flag.String("access-token", "", "Access token, or use the GDQBOT_ACCESS_TOKEN environment variable")
	user := flag.String("user", "", "Matrix ID for the bot, @bot:domain.tld")
	showVersion := flag.Bool("version", false, "show GDQBot version and build info")
	debug := flag.Bool("debug", false, "enable debug output")
	format := flag.String("log.format", "console", "one of json or console")
	formatTime := flag.Bool("log.timestamp", true, "include timestamp in log output")
	event := flag.String("event", "", "Event ID or name for the bot to use")
	userAgent := flag.String("user-agent", "", "user-agent to use when querying. Set this to somewhere the GDQ can contact you in case your deployment misbehaves")

	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "{\"version\": \"%s\", \"commit\": \"%s\", \"date\": \"%s\"}\n", version, commit, date)
		os.Exit(0)
	}

	if *hs == "" {
		fmt.Fprintln(os.Stderr, "No homeserver specified, please specify using -homeserver")
		os.Exit(1)
	}
	if *user == "" {
		fmt.Fprintln(os.Stderr, "No username specified, please specify using -user")
		os.Exit(1)
	}
	if *token == "" {
		*token = os.Getenv("GDQBOT_ACCESS_TOKEN")
	}
	if *token == "" {
		fmt.Fprintln(os.Stderr, "No access token specified, please specify using -access-token or set the GDQBOT_ACCESS_TOKEN environment variable")
	}
	if *userAgent == "" {
		fmt.Fprintln(os.Stderr, "No user-agent specified. Please set this to something the GDQ staff can use to contact you")
	}

	l := bot.NewLogger(*debug, *format, *formatTime)
	b, err := bot.New(*hs, *user, *token, *event, *userAgent, l)
	if err != nil {
		l.Error("failed to initialise", zap.Error(err))
	}

	l.Info("syncing timeline and handling requests")

	go func() {
		if err := b.Client.Sync(); err != nil {
			b.Client.Client.CloseIdleConnections()
			l.Error("sync encountered an error", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func(ctx context.Context) {
		b.Announce(ctx)
	}(ctx)

	<-ctx.Done()
	stop()
	fmt.Fprintf(os.Stdout, "initiating graceful shutdown, Ctrl+C again to force")

	b.Client.StopSync()
	l.Sync()
	b.Client.Client.CloseIdleConnections()
	os.Exit(0)
}
