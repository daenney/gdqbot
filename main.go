package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	hs := flag.String("homeserver", "", "Homeserver URL, for example 'matrix.org', or 'https://domain.tld:8843/'")
	domain := flag.String("domain", "", "Matrix domain, for example 'matrix.org'. If it's the same as -homeserver you can omit it")
	token := flag.String("access-token", "", "Access token, or use the GDQBOT_ACCESS_TOKEN environment variable")
	user := flag.String("user", "", "Username for the bot, without the @. For example 'gdqbot'")

	flag.Parse()

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
	if *domain == "" {
		*domain = *hs
	}

	b, err := newBot(*hs, *user, *domain, *token)
	if err != nil {
		log.Fatalln(fmt.Errorf("couldn't initialise the bot: %s", err))
	}

	log.Print("syncing timeline and handling requests")
	if err := b.client.Sync(); err != nil {
		log.Fatalln(fmt.Errorf("sync encountered an error: %s", err))
	}
}
