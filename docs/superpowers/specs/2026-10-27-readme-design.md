# README rewrite: design

Status: approved, ready for implementation plan
Date: 2026-10-27
Replaces: `FEATURES.md`

## Goal

Replace `FEATURES.md` with a `README.md` that works as a GitHub landing page
for someone from the Akashic Record of Chisa server who wants to contribute.

Two problems to solve. The file is named and shaped like an internal inventory,
not a README. And it has drifted roughly seven months behind the code: its last
substantive edit was `eaa545f` (2026-03-13, "Prepare to refactor project"), and
the only later touch was a re-save during the Lavalink removal.

## Audience

Contributors from the Chisa Discord server. Readers know the server and the bot
as users. They do not know the codebase, the toolchain, or how a bot is
deployed. The README therefore opens with what the bot is and what it does, then
gets to a working local run as fast as possible, then explains where code lives.

Readers outside the server are not the target but are not excluded. Nothing in
the document should require server membership to follow except the channel and
role ID table, which is inherently server-specific.

## Constraints

- Language: English. Matches the existing file and the code comments.
- Voice: plain. Short sentences. No marketing, no emoji section headers, no
  tick/cross bullet lists, no "Overview:" lead-ins, no em-dash-per-sentence
  cadence.
- One file. The detailed inventory is not moving to `docs/`; it is being
  replaced by prose that covers the same ground.
- Every factual claim must be checkable against the tree. Nothing is carried
  over on faith.
- The "Current Limitations" section and the `REFACTORING_PLAN.md` pointer are
  dropped. The pointer already dead-ends, and the limitation list reads as an
  inventory item rather than a contribution invitation.

## Approach

Chosen: rewrite from scratch, keeping only verified facts.

