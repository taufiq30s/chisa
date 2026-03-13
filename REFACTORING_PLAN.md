# Chisa Discord Bot - Feature Analysis & Refactoring Plan

## Project Overview
Chisa is a multi-featured Discord bot built with Go, providing music player, currency conversion, and moderation capabilities.

---

## Current Features

### 1. Music Player System
**Location:** `internal/music/`

**Features:**
- Play music from URLs or search queries (YouTube, Spotify)
- Queue management (add, view queue)
- Playback controls (play, pause, resume, skip, stop)
- Voice channel connection and disconnection
- Music search with pagination (next/previous page navigation)
- Interactive search result selection via buttons and dropdowns
- Track information display (title, artist, duration)
- Lavalink integration for audio streaming
- Multiple Lavalink node support with health checks
- Track event listeners (onTrackStart, onTrackEnd, onTrackStuck, onTrackException)
- Playlist support (add whole playlists to queue)
- Music card prototype display

**Commands:**
- `/music play <query>` - Play music
- `/music pause` - Pause playback
- `/music resume` - Resume playback
- `/music skip` - Skip current track
- `/music stop` - Stop player
- `/music disconnect` - Leave voice channel
- `/music shownp` - Show now playing card

### 2. Currency Management
**Location:** `internal/currency/`

**Features:**
- Real-time currency conversion
- Multiple provider support (Currency API, Wise)
- Autocomplete for currency selection
- Exchange rate display
- Wise money transfer simulation
- Fee calculation for Wise transfers
- Redis caching for rates
- Support for 160+ currencies

**Commands:**
- `/currency convert <amount> <from> <to>` - Convert currency
- `/currency simulate <mode> <amount> <source> <target>` - Simulate Wise transfer

**Sub-modules:**
- Currency API provider integration
- Wise API integration
- Rate caching and retrieval
- Currency validation

### 3. Moderation System
**Location:** `internal/moderation/`

**Features:**
- Scam link detection (using Discord Anti-Scam database)
- Automatic message scanning for scam URLs
- New member verification system
- Admin approval workflow for verification
- Automatic timeout for suspected scammers
- Ban functionality for confirmed scammers
- Message content analysis for @everyone/@here spam
- Daily scam dataset updates via cron job

**Commands:**
- `/verify` - Request verification

**Admin Actions:**
- Accept/reject verification requests via buttons
- Ban scammer button
- Remove timeout button

### 4. Backend Infrastructure

**Redis Integration:**
- Connection pool management (10 connections)
- Scam URL dataset storage
- Currency rate caching
- Session management

**Cron Jobs:**
- Daily scam dataset update (00:00:00)
- Automated maintenance tasks

**Event Handlers:**
- Message creation (scam detection)
- Voice server updates (music player)
- Voice state updates (user join/leave tracking)

**Response System:**
- Centralized error handling
- Message embed builder
- Interaction response utilities
- Ephemeral message support

**Utilities:**
- Environment variable management
- Logging system (debug, info, warning, error)
- Number formatting helpers

---

## Code Quality Issues Identified

### 1. **Architecture Issues**

#### a. Tight Coupling
- Direct dependencies between handlers and bot instance
- Music module tightly coupled to Discord session
- No dependency injection pattern
- Hard to test individual components

#### b. Missing Abstraction Layers
- Direct API calls in business logic
- No repository pattern for data access
- Redis operations scattered across modules

#### c. Monolithic Handler Registration
- All command registration in single init function
- Handler maps mixed with command definitions
- Difficult to add/remove features

### 2. **Code Organization Issues**

#### a. Mixed Responsibilities
- `bot.Bot` struct acts as service locator and coordinator
- Handlers contain both routing and business logic
- No clear boundary between layers

#### b. Inconsistent Naming
- File: `modetation_commands.go` (typo: should be "moderation")
- Mixed naming conventions (camelCase vs snake_case)
- Function names not always descriptive

#### c. Global Variables
- Heavy use of package-level variables in handlers
- Makes testing difficult
- Can cause race conditions

### 3. **Error Handling**

