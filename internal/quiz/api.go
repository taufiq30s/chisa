package quiz

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/utils"
)

// discordSession is set once when InitializeQuizService wires the Bot's session.
var discordSession *discordgo.Session

// SetDiscordSession wires the Discord session used by HTTP handlers.
func SetDiscordSession(s *discordgo.Session) {
	discordSession = s
}

// ---- request / response types -----------------------------------------------

type promoteRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

type suppressRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

type promoteResponse struct {
	OK             bool   `json:"ok"`
	PromotedUserID string `json:"promoted_user_id"`
}

type suppressResponse struct {
	OK bool `json:"ok"`
}

type queueResponse struct {
	SessionID string       `json:"session_id"`
	ChannelID string       `json:"channel_id"`
	Queue     []QueueEntry `json:"queue"`
}

// ---- auth middleware ---------------------------------------------------------

func authMiddleware(quiz QuizService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(authHeader)
		if token == "" || token != quiz.GetSessionID() || !quiz.IsActive() {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// ---- handler factories -------------------------------------------------------

func makePromoteHandler(quiz QuizService) http.HandlerFunc {
	return authMiddleware(quiz, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req promoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.UserID == "" {
			jsonError(w, "user_id is required", http.StatusBadRequest)
			return
		}

		sess := quiz.GetSession()
		sess.mu.Lock()
		channelID := sess.ChannelID
		guildID := sess.GuildID
		sess.mu.Unlock()

		// Promote user in Discord Stage (suppress: false)
		if discordSession != nil {
			err := discordSession.GuildMemberMove(guildID, req.UserID, &channelID)
			if err != nil {
				utils.WarningLog.Printf("Quiz promote: GuildMemberMove err: %v\n", err)
			}
			// Unsuppress (make speaker)
			_, err = discordSession.Request("PATCH",
				fmt.Sprintf("/guilds/%s/voice-states/%s", guildID, req.UserID),
				map[string]interface{}{"suppress": false, "channel_id": channelID},
			)
			if err != nil {
				utils.ErrorLog.Printf("Quiz promote: unsuppress err: %v\n", err)
			}
		}

		quiz.GetSession().MarkPromoted(req.UserID)
		quiz.GetHub().BroadcastQueueUpdate(quiz.GetSessionID(), quiz.GetSession().Snapshot())

		jsonOK(w, promoteResponse{OK: true, PromotedUserID: req.UserID})
	})
}

func makeSuppressHandler(quiz QuizService) http.HandlerFunc {
	return authMiddleware(quiz, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req suppressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.UserID == "" {
			jsonError(w, "user_id is required", http.StatusBadRequest)
			return
		}

		sess := quiz.GetSession()
		sess.mu.Lock()
		channelID := sess.ChannelID
		guildID := sess.GuildID
		sess.mu.Unlock()

		// Move user back to audience (suppress: true)
		if discordSession != nil {
			_, err := discordSession.Request("PATCH",
				fmt.Sprintf("/guilds/%s/voice-states/%s", guildID, req.UserID),
				map[string]interface{}{"suppress": true, "channel_id": channelID},
			)
			if err != nil {
				utils.ErrorLog.Printf("Quiz suppress: err: %v\n", err)
			}
		}

		jsonOK(w, suppressResponse{OK: true})
	})
}

func makeQueueHandler(quiz QuizService) http.HandlerFunc {
	return authMiddleware(quiz, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sess := quiz.GetSession()
		sess.mu.Lock()
		channelID := sess.ChannelID
		sess.mu.Unlock()

		jsonOK(w, queueResponse{
			SessionID: quiz.GetSessionID(),
			ChannelID: channelID,
			Queue:     quiz.GetSession().Snapshot(),
		})
	})
}
