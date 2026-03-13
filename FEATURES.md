# Chisa Discord Bot - Feature List

## Overview
Chisa is a multi-purpose Discord bot providing music playback, currency conversion, and server moderation capabilities.

---

## 🎵 Music Features

### Playback Commands
| Command | Description |
|---------|-------------|
| `/music play <query>` | Play music from URL or search query (YouTube, Spotify) |
| `/music pause` | Pause the currently playing track |
| `/music resume` | Resume playback |
| `/music skip` | Skip to the next track in queue |
| `/music stop` | Stop the player and clear queue |
| `/music disconnect` | Disconnect bot from voice channel |
| `/music shownp` | Show now playing card (prototype feature) |

### Music Capabilities
- ✅ YouTube URL support (various formats: watch, share, embed)
- ✅ Spotify track URL support
- ✅ Search by song name (uses YouTube search)
- ✅ Queue management (automatic queue progression)
- ✅ Interactive search results with pagination
- ✅ Playlist support (add entire playlists)
- ✅ Track information display (title, artist, duration)
- ✅ Multiple Lavalink node support for reliability
- ✅ Auto-reconnection if node fails
- ✅ Voice channel auto-join when playing

### Technical Features
- Lavalink integration for high-quality audio
- Disgolink library for Lavalink communication
- Multiple node load balancing
- Track event handling (start, end, stuck, exception)
- WebSocket connection management

---

## 💰 Currency Features

### Conversion Commands
| Command | Description |
|---------|-------------|
| `/currency convert <amount> <from> <to>` | Convert between currencies with real-time rates |
| `/currency simulate <mode> <amount> <source> <target>` | Simulate Wise money transfer with fees |

### Currency Capabilities
- ✅ Real-time exchange rates
- ✅ 160+ supported currencies
- ✅ Autocomplete for currency selection
- ✅ Rate caching (via Redis)
- ✅ Last update timestamp display
- ✅ Wise transfer simulation with:
  - Accurate fee calculation
  - Source/target amount modes
  - Rate display
  - Total cost breakdown
  - Recipient amount calculation

### Supported Providers
- **Currency API** - General currency conversion
- **Wise** - Money transfer simulation and rates

---

## 🛡️ Moderation Features

### Verification System
| Command | Description |
|---------|-------------|
| `/verify` | Request verification from moderators |

**Workflow:**
1. New member runs `/verify` in verification channel
2. Request sent to moderator channel
3. Moderators can Accept ✅ or Reject ❎
4. On accept: User gets verified role and access to server
5. On reject: Request dismissed

### Auto-Moderation
- ✅ **Scam Link Detection**
  - Automatic scanning of all messages
  - Uses Discord Anti-Scam Project database (170k+ known scam domains)
  - Daily automatic database updates (00:00 UTC)
  - Real-time checking against scam URL patterns

- ✅ **Automated Actions**
  - Suspicious messages (@everyone/@here spam): Timeout user, alert mods
  - Confirmed scam links: Immediate timeout, alert mods
  - Interactive mod panel with ban/release buttons

### Moderator Actions (via buttons)
- Ban detected scammer
- Remove timeout from false positive

### Protection Features
- URL extraction and validation
- Regex-based pattern matching
- @everyone/@here mention detection
- Prevents already-verified users from re-verifying

---

## 🔧 Backend Infrastructure

### Database & Caching
- **Redis** integration for:
  - Scam URL dataset storage
  - Currency rate caching
  - Session management
  - Connection pooling (10 connections)

### Scheduled Jobs
- Daily scam dataset update (runs at 00:00 UTC)
- Automatic maintenance via cron scheduler

### Event Listeners
- `MessageCreate` - Scam detection on every message
- `VoiceServerUpdate` - Music player voice updates
- `VoiceStateUpdate` - User voice channel changes

### Utilities
- Structured logging (debug, info, warning, error)
- Number formatting with grouping (1,000,000.00)
- Environment variable management
- Graceful shutdown handling

### Response System
- Centralized embed builder
- Error response templates
- Deferred responses for slow operations
- Ephemeral messages for privacy
- Button and select menu support

---

## 📊 Technical Stack

### Core Dependencies
- **discordgo** (v0.28.1) - Discord API wrapper
- **disgolink** (v3.0.4) - Lavalink client (custom fork)
- **go-redis** (v9.7.3) - Redis client
- **gocron** (v2.16.1) - Cron job scheduler
- **godotenv** (v1.5.1) - Environment configuration

### Graphics & Rendering
- **gg** (v1.3.0) - 2D graphics library
- **freetype** - Font rendering
- **resize** - Image resizing for music cards

### Language
- Go 1.23.0+ (with Go 1.24.1 toolchain)

---

## 🎯 Current Limitations

### Music
- ❌ No volume control
- ❌ No queue reordering/removal
- ❌ No loop/repeat functionality
- ❌ No shuffle mode
- ❌ Queue doesn't persist across bot restarts
- ❌ Limited music platform support (YT, Spotify only)

### Currency
- ❌ No historical rate tracking
- ❌ No rate alerts/notifications
- ❌ Single provider at a time
- ❌ No visual graphs/charts

### Moderation
- ❌ No custom auto-mod rules
- ❌ No word filter
- ❌ No raid protection
- ❌ No mod action history/audit log
- ❌ No slowmode management
- ❌ No user warnings system

### General
- ❌ No persistent configuration (uses env vars only)
- ❌ No user preferences
- ❌ No analytics/usage stats
- ❌ No admin dashboard
- ❌ No backup/restore functionality

---

## 🚀 Deployment

### Requirements
- Go 1.23+
- Redis server
- Lavalink server(s)
- Discord bot token
- Currency API token or Wise API token

### Environment Variables
```env
BOT_TOKEN=                 # Discord bot token
AKASHIC_SERVER_ID=         # Guild ID for registration
REDIS_URL=                 # Redis connection string
WISE_TOKEN=                # Wise API token (for currency)
CURRENCY_API_TOKEN=        # Alternative currency provider
```

### Running
```bash
go run cmd/main.go
```

---

## 📝 Feature Categorization

### Production-Ready ✅
- Music playback
- Currency conversion
- Scam detection
- Verification system

### Beta/Experimental 🧪
- Music card display (prototype)
- Wise transfer simulation

### Planned/TODO 📋
- See REFACTORING_PLAN.md for detailed roadmap

---

## 🤝 Contributing

Features are organized in a modular structure:
- `internal/music/` - Music player
- `internal/currency/` - Currency conversion
- `internal/moderation/` - Moderation & verification
- `internal/handlers/` - Command handlers
- `internal/events/` - Discord event handlers
- `internal/responses/` - Response utilities

Each module is relatively self-contained, making it easy to add new features or modify existing ones.

---

*Last updated: March 12, 2026*
