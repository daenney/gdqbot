<h1 align="center">
🤖 GDQBot 🎮
</h1>
<h4 align="center">A Matrix bot for Games Done Quick</h4>
<p align="center">
    <a href="https://github.com/daenney/gdqbot/actions?query=workflow%3ACI"><img src="https://github.com/daenney/gdqbot/workflows/CI/badge.svg" alt="Build Status"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/daenney/gdqbot" alt="License: AGPLv3"></a>
</p>

[Games Done Quick (GDQ)](https://gamesdonequick.com/) is a regular
speedrunning event that collects money for charity. The event is incredibly
fun, especially if you enjoy seeing your favourite games torn to shreds by
amazing runners and supported with great commentary and prizes to win.

This [Matrix](https://matrix.org) bot lets you:
* Query for information about the current/upcoming GDQ schedule
* Announces upcoming events into all joined rooms (except if rooms have less
  than three participants because that's our heuristic for a DM)

Add the bot to a server, open a DM and issue `!gdq` or `!gdq help` to get
instructions.

This bot is built using the [GDQ Go library](https://github.com/daenney/gdq).

## Installation

There are prebuilt binaries as well as Docker images available for every release
from v0.1.4 onwards. You can find them [over here](https://github.com/daenney/gdqbot/releases).

|Platform|Architecture|Binary|Docker
|---|---|---|---|
|Windows|amd64|✅|❌|
|macOS|amd64|✅|❌|
|macOS|arm64/M1<sup id="a1">[1](#f1)</sup>|❌|❌|
|Linux|amd64|✅|✅|
|Linux|arm64|✅|✅|
|Linux|armv7/amrhf|✅|✅|

<b id="f1"><sup>1</sup></b> Pending Go 1.16 release [↩](#a1)

### Docker

All Docker images use [distroless as the base](https://github.com/GoogleContainerTools/distroless)
and builds on the `nonroot` version of the image. This means the bot never
runs as root, regardless of user remapping/user namespacing. This image
can/should be run as read-only.

#### `docker run`

```sh
$ docker run --name matrix-gdqbot -e GDQBOT_ACCESS_TOKEN=<token> ghcr.io/daeney/gdqbot:<tag> \
  -homeserver https://example.com \
  -user @gdqbot:example.com
```

#### `docker-compose`

```yaml
---
version: "2.0"

services:
  matrix-gdq-bot:
    image: ghcr.io/daenney/gdqbot:<tag>
    container_name: matrix-gdqbot
    restart: unless-stopped
    command:
      - -homeserver https://example.com
      - -user @gdqbot:example.com
    environment:
      - GDQBOT_ACCESS_TOKEN=secret_value_goes_here

```

## Building

You can `go get` the code, or `git clone` and then run a `go build` followed
by a `go test` to ensure everything is OK.

You can build the bot using `go build -trimpath` or install it directly using
`go install github.com/daenney/gdqbot`. See `go help install` for where the
binaries will end up.

To embed the version, commit and date at build time you'll need to add
`-X main.version=VERSION -X main.commit=SHA -X main.date=DATE` and compute
the right values yourself.

## Contributing

PRs welcome! Fork+clone the repo and send me a patch. Please ensure that:
* Make small commits that encapsulate one functional change at a time
  (implementation change, the associated tests and any doc changes)
* Every commit explains what it's trying to achieve and why
* The tests pass