#### a. Inconsistent Patterns
- Some errors logged and ignored
- Others cause fatal exits
- Mixed use of fmt.Println and utils.ErrorLog

#### b. Poor Error Context
- Generic error messages
- Missing stack traces
- No error wrapping with context

### 4. **Configuration Management**

#### a. Hardcoded Values
- Channel IDs, role IDs in moderation code
- No configuration file
- Magic numbers scattered in code

#### b. Environment Variables
- No validation at startup
- Fatal errors only during use
- No default values

### 5. **Code Duplication**

- Response creation patterns repeated
- Error handling boilerplate duplicated
- Similar validation logic in multiple places

### 6. **Testing**

- No unit tests
- No integration tests
- No mocks or test fixtures

### 7. **Documentation**

- Minimal code comments
- No API documentation
- Missing architecture documentation

---

## Refactoring Plan

### Phase 1: Foundation & Structure (Week 1-2)

#### 1.1 Fix Critical Issues
- [ ] Rename `modetation_commands.go` to `moderation_commands.go`
- [ ] Fix typos in comments and variable names
- [ ] Add comprehensive error wrapping with context
- [ ] Standardize logging (remove fmt.Println for errors)

#### 1.2 Configuration Management
- [ ] Create `internal/config/` package
- [ ] Move all environment variables to config struct
- [ ] Add configuration validation at startup
- [ ] Create example `.env.example` file
- [ ] Extract hardcoded values (channel IDs, role IDs) to config
- [ ] Add config loader with sensible defaults

**Files to create:**
```
internal/config/
  ├── config.go          # Main config struct and loader
  ├── validation.go      # Config validation
  └── defaults.go        # Default values
```

#### 1.3 Project Structure Reorganization
- [ ] Create `pkg/` directory for reusable packages
- [ ] Separate domain logic from infrastructure
- [ ] Organize by feature (domain-driven design)

**New structure:**
```
pkg/
  ├── logger/           # Custom logger wrapper
  ├── validator/        # Input validation
  └── formatter/        # Format utilities

internal/
  ├── config/          # Configuration
  ├── domain/          # Business logic (entities, services)
  │   ├── music/
  │   ├── currency/
  │   └── moderation/
  ├── infrastructure/  # External dependencies
  │   ├── discord/     # Discord client wrapper
  │   ├── redis/       # Redis repository
  │   ├── lavalink/    # Lavalink client wrapper
  │   └── http/        # HTTP client utilities
  ├── application/     # Use cases/handlers
  │   ├── commands/    # Command handlers
  │   └── events/      # Event handlers
  └── interfaces/      # External interfaces
      └── api/         # API clients
```

### Phase 2: Dependency Injection & Interfaces (Week 3)

#### 2.1 Define Core Interfaces
- [ ] Create repository interfaces
- [ ] Create service interfaces
- [ ] Define provider contracts

**Files to create:**
```
internal/domain/
  ├── repository.go     # Repository interfaces
  └── service.go        # Service interfaces
```

#### 2.2 Implement Dependency Injection
- [ ] Refactor `bot.Bot` to use constructor injection
- [ ] Create service container/registry
- [ ] Remove global variables from handlers
- [ ] Use interfaces for all external dependencies

**Example:**
```go
// internal/application/container.go
type Container struct {
    logger       logger.Logger
    config       *config.Config
    redis        repository.RedisRepository
    musicService domain.MusicService
    currService  domain.CurrencyService
    modService   domain.ModerationService
}
```

#### 2.3 Repository Pattern
- [ ] Create Redis repository interface
- [ ] Implement scam dataset repository
- [ ] Implement currency cache repository
- [ ] Abstract data access layer

### Phase 3: Business Logic Separation (Week 4)

#### 3.1 Extract Services
- [ ] Create `MusicService` interface and implementation
- [ ] Create `CurrencyService` interface and implementation
- [ ] Create `ModerationService` interface and implementation
- [ ] Move business logic from handlers to services

#### 3.2 Clean Up Handlers
- [ ] Handlers should only handle Discord interactions
- [ ] Delegate business logic to services
- [ ] Standardize response patterns
- [ ] Remove duplication in error handling

