# Quick Start Refactoring Checklist

A prioritized checklist to begin refactoring immediately. Start here for maximum impact with minimum risk.

---

## 🔥 Critical Fixes (Do First - 1 day)

### File/Code Corrections
- [ ] **Rename file:** `internal/handlers/modetation_commands.go` → `moderation_commands.go`
  ```bash
  mv internal/handlers/modetation_commands.go internal/handlers/moderation_commands.go
  ```

- [ ] **Fix all imports** that reference the old filename
  - Search for: `modetation` 
  - Replace with: `moderation`

- [ ] **Add `.env.example`** file with all required environment variables
  ```env
  BOT_TOKEN=your_discord_bot_token_here
  AKASHIC_SERVER_ID=your_guild_id_here
  REDIS_URL=redis://localhost:6379
  WISE_TOKEN=your_wise_api_token_here
  CURRENCY_API_TOKEN=your_currency_api_token_here
  ```

### Logging Standardization
- [ ] Replace all `fmt.Println()` error cases with `utils.ErrorLog.Println()`
- [ ] Replace all `fmt.Printf()` error cases with `utils.ErrorLog.Printf()`
- [ ] Keep `fmt.Println()` ONLY for startup info messages that duplicate to logs

**Pattern to find:**
```go
// ❌ BEFORE
fmt.Println("Error:", err)

// ✅ AFTER  
utils.ErrorLog.Println("Error:", err)
```

---

## ⚡ Quick Wins (2-3 days)

### Extract Constants
- [ ] Create `internal/moderation/config.go` to hold hardcoded IDs
  ```go
  package moderation
  
  import "github.com/taufiq30s/chisa/utils"
  
  type Config struct {
      VerificationChannelID string
      ModeratorChannelID    string
      LogChannelID          string
      VerifiedRoleID        string
  }
  
  func LoadConfig() (*Config, error) {
      return &Config{
          VerificationChannelID: utils.GetEnv("VERIFICATION_CHANNEL_ID"),
          ModeratorChannelID:    utils.GetEnv("MODERATOR_CHANNEL_ID"),
          LogChannelID:          utils.GetEnv("LOG_CHANNEL_ID"),
          VerifiedRoleID:        utils.GetEnv("VERIFIED_ROLE_ID"),
      }, nil
  }
  ```

- [ ] Extract magic numbers to named constants
  ```go
  const (
      MinMessageLengthForScamCheck = 10
      RedisPoolSize                = 10
      MaxSearchResults             = 20
      SearchResultPageSize         = 5
      SearchSessionTimeout         = 5 * time.Minute
  )
  ```

### Error Wrapping
- [ ] Add `golang.org/x/xerrors` or use Go 1.13+ error wrapping
  ```go
  // ❌ BEFORE
  if err != nil {
      return err
  }
  
  // ✅ AFTER
  if err != nil {
      return fmt.Errorf("failed to update scam dataset: %w", err)
  }
  ```

- [ ] Update all error returns to include context

### Documentation
- [ ] Add package documentation comments to all packages
  ```go
  // Package music provides Discord music player functionality using Lavalink.
  // It supports playback control, queue management, and interactive search.
  package music
  ```

- [ ] Add function comments to all exported functions
  ```go
  // New creates a new MusicBot instance with the given session and guild.
  // The bot will use Lavalink for audio streaming and maintain its own queue.
  func New(s *discordgo.Session, botId string, guildId string) *MusicBot {
  ```

---

## 🏗️ Foundation (Week 1)

### Configuration Package
- [ ] Create `internal/config/config.go`
- [ ] Move all environment variable loading to config
- [ ] Add validation for required fields
- [ ] Load config once at startup, pass around as dependency

**File structure:**
```
internal/config/
├── config.go       # Main config struct and loader
├── validation.go   # Validate() method
└── defaults.go     # Default values for optional configs
```

**Example implementation:**
```go
package config

import (
    "fmt"
    "github.com/joho/godotenv"
    "github.com/taufiq30s/chisa/utils"
)

type Config struct {
    // Bot
    BotToken  string
    GuildID   string
    
    // Redis
    RedisURL string
    
    // Currency
    WiseToken        string
    CurrencyAPIToken string
    
    // Moderation
    VerificationChannelID string
    ModeratorChannelID    string
    LogChannelID          string
    VerifiedRoleID        string
}

func Load() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        return nil, fmt.Errorf("failed to load .env: %w", err)
    }
    
    cfg := &Config{}
    var err error
    
    cfg.BotToken, err = utils.GetEnv("BOT_TOKEN")
    if err != nil {
        return nil, err
    }
    
    // ... load other fields
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return cfg, nil
}

func (c *Config) Validate() error {
    if c.BotToken == "" {
        return fmt.Errorf("BOT_TOKEN is required")
    }
    // ... validate other required fields
    return nil
}
```

### Update main.go
- [ ] Use config package instead of direct env access
- [ ] Pass config to bot initialization
- [ ] Validate config before starting any services

---

## 🧪 Testing Setup (Week 1-2)

### Test Infrastructure
- [ ] Create `Makefile` for common tasks
  ```makefile
  .PHONY: test build run clean lint
  
  test:
      go test -v -race -coverprofile=coverage.out ./...
  
  coverage:
      go tool cover -html=coverage.out
  
  build:
      go build -o bin/chisa cmd/main.go
  
  run:
      go run cmd/main.go
  
  lint:
      golangci-lint run
  
  clean:
      rm -rf bin/ coverage.out
  ```

