package quiz

import (
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

// QuizService is the interface the rest of the bot uses to interact with the
// quiz feature. All methods are safe for concurrent use.
type QuizService interface {
	// Start activates a new quiz session on the given stage channel.
	// Returns the WS session ID and WS endpoint URL, or an error.
	Start(s *discordgo.Session, guildID, channelID string) (sessionID, wsEndpoint string, err error)

	// Stop terminates the active session and disconnects all WS clients.
	// Returns how many users were promoted during the session.
	Stop() (totalPromoted int, channelID string, err error)

	// IsActive reports whether a session is currently running.
	IsActive() bool

	// HandleVoiceStateUpdate processes a Discord VOICE_STATE_UPDATE event.
	HandleVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate)

	// GetHub exposes the Hub so the HTTP server can call BroadcastQueueUpdate.
	GetHub() *Hub

	// GetSession exposes the Session so the HTTP handlers can read/mutate it.
	GetSession() *Session

	// GetSessionID returns the current active ws session UUID (empty if inactive).
	GetSessionID() string
}

// QuizManager is the concrete implementation of QuizService.
type QuizManager struct {
	session   *Session
	hub       *Hub
	sessionID string // UUID for the active WS session (empty when inactive)
	wsPort    string
}

// New creates a QuizManager. Call StartServer separately to begin accepting
// WebSocket connections.
func New(wsPort string) *QuizManager {
	hub := NewHub()
	go hub.Run()
	return &QuizManager{
		session: newSession(),
		hub:     hub,
		wsPort:  wsPort,
	}
}

func (q *QuizManager) IsActive() bool {
	q.session.mu.Lock()
	defer q.session.mu.Unlock()
	return q.session.Active
}

func (q *QuizManager) GetHub() *Hub         { return q.hub }
func (q *QuizManager) GetSession() *Session { return q.session }
func (q *QuizManager) GetSessionID() string { return q.sessionID }

func (q *QuizManager) Start(s *discordgo.Session, guildID, channelID string) (string, string, error) {
	// Generate a fresh session ID for each quiz session.
	q.sessionID = uuid.NewString()

	q.session.mu.Lock()
	q.session.Active = true
	q.session.GuildID = guildID
	q.session.ChannelID = channelID
	q.session.mu.Unlock()

	return q.sessionID, buildWSEndpoint(q.wsPort, q.sessionID), nil
}

func (q *QuizManager) Stop() (int, string, error) {
	q.session.mu.Lock()
	total := q.session.TotalPromoted
	channelID := q.session.ChannelID
	q.session.Active = false
	q.session.mu.Unlock()

	q.hub.CloseSession(q.sessionID, wsCloseSessionEnded, "Session ended by moderator")
	q.session.Reset()
	return total, channelID, nil
}

func (q *QuizManager) HandleVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	q.session.mu.Lock()
	active := q.session.Active
	guildID := q.session.GuildID
	channelID := q.session.ChannelID
	q.session.mu.Unlock()

	if !active {
		return
	}
	if e.GuildID != guildID {
		return
	}

	// User left the stage channel or disconnected entirely
	if e.ChannelID != channelID {
		q.session.RemoveFromQueue(e.UserID)
		q.hub.BroadcastQueueUpdate(q.sessionID, q.session.Snapshot())
		return
	}

	// User raised hand (RequestToSpeakTimestamp set)
	if e.RequestToSpeakTimestamp != nil {
		member, err := s.GuildMember(e.GuildID, e.UserID)
		username := e.UserID
		avatar := ""
		if err == nil && member.User != nil {
			username = member.User.Username
			avatar = member.User.AvatarURL("128")
		}
		q.session.AddToQueue(QueueEntry{
			UserID:    e.UserID,
			Username:  username,
			Avatar:    avatar,
			Timestamp: *e.RequestToSpeakTimestamp,
		})
	} else {
		q.session.RemoveFromQueue(e.UserID)
	}

	q.hub.BroadcastQueueUpdate(q.sessionID, q.session.Snapshot())
}

func buildWSEndpoint(port, sessionID string) string {
	return "ws://localhost:" + port + "/ws?session=" + sessionID
}