**Example refactor:**
```go
// Before (in handler)
func musicCommandHandler(chisa *bot.Bot, i *discordgo.InteractionCreate) {
    // Parse, validate, execute, respond all in one place
}

// After
func (h *MusicHandler) HandlePlayCommand(i *discordgo.InteractionCreate) error {
    dto := h.parser.ParsePlayCommand(i)
    if err := h.validator.Validate(dto); err != nil {
        return h.responder.SendError(i, err)
    }
    
    result, err := h.musicService.Play(dto)
    if err != nil {
        return h.responder.SendError(i, err)
    }
    
    return h.responder.SendSuccess(i, result)
}
```

### Phase 4: Error Handling & Logging (Week 5)

#### 4.1 Custom Error Types
- [ ] Define domain-specific errors
- [ ] Create error hierarchy
- [ ] Add error codes for categorization
- [ ] Implement error wrapping with context

**Files to create:**
```
pkg/errors/
  ├── errors.go        # Base error types
  ├── music.go         # Music-specific errors
  ├── currency.go      # Currency-specific errors
  └── moderation.go    # Moderation-specific errors
```

#### 4.2 Centralized Logging
- [ ] Create logger interface
- [ ] Implement structured logging
- [ ] Add log levels and filtering
- [ ] Add context to log entries
- [ ] Remove fmt.Println from production code

### Phase 5: Testing Infrastructure (Week 6)

#### 5.1 Setup Testing Framework
- [ ] Add testing dependencies (testify, mockgen)
- [ ] Create test helpers and fixtures
- [ ] Setup mock generation
- [ ] Create integration test environment

#### 5.2 Unit Tests
- [ ] Test services (music, currency, moderation)
- [ ] Test validators
- [ ] Test formatters and utilities
- [ ] Test configuration loader

#### 5.3 Integration Tests
- [ ] Test Redis integration
- [ ] Test Discord interaction flow
- [ ] Test API clients
- [ ] Test end-to-end command flows

### Phase 6: Feature-Specific Improvements (Week 7-8)

#### 6.1 Music Module
- [ ] Add persistent queue (survive restarts)
- [ ] Implement playlist management
- [ ] Add repeat/shuffle functionality
- [ ] Improve search UX
- [ ] Add volume control
- [ ] Implement favorites/bookmarks
- [ ] Add music history

#### 6.2 Currency Module
- [ ] Add rate history tracking
- [ ] Implement rate alerts
- [ ] Support more currency providers
- [ ] Add currency trend visualization
- [ ] Cache optimization

#### 6.3 Moderation Module
- [ ] Improve scam detection algorithm
- [ ] Add whitelist/blacklist management
- [ ] Implement auto-mod rules
- [ ] Add moderation action logging
- [ ] Create admin dashboard commands

#### 6.4 Response System
- [ ] Create response builder pattern
- [ ] Add template system for embeds
- [ ] Implement pagination helper
- [ ] Add reaction role support

### Phase 7: Performance & Optimization (Week 9)

#### 7.1 Caching Strategy
- [ ] Implement intelligent cache invalidation
- [ ] Add cache warming for common requests
- [ ] Optimize Redis usage patterns
- [ ] Add in-memory cache layer (for read-heavy data)

#### 7.2 Concurrency
- [ ] Review goroutine usage
- [ ] Add proper context cancellation
- [ ] Implement graceful shutdown
- [ ] Add connection pooling where needed

#### 7.3 Resource Management
- [ ] Implement circuit breakers for external APIs
- [ ] Add rate limiting
- [ ] Optimize memory allocations
- [ ] Profile and optimize hot paths

### Phase 8: Documentation & DevEx (Week 10)

#### 8.1 Code Documentation
- [ ] Add package documentation
- [ ] Document all public APIs
- [ ] Add architecture decision records (ADRs)
- [ ] Create CONTRIBUTING.md

#### 8.2 User Documentation
- [ ] Create user guide
- [ ] Document all commands
- [ ] Add setup guide
- [ ] Create troubleshooting guide

