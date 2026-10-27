package music

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

// trackTitleLimit keeps one queue line from eating the whole embed. Discord
// caps an embed description at 4096 characters; with CJK titles (1 character
// each, and yt-dlp happily returns very long ones) ten untruncated lines can
// blow the limit even at only ten tracks per page.
const trackTitleLimit = 60

// queueListState is the pagination state for one user's /music list panel.
// Modelled on searchResultState, with one important difference: the tracks are
// NOT copied in. The queue is mutable and the player goroutine rewrites it
// constantly (dequeues, clears), so a copy would show stale tracks after a
// skip. Only the page number is kept and the rows are rebuilt from the live
// queue on every render.
type queueListState struct {
	page      int
	timestamp time.Time
	channelId string
	messageId string
}

// ListQueue renders the queue for the invoking user and remembers the message
// so the Prev/Next buttons can edit it in place.
func (m *MusicBot) ListQueue(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.queueListMu.Lock()
	m.queueLists[i.Member.User.ID] = &queueListState{
		page:      InitialPage,
		timestamp: time.Now(),
		channelId: i.ChannelID,
	}
	m.queueListMu.Unlock()

	m.renderQueueList(s, i, InitialPage, false)
}

// ListQueueNextPage and ListQueuePreviousPage are the button handlers.
func (m *MusicBot) ListQueueNextPage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.moveQueuePage(s, i, true)
}

func (m *MusicBot) ListQueuePreviousPage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.moveQueuePage(s, i, false)
}

func (m *MusicBot) moveQueuePage(s *discordgo.Session, i *discordgo.InteractionCreate, isNext bool) {
	m.queueListMu.Lock()
	state := m.queueLists[i.Member.User.ID]
	if state == nil {
		m.queueListMu.Unlock()
		// Execute() not ExecuteDefer(): a component interaction has no deferred
		// reply to edit, and leaving it unanswered shows "This interaction
		// failed" to the user.
		_ = responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Title:       "Session Expired",
			Description: "This queue panel expired, run /music list again.",
			Feature:     m.featureName,
		}).Execute()
		return
	}
	// Clamp against the CURRENT queue length, not the length at open time, so a
	// page that no longer exists (tracks were cleared) falls back instead of
	// rendering blank.
	maxPage := m.queuePageCount()
	if maxPage < InitialPage {
		maxPage = InitialPage
	}
	if isNext && state.page < maxPage {
		state.page++
	} else if !isNext && state.page > InitialPage {
		state.page--
	}
	page := state.page
	m.queueListMu.Unlock()

	m.renderQueueList(s, i, page, true)
}

// queuePageCount returns how many pages the live queue occupies. Caller must
// NOT hold queueMu or queueListMu (it takes queueMu itself).
func (m *MusicBot) queuePageCount() int {
	m.queueMu.Lock()
	length := len(m.queue)
	m.queueMu.Unlock()
	return (length + QueueResultsPerPage - 1) / QueueResultsPerPage
}

