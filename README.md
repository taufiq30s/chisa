# Chisa

A Discord bot for the Akashic Record of Chisa server. Plays music, converts
currency, screens scam links, and runs interactive quizzes.

The bot is named after Chisa Arfa, the server's mascot.

## What it does

### Music

yt-dlp extracts the audio and the YouTube metadata, ffmpeg transcodes it to
Opus, and the result streams over a direct Discord voice connection. There is
no external audio server to run. The bot joins voice only once a track actually
starts, and leaves again after ten minutes with nothing playing.

A playlist URL expands to at most 100 tracks and answers with one summary card
carrying the count and total duration. A private or deleted playlist is reported
as such instead of failing quietly. A `watch?v=X&list=Y` URL plays just that
video, since that is what someone pasting it means.

### Currency

Live rates from a currency provider, plus a simulation of a Wise transfer that
prices the amount from either the sender's or the receiver's side. The list of
supported currencies comes from whichever provider is active and is cached in
Redis for a week, so there is no count here: the set is whatever the provider
returns, not a number maintained in this file.

### Moderation

Every message is checked against the [Discord AntiScam
list](https://github.com/Discord-AntiScam/scam-links), refreshed daily at 00:00
UTC. A confirmed scam link or an `@everyone`/`@here` burst gets the sender
timed out and drops an alert in the moderator channel with buttons to ban or
release. Anyone who posts in the honeypot channel is banned on the spot.

Separately, `/verify` sends a request to the moderator channel, where accepting
it grants the verified role and opens the server.

### Quiz

A WebSocket server and a handful of HTTP endpoints that an external web client
drives. The bot owns the session and manages who may speak on a Stage channel:
promote, suppress, reject, and read the queue.

## Commands

### Music

| Command | What it does |
| --- | --- |
| `/music play <query>` | Queue a song by name or URL |
| `/music pause` | Pause the current track |
| `/music resume` | Resume playback |
| `/music skip` | Skip to the next track |
| `/music stop` | Stop playback and empty the queue |
| `/music disconnect` | Leave the voice channel |
| `/music clear` | Remove every queued track (the current track keeps playing) |
| `/music list` | Show the queue, ten tracks per page |

### Currency

| Command | What it does |
| --- | --- |
| `/currency convert <amount> <from> <to>` | Convert between two currencies |
| `/currency simulate <mode> <amount> <source> <target>` | Simulate a Wise transfer, priced from either end |

`mode` is `send` to give the amount in the source currency, or `receiver` to
give it in the target currency.

### Verification

| Command | What it does |
| --- | --- |
| `/verify` | Send a verification request for a moderator to review |

### Interactive quiz

| Command | What it does |
| --- | --- |
| `/interactive-quiz start <channel>` | Start a quiz session on a Stage channel |
| `/interactive-quiz stop` | End the current session |

## Running it

You need Go 1.27, a Redis server, ffmpeg, and yt-dlp. ffmpeg must be on your
`PATH`. yt-dlp can be too, but the bot looks in `./bin/` first, which is where
the `Dockerfile` puts it. Both directories are yours to populate; `bin/` is
gitignored.

yt-dlp plugins go in `plugins/<name>/yt_dlp_plugins/`; the containers mount that
directory as `bin/yt-dlp-plugins`, which yt-dlp scans automatically.

1. Copy the example config: `cp .env.example .env`
2. Fill in the eleven required variables below.
3. Build and start: `make run`

### Required variables

All eleven are mandatory. If any is missing the bot refuses to start and lists
every missing key in one message, so you can fix them all in a single pass.

| Variable | Value |
| --- | --- |
| `VERSION` | Any string. Appears in embed footers. |
| `BOT_TOKEN` | Bot token, from the Bot tab of the Discord developer portal. |
| `AKASHIC_SERVER_ID` | Server ID. Right-click the server icon, Copy Server ID. |
| `REDIS_URL` | Redis connection string, for example `redis://localhost:6379`. |
| `AKASHIC_VERIFIED_ROLE_ID` | The role `/verify` grants. Right-click the role, Copy Role ID. |
| `AKASHIC_VERIF_CHANNEL_ID` | Channel where `/verify` is allowed. |
| `CHISA_MOD_CHANNEL_ID` | Channel that receives moderation alerts and verification requests. |
| `CHISA_LOG_CHANNEL_ID` | Channel that receives the bot's log embeds. |
| `AKASHIC_WELCOME_CHANNEL_ID` | Channel for welcome messages. |
| `AKASHIC_RULE_CHANNEL_ID` | Channel linked from the welcome message. |
| `AKASHIC_HONEYPOT_CHANNEL_ID` | Trap channel. Anyone who posts in it is banned. |

The IDs are only visible in the Discord client with developer mode enabled:
User Settings, Advanced, Developer Mode.

### Optional variables

`CURRENCY_PROVIDER` selects the rate source and defaults to `wise`, which
requires `WISE_TOKEN`. Setting it to `currency_api` requires
`CURRENCY_API_TOKEN` instead.

`QUIZ_WS_PORT` defaults to 8080. `PROXY_URL` routes outbound yt-dlp traffic
through a SOCKS5 proxy, useful where YouTube is blocked. `YTDLP_PATH` overrides
the bundled binary, and falls back to `yt-dlp` on `PATH`.

One variable is not in `.env.example`: `YOUTUBE_API_KEY`. Music search needs it.
Without it the music provider fails to construct and the bot starts anyway,
answering every `/music` command with "unavailable" instead of refusing to
boot. That is deliberate — a bot with no music key still does moderation and
currency — but it is easy to misread as a broken install.

## Deployment

`docker compose up` brings up Redis and the bot together, reading `.env` and
mounting `logs/`.

The `Dockerfile` builds for whatever platform the build targets (`TARGETARCH`)
and bakes ffmpeg and a checksum-verified upstream yt-dlp into `./bin/yt-dlp`.
`.github/workflows/deployment.yml` builds that image and pushes it to Docker
Hub on every push to `development`, then restarts the container on the VPS.

To add an extractor without rebuilding, drop it on the host as
`plugins/<name>/yt_dlp_plugins/extractor/<file>.py` and mount the directory over
`/app/bin/yt-dlp-plugins:ro` (compose does this already; on the VPS append the
same `-v` to `VPS_DOCKER_OPTS`). yt-dlp picks it up on the next `/music` call.
The layout matters: a plugin directory pointing straight at `yt_dlp_plugins/`
loads nothing without `--plugin-dirs`, which the bot never passes.

## Development

### Layout

| Package | Responsibility |
| --- | --- |
| `internal/bot` | Startup, Discord session, cron jobs |
| `internal/config` | Environment loading and validation |
| `internal/handlers` | Slash command definitions and routing |
| `internal/events` | Discord event listeners |
| `internal/responses` | Embed builders and interaction replies |
| `internal/music` | Player, queue, yt-dlp and ffmpeg pipeline |
| `internal/currency` | Rates, providers, Wise simulation |
| `internal/moderation` | Scam detection, timeouts, verification |
| `internal/quiz` | WebSocket and HTTP session server |
| `internal/model` | Shared data types |

`cmd/` holds the entry point and wires the rest together.

### Make targets

`make build` compiles to `bin/chisa` and `make run` builds then starts it.
`make dev` runs `air` for live reload, if you have it installed. `make fmt`,
`make vet`, `make lint`, and `make tidy` are the usual Go chores.
`make docker-up` and `make docker-down` drive the compose stack, and
`make clean` removes `bin/`.

There is no `make test` target. Run the tests directly with `go test ./...`.

### Tests and CI

Most tests are offline. The exception is
`internal/music/providers/ytdlp_live_test.go`, which really does fetch from
YouTube and fails without network access.

CI runs three jobs on every push and pull request: build and vet, test, and
lint. Lint is golangci-lint v2.15.0 reading `.golangci.yml`. Work on a branch
off `development` and make sure CI is green before opening a pull request.
