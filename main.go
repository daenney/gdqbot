package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	hs := flag.String("homeserver", "", "Homeserver URL, for example 'matrix.org', or 'https://domain.tld:8843/'")
	token := flag.String("access-token", "", "Access token, or use the GDQBOT_ACCESS_TOKEN environment variable")
	user := flag.String("user", "", "Matrix ID for the bot, @bot:domain.tld")
	showVersion := flag.Bool("version", false, "show GDQBot version and build info")

	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "{\"version\": \"%s\", \"commit\": \"%s\", \"date\": \"%s\"}\n", version, commit, date)
		os.Exit(0)
	}

	if *hs == "" {
		log.Fatalln("No homeserver specified, please specify using -homeserver")
	}
	if *user == "" {
		log.Fatalln("No username specified, please specify using -user")
	}
	if *token == "" {
		*token = os.Getenv("GDQBOT_ACCESS_TOKEN")
	}
	if *token == "" {
		log.Fatalln("No access token specified, please specify using -access-token or set the GDQBOT_ACCESS_TOKEN environment variable")
	}

	b, err := newBot(*hs, *user, *token)
	if err != nil {
		log.Fatalln(fmt.Errorf("couldn't initialise the bot: %s", err))
	}

	log.Print("syncing timeline and handling requests")
	if err := b.client.Sync(); err != nil {
		log.Fatalln(fmt.Errorf("sync encountered an error: %s", err))
	}
}