func (m *MusicBot) renderQueueList(s *discordgo.Session, i *discordgo.InteractionCreate, page int, isUpdate bool) {
	// One snapshot under the lock: the player goroutine can dequeue at any
	// moment, so reading m.queue fields outside the lock is a data race.
	m.queueMu.Lock()
	pending := make([]trackState, len(m.queue))
	copy(pending, m.queue)
	var current *trackState
	if m.currentTrack != nil {
		snapshot := *m.currentTrack
		current = &snapshot
	}
	m.queueMu.Unlock()

	totalLength := 0
	for _, state := range pending {
		totalLength += state.track.Length
	}
	if current != nil {
		totalLength += current.track.Length
	}

	totalPages := (len(pending) + QueueResultsPerPage - 1) / QueueResultsPerPage
	if totalPages < InitialPage {
		totalPages = InitialPage
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - InitialPage) * QueueResultsPerPage
	end := min(start+QueueResultsPerPage, len(pending))

	description := "The queue is empty."
	if len(pending) > 0 {
		var b strings.Builder
		for idx, state := range pending[start:end] {
			fmt.Fprintf(&b, "%d. %s\n", start+idx+1, formatQueueLine(state))
		}
		description = b.String()
	}

	nowPlaying := "Nothing"
	if current != nil {
		nowPlaying = formatQueueLine(*current)
	}

	embed := responses.CreateMessageEmbed(
		s, "Queue",
		description,
		m.featureName,
		responses.SetColor("5e11d9"),
		responses.SetFields([]*discordgo.MessageEmbedField{
			{Name: "Now Playing", Value: truncateTitle(nowPlaying, 1024), Inline: false},
			{Name: "Page", Value: fmt.Sprintf("%d/%d", page, totalPages), Inline: true},
			{Name: "Queued tracks", Value: fmt.Sprintf("%d", len(pending)), Inline: true},
			{Name: "Total duration", Value: formatDuration(totalLength), Inline: true},
		}),
	)

	response := responses.InteractionResponse(s, i.Interaction)
	if isUpdate {
		response.SetResponseTypeAsUpdate()
	}

	// No AddNewRow() here: SendDefer/Send flush the pending buttons into one
	// row themselves. Calling AddNewRow() before the buttons built would push
	// an empty ActionsRow, and Discord rejects the whole response for it.
	response.WithEmbed(embed).WithButton(
		"Previous", discordgo.PrimaryButton,
		fmt.Sprintf("queue-previous-%s", i.Interaction.ID), page == InitialPage,
	).WithButton(
		"Next", discordgo.PrimaryButton,
		fmt.Sprintf("queue-next-%s", i.ID), page >= totalPages,
	)

	// isUpdate edits the existing panel; otherwise answer the deferral that
	// musicCommandHandler already created. Never Defer() again here - a second
	// acknowledgement fails with 40060.
	var err error
	if isUpdate {
		err = response.Send()
	} else {
		err = response.SendDefer()
	}
	if err != nil {
		utils.ErrorLog.Println("Failed to send queue list:", err)
		return
	}

	// Fetch the message ID so the sweeper can delete an abandoned panel.
	msg, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		utils.ErrorLog.Println("Failed to get queue list response:", err)
		return
	}
	m.queueListMu.Lock()
	if state := m.queueLists[i.Member.User.ID]; state != nil {
		state.messageId = msg.ID
	}
	m.queueListMu.Unlock()
}

// formatQueueLine renders one row: title, duration, and who queued it.
func formatQueueLine(state trackState) string {
	addedBy := "unknown"
	if state.addedBy != nil {
		addedBy = *state.addedBy
	}
	return fmt.Sprintf("**%s** · `%s` · @%s",
		truncateTitle(state.track.Title, trackTitleLimit),
		formatDuration(state.track.Length),
		addedBy,
	)
}

// truncateTitle shortens on rune boundaries so a multibyte character is never
// split in half (which Discord renders as a replacement glyph).
func truncateTitle(title string, limit int) string {
	runes := []rune(title)
	if len(runes) <= limit {
		return title
	}
	return string(runes[:limit-1]) + "…"
}

// cleanQueueLists evicts panels older than timeout and removes their messages.
// Called from the same ticker that sweeps search results - one goroutine for
// both caches is enough, and a second ticker would only add shutdown surface.
func (m *MusicBot) cleanQueueLists(timeout time.Duration) {
	m.queueListMu.Lock()
	defer m.queueListMu.Unlock()
	for key, state := range m.queueLists {
		if time.Since(state.timestamp) <= timeout {
			continue
		}
		if state.channelId != "" && state.messageId != "" {
			if err := m.session.ChannelMessageDelete(state.channelId, state.messageId); err != nil {
				utils.ErrorLog.Println("Failed to remove queue list message:", err)
			}
		}
		delete(m.queueLists, key)
	}
}
