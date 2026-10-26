package music

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

func (m *MusicBot) ensureVoiceState(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Error while getting voice state",
			Description: err.Error(),
		}).ExecuteDefer()
		return ""
	}
	if voiceState == nil || voiceState.ChannelID == "" {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Not in a voice channel",
			Description: "Join a voice channel first, then run the command again.",
		}).ExecuteDefer()
		return ""
	}
	return voiceState.ChannelID
}

// checkVoicePermissions verifies the bot can view, connect to and speak in the
// target voice channel. Administrator overrides all channel-level checks.
func (m *MusicBot) checkVoicePermissions(s *discordgo.Session, channelID string) error {
	perms, err := s.State.UserChannelPermissions(s.State.User.ID, channelID)
	if err != nil {
		return err
	}
	if perms&discordgo.PermissionAdministrator != 0 {
		return nil
	}

	var missing []string
	if perms&discordgo.PermissionViewChannel == 0 {
		missing = append(missing, "View Channel")
	}
	if perms&discordgo.PermissionVoiceConnect == 0 {
		missing = append(missing, "Connect")
	}
	if perms&discordgo.PermissionVoiceSpeak == 0 {
		missing = append(missing, "Speak")
	}
	if len(missing) > 0 {
		return fmt.Errorf("Missing permission: %s", strings.Join(missing, ", "))
	}
	return nil
}

// currentStreamer and setStreamer are the ONLY safe ways to touch m.streamer.
// The pointer is written when joining/leaving voice (interaction goroutines)
// and read by the player goroutine, so touching the field directly is a data
// race - and a race here is not cosmetic: a torn read passes the nil check and
// then panics on Play, which kills the entire bot process.
func (m *MusicBot) currentStreamer() *Streamer {
	m.streamerMu.RLock()
	defer m.streamerMu.RUnlock()
	return m.streamer
}

func (m *MusicBot) setStreamer(s *Streamer) {
	m.streamerMu.Lock()
	m.streamer = s
	m.streamerMu.Unlock()
}

// takeStreamer clears the field and returns what was there, in one critical
// section. Used by disconnect so two concurrent disconnects cannot both see the
// same Streamer and call Disconnect twice.
func (m *MusicBot) takeStreamer() *Streamer {
	m.streamerMu.Lock()
	defer m.streamerMu.Unlock()
	old := m.streamer
	m.streamer = nil
	return old
}

// verifyVoicePermissions is the cheap up-front half of joining: it only reads
// permissions and never touches the voice gateway. Load calls it before it
// resolves anything, so a missing Speak permission is reported immediately
// instead of after the user has already picked from 20 search results.
func (m *MusicBot) verifyVoicePermissions(s *discordgo.Session, channelID string) error {
	if channelID == "" {
		return fmt.Errorf("no voice channel to join")
	}
	return m.checkVoicePermissions(s, channelID)
}

// ensureVoiceJoined connects the bot to channelID if it is not connected yet.
// It is the expensive half and is only called once a track is actually about to
// be queued - joining on /music play alone would leave the bot idling in the
// channel during a search the user may cancel or never finish.
//
// It returns the error instead of replying so callers can pick the right reply
// path: a slash command still owns its deferred reply, while the search picker
// was already answered and must fall back to a channel message.
func (m *MusicBot) ensureVoiceJoined(s *discordgo.Session, channelID string) error {
	if m.currentStreamer() != nil {
		return nil
	}
	if err := m.verifyVoicePermissions(s, channelID); err != nil {
		return err
	}

	vc, err := connectVoice(s, m.guildID.String(), channelID)
	if err != nil {
		return err
	}
	proxyURL, _ := utils.GetEnv("PROXY_URL")
	m.setStreamer(newStreamer(vc, proxyURL))
	return nil
}

// connectVoice joins the voice channel, retrying once on a transport failure.
// discordgo's onVoiceServerUpdate logs a failed voice.open and gives up, so the
// first transient DNS/IPv6 hiccup is fatal for that attempt and
// waitUntilConnected reports the misleading "timeout waiting for voice". The
// retry re-sends op4, which makes Discord emit a fresh VoiceServerUpdate and
// with it a new voice.open.
func connectVoice(s *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, false)
	if err == nil || !isTransientVoiceError(err) {
		return vc, err
	}
	utils.ErrorLog.Println("Voice join failed, retrying once:", err)
	time.Sleep(2 * time.Second)
	return s.ChannelVoiceJoin(guildID, channelID, false, false)
}

// isTransientVoiceError reports whether a voice join error looks like the
// network dropped the handshake rather than the channel/permissions being
// wrong. Retrying the latter is pointless and just doubles the latency.
func isTransientVoiceError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, s := range []string{
		"timeout waiting for voice",
		"network is unreachable",
		"server misbehaving",
		"i/o timeout",
		"connection reset",
		"connection refused",
	} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// disconnectFromVoiceChannel leaves the voice channel and forgets the streamer.
// Callers are responsible for stopping playback first: a running playerLoop
// holds its own Streamer reference, so clearing this field mid-play only makes
// the *next* track fail with "Streamer is not initialized".
func (m *MusicBot) disconnectFromVoiceChannel() {
	streamer := m.takeStreamer()
	if streamer != nil && streamer.voiceConnection != nil {
		_ = streamer.voiceConnection.Disconnect()
	}
}