- [ ] Add test for configuration loading
  ```go
  // internal/config/config_test.go
  package config_test
  
  import (
      "testing"
      "github.com/taufiq30s/chisa/internal/config"
  )
  
  func TestValidate(t *testing.T) {
      tests := []struct {
          name    string
          config  config.Config
          wantErr bool
      }{
          {
              name: "valid config",
              config: config.Config{
                  BotToken: "test_token",
                  GuildID:  "123456",
                  RedisURL: "redis://localhost",
              },
              wantErr: false,
          },
          {
              name: "missing bot token",
              config: config.Config{
                  GuildID:  "123456",
                  RedisURL: "redis://localhost",
              },
              wantErr: true,
          },
      }
      
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              err := tt.config.Validate()
              if (err != nil) != tt.wantErr {
                  t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
              }
          })
      }
  }
  ```

- [ ] Add test for utils package
- [ ] Setup test fixtures directory

---

## 🎯 Handler Refactoring (Week 2)

### Command Handler Cleanup
- [ ] Extract command definitions from handler logic
  ```go
  // ❌ BEFORE: Everything mixed together
  var musicCommandHandler = func(chisa *bot.Bot, i *discordgo.InteractionCreate) {
      // Heavy logic here
  }
  var musicCommands = []*discordgo.ApplicationCommand{ /* ... */ }
  
  // ✅ AFTER: Separated
  // internal/handlers/music/commands.go
  func GetCommands() []*discordgo.ApplicationCommand {
      return musicCommands
  }
  
  // internal/handlers/music/handler.go
  type Handler struct {
      musicService music.Service
  }
  
  func (h *Handler) Handle(i *discordgo.InteractionCreate) error {
      // Delegate to service
  }
  ```

- [ ] Create handler structs instead of functions
- [ ] Move business logic to service layer
- [ ] Handlers should only parse input and format output

### Response Standardization
- [ ] Create response builder
  ```go
  // internal/responses/builder.go
  type Builder struct {
      session *discordgo.Session
  }
  
  func (b *Builder) Success(title, description string) *discordgo.MessageEmbed {
      return CreateMessageEmbed(
          b.session,
          title,
          description,
          "",
          SetColor("0bdd47"),
      )
  }
  
  func (b *Builder) Error(title string, err error) *discordgo.MessageEmbed {
      return CreateMessageEmbed(
          b.session,
          title,
          err.Error(),
          "",
          SetColor("df0000"),
      )
  }
  ```

---

## 📊 Dependency Injection (Week 3)

### Service Interfaces
- [ ] Define service interfaces
  ```go
  // internal/music/service.go
  type Service interface {
      Play(ctx context.Context, query string) (*Track, error)
      Pause(ctx context.Context) error
      Resume(ctx context.Context) error
      Skip(ctx context.Context) (*Track, error)
      Stop(ctx context.Context) error
  }
  ```

### Container/Registry
- [ ] Create service container
  ```go
  // internal/app/container.go
  type Container struct {
      config      *config.Config
      redis       *redis.Client
      discord     *discordgo.Session
      
      musicSvc    music.Service
      currencySvc currency.Service
      moderSvc    moderation.Service
  }
  
  func NewContainer(cfg *config.Config) (*Container, error) {
      // Initialize all services
  }
  ```

- [ ] Update main.go to use container
- [ ] Pass services through container instead of global bot

---

## ✅ Validation Points

After each section, verify:

### After Critical Fixes:
- [ ] Project compiles without errors
- [ ] All tests pass (if any)
- [ ] Bot starts successfully
- [ ] No console spam from logging changes

### After Quick Wins:
- [ ] No hardcoded values in main code paths
- [ ] All errors have context
- [ ] Documentation visible in IDE

### After Foundation:
- [ ] Config validates at startup
- [ ] Invalid config prevents bot start
- [ ] All env vars documented in .env.example

### After Testing Setup:
- [ ] `make test` runs successfully
- [ ] Coverage report generates
- [ ] At least one test passes

### After Handler Refactoring:
- [ ] All commands still work
- [ ] Response formats unchanged
- [ ] No regression in functionality

### After Dependency Injection:
- [ ] Services are mockable
- [ ] No global mutable state
- [ ] Clear dependency tree

---

## 🚨 Red Flags to Watch For

Stop and review if you see:
- ⛔ Existing features stop working
- ⛔ Tests start failing that passed before
- ⛔ Memory usage spikes
- ⛔ Compile times increase significantly
- ⛔ More code added than removed (refactoring should simplify)

---

## 📈 Progress Tracking

### Daily Checklist
- [ ] Morning: Review yesterday's changes
- [ ] Run tests before starting
- [ ] Make small, incremental commits
- [ ] Run tests after each logical change
- [ ] Evening: Document any issues or learnings

### Weekly Review
- [ ] Review test coverage
- [ ] Check performance metrics
- [ ] Update refactoring plan based on progress
- [ ] Identify blockers

---

## 🎓 Learning Resources

### Go Best Practices
- Effective Go: https://go.dev/doc/effective_go
- Go Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments

### Refactoring Patterns
- Refactoring Guru: https://refactoring.guru/
- Martin Fowler's Refactoring Catalog: https://refactoring.com/catalog/

### Testing in Go
- Table-Driven Tests: https://go.dev/wiki/TableDrivenTests
- Testify Package: https://github.com/stretchr/testify

---

**Remember:** 
- Small, incremental changes
- Keep the bot working after each change
- Test before and after
- Commit frequently with clear messages
- Ask for help when stuck

Good luck! 🚀