#### 8.3 Developer Experience
- [ ] Add Makefile for common tasks
- [ ] Setup CI/CD pipeline
- [ ] Add pre-commit hooks
- [ ] Create docker-compose for local development
- [ ] Add database migrations (if needed)

### Phase 9: Additional Features (Week 11-12)

#### 9.1 Observability
- [ ] Add metrics collection (Prometheus)
- [ ] Implement health checks
- [ ] Add distributed tracing
- [ ] Create monitoring dashboard

#### 9.2 Security
- [ ] Input sanitization audit
- [ ] Add rate limiting per user
- [ ] Implement permission system
- [ ] Add audit logging for admin actions
- [ ] Secrets management review

#### 9.3 Reliability
- [ ] Implement retry logic with exponential backoff
- [ ] Add timeout configuration
- [ ] Improve error recovery
- [ ] Add fallback mechanisms for external services

---

## Refactoring Principles to Follow

### 1. **SOLID Principles**
- Single Responsibility: Each component should have one reason to change
- Open/Closed: Open for extension, closed for modification
- Liskov Substitution: Interfaces should be substitutable
- Interface Segregation: Small, focused interfaces
- Dependency Inversion: Depend on abstractions, not concretions

### 2. **Clean Code Practices**
- Meaningful names for variables, functions, types
- Functions should be small and do one thing
- Avoid deep nesting (max 3 levels)
- Use early returns to reduce complexity
- Keep files small and focused (<300 lines)

### 3. **Go Idioms**
- Accept interfaces, return structs
- Use context for cancellation and timeouts
- Handle errors explicitly
- Use defer for cleanup
- Avoid premature optimization

### 4. **Testing Strategy**
- Write tests before refactoring
- Maintain test coverage >80%
- Use table-driven tests
- Mock external dependencies
- Test behavior, not implementation

---

## Success Metrics

### Code Quality
- [ ] Zero use of global mutable state
- [ ] All external dependencies injected
- [ ] All errors properly wrapped with context
- [ ] Test coverage >80%
- [ ] No files >400 lines
- [ ] No functions >50 lines

### Maintainability
- [ ] New features can be added without modifying existing code
- [ ] Components can be tested in isolation
- [ ] Configuration changes don't require code changes
- [ ] Clear separation between layers

### Performance
- [ ] Response time <500ms for all commands
- [ ] Memory usage stable over time
- [ ] No goroutine leaks
- [ ] Graceful degradation under load

### Documentation
- [ ] All public APIs documented
- [ ] Architecture documented
- [ ] Setup guide available
- [ ] Contribution guide available

---

## Risk Mitigation

### High-Risk Changes
1. Handler registration refactoring
2. Bot initialization flow changes
3. Redis connection management changes

**Mitigation:**
- Create feature branches for risky changes
- Extensive testing before merge
- Incremental rollout
- Keep old code until new code proven stable

### Dependencies
- Lavalink availability
- Redis availability
- External API reliability

**Mitigation:**
- Implement circuit breakers
- Add fallback mechanisms
- Cache critical data
- Graceful degradation

---

## Next Steps

1. **Review this plan** with team/stakeholders
2. **Set up version control branch** for refactoring
3. **Create tracking board** (GitHub Projects/Jira)
4. **Start with Phase 1** (low-risk, high-impact changes)
5. **Iterate and adapt** based on learnings

---

## Notes

- This plan assumes ~12 weeks of dedicated work (1-2 developers)
- Phases can be parallelized where dependencies allow
- Priority should be given to foundation (Phases 1-4) before features
- Each phase should maintain backward compatibility
- Regular testing and validation after each phase
- Consider creating a "refactoring" branch to avoid disrupting main development

## Quick Wins (Can be done immediately)

1. Fix file naming typo (`modetation_commands.go`)
2. Add `.env.example` file
3. Standardize error logging (remove fmt.Println for errors)
4. Add basic comments to exported functions
5. Create `CONTRIBUTING.md` with code style guide
6. Extract magic numbers to constants
7. Add graceful shutdown handler improvement