The existing file's structure is what reads as machine-written, more than its
word choice. `## Music Features` followed by `### Music Capabilities` followed
by nine ticked bullets, then `### Technical Features`, then a second ticked list
under Moderation, then closing sections named after the document ("Technical
Stack", "Feature Categorization"). Editing sentences inside that skeleton leaves
the skeleton. A fresh pass writes in one voice and the decoration cannot survive
because nothing carries it forward.

Rejected: a surgical edit (rename plus factual fixes). Smaller diff, but every
heading stays, and the headings are most of the problem.

Rejected: a two-register rewrite, narrative on top and a dense reference half
below. Attractive for a README people return to, but the reference half would
restate what `/help` already shows in Discord and would rot the same way
`FEATURES.md` did. A README stays correct by not duplicating the code.

## Structure

Target length 150-180 lines. The current file is 254 and carries less
information, because of the tick/cross padding.

### 1. Title and opening

```markdown
# Chisa

A Discord bot for the Akashic Record of Chisa server. Plays music, converts
currency, screens scam links, and runs interactive quizzes.
```

Followed by one short paragraph naming the mascot origin: the bot is named after
Chisa Arfa, the server's mascot. That is the one piece of context a newcomer
cannot infer from the code.

No badges. No CI badge, because `development` has no green run yet. No license
badge, because there is no `LICENSE` file.

### 2. What it does

One short subsection per feature area, in prose. Not capability lists; a
sentence or two on what the thing does plus the one detail that is not obvious
from the command names.

- **Music.** yt-dlp extracts audio and YouTube metadata, ffmpeg transcodes to
  Opus, and the result streams over a direct Discord voice connection. No
  external audio server. Playlists queue up to 100 tracks behind a single
  summary card. The bot joins voice lazily, only once a track actually starts,
  and disconnects after 10 minutes idle.
- **Currency.** Real-time conversion plus Wise transfer simulation. The
  currency list comes from the active provider and is cached in Redis for a
  week, so the README quotes no count: the supported set is whatever the
  provider returns, not a number maintained here.
- **Moderation.** Scans every message against the Discord AntiScam list
  (`Discord-AntiScam/scam-links`), refreshed once a day at 00:00 UTC. Confirmed
  scam links and mention spam get a timeout and a moderator alert with ban and
  release buttons. A message in the honeypot channel bans the sender outright.
  Separately, `/verify` runs a request-and-approve flow that grants the verified
  role.

  Neither a scam-domain count nor a currency count is quoted. Both live
  upstream, so a figure here would go stale the moment either list changes.
- **Quiz.** A WebSocket server plus HTTP endpoints that an external web client
  drives. The bot tracks the session and manages who may speak on a Stage
  channel: promote, suppress, reject, and inspect the queue.

### 3. Commands

Four tables, one per command group, transcribed from the `ApplicationCommand`
definitions so they cannot drift from what Discord shows:

| Group | Source |
| --- | --- |
| `/music` | `internal/handlers/music_commands.go` |
| `/currency` | `internal/handlers/currency_commands.go` |
| `/verify` | `internal/handlers/moderation_commands.go` |
| `/interactive-quiz` | `internal/handlers/quiz_commands.go` |

The `/music` table lists play, pause, resume, skip, stop, disconnect, clear,
list. The current file is missing `clear` and lists a `shownp` subcommand that
does not exist.

### 4. Running it

Numbered steps, not prose paragraphs:

1. Prerequisites: Go 1.27, Redis, ffmpeg, yt-dlp. External binaries are
   optional if the bundled versions in `./bin/` are used.
2. `cp .env.example .env`
3. Fill in the required variables (table follows).
4. `make run`

The required-variable table is long enough to matter: 11 values, 7 of them
Discord channel or role IDs. Each row gets a one-line note on where the value
comes from in the Discord client, so the table does not read as busywork.

Optional variables covered in a note rather than the table: `WISE_TOKEN` or
`CURRENCY_API_TOKEN` (whichever `CURRENCY_PROVIDER` selects), `YOUTUBE_API_KEY`,
`QUIZ_WS_PORT`, `PROXY_URL`, `YTDLP_PATH`.

`YOUTUBE_API_KEY` is not in `.env.example` but music search needs it. Without
it, `NewYoutubeProvider` returns an error and the bot starts with a nil music
service: every music command answers "unavailable" rather than refusing to
boot. The README must call this out, because the lazy-init behaviour is
deliberate but invisible.

### 5. Deployment

One short paragraph on each of the two paths that exist:

- `docker compose up` brings up Redis and the bot together.
- The `Dockerfile` builds a `linux/arm64` image that bundles ffmpeg and yt-dlp.
- `.github/workflows/deployment.yml` pushes that image to Docker Hub on pushes
  to `master`.

The deployment workflow still targets `master` while the default branch is
`development`. The README states this factually as what the workflow does. It
does not present it as recommended practice and does not propose a fix.

### 6. Development

The section the target audience most needs.

- **Layout.** One line per package under `internal/`: `music`, `currency`,
  `moderation`, `quiz`, `handlers`, `events`, `responses`, `bot`, `config`,
  `model`.
- **Make targets.** `make build`, `make run`, `make dev` (air live reload),
  `make fmt`, `make lint`, `make tidy`, `make docker-up`, `make docker-down`,
  `make clean`. There is no `make test` target; document `go test ./...`
  directly rather than inventing one.
- **CI expectations.** Three jobs run on every push and pull request: build and
  vet, test, lint. Lint is golangci-lint v2.15.0 configured by `.golangci.yml`.
- **Testing note.** One sentence: music provider tests are mostly offline, but
  `internal/music/providers/ytdlp_live_test.go` hits the network and is the file
  to skip when a run fails for unrelated reasons.
- **Branch flow.** One sentence: work off `development`, CI must pass.

No "Contributing" section. There is no `CONTRIBUTING.md`, no pull request
template, and no issue templates, so a section describing that process would
promise something that does not exist.

## Factual corrections

Every item below was verified against the tree during design. Each is either
removed or restated.

| Current claim | Verified reality |
| --- | --- |
| `/music shownp` shows a now-playing card | No such subcommand. Real set is play, pause, resume, skip, stop, disconnect, clear, list |
| `/music clear` absent | Exists: "Remove every queued track (current track keeps playing)" |
| No mention of quizzes | `/interactive-quiz start <stage channel>` and `/interactive-quiz stop` exist, backed by a WebSocket and HTTP server |
| Graphics stack: `gg`, `freetype`, `resize` | Zero imports anywhere in the tree. Removed with Lavalink; `assets/fonts/` remains but nothing reads it |
| `VoiceServerUpdate` event listener | Not registered. Only `MessageCreate` and `VoiceStateUpdate` remain |
| discordgo v0.28.1 | v0.29.0, plus a `replace` to `github.com/yeongaori/discordgo` |
| go-redis v9.7.3 | v9.23.0 |
| gocron v2.16.1 | v2.22.0, and the path is `go-co-op/gocron/v2` |
| godotenv v1.5.1 | Unchanged |
| Go 1.23.0+ with Go 1.24.1 toolchain | `go 1.27`, no toolchain pin |
| 7 environment variables listed | 18 consumed across the tree, 11 of them required by `config.Load` |
| "See REFACTORING_PLAN.md" | File deleted in `4e34b2f` |
| Last updated March 12, 2026 | Removed. A date in a README is a promise to maintain it |

## Out of scope

Found during design, deliberately not fixed as part of the README rewrite. Each
wants its own commit.

- `VERSION` is treated as required by `internal/config/loader.go` but is absent
  from `.env.example`, so following the example fails at startup with
  `missing required environment variables: VERSION`.
- `.env.example` omits `YOUTUBE_API_KEY`, which music search needs, despite
  listing the other optional variables.
- `Dockerfile` still carries a Lavalink fallback block referencing
  `lavalink_nodes.json`, which no longer exists.
- `.github/workflows/deployment.yml` targets `master` while the default branch
  is `development`.
- `internal/music/providers/ytdlp_live_test.go` runs on every `go test ./...`
  and will fail without network access.
- No `LICENSE` file, despite the module path being public on GitHub.

## Validation

The rewrite is done when:

- `FEATURES.md` is gone and `README.md` exists at the repo root.
- Every command table row matches a `Name:` field in the corresponding handler
  file.
- Every environment variable named in the README appears in either
  `internal/config/loader.go` or the relevant `utils.GetEnv` call.
- No emoji section headers, tick/cross bullets, or "Overview:"-style lead-ins
  remain.
- `git grep` for each removed claim (`shownp`, `freetype`, `gg v1.3.0`,
  `REFACTORING_PLAN`) returns nothing.
