package music

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/model"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

type searchResultState struct {
	tracks    []model.Track
	page      int
	length    int
	timestamp time.Time
	channelId string
	messageId string
}

func (m *MusicBot) search(s *discordgo.Session, i *discordgo.InteractionCreate, result []model.Track) {
	var (
		length int
		tracks []model.Track
	)

	if len(result) >= MaxSearchResults {
		length = MaxSearchResults
		tracks = result[:MaxSearchResults]
	} else {
		length = len(result)
		tracks = result
	}

	if length > InitialQueueCapacity {
		m.searchMu.Lock()
		m.searchResults[i.Member.User.ID] = &searchResultState{
			tracks:    tracks,
			page:      InitialPage,
			length:    length,
			timestamp: time.Now(),
			channelId: i.ChannelID,
		}
		m.searchMu.Unlock()
	}
	m.searchResult(s, i, InitialPage, tracks, false)
}

func (m *MusicBot) moveSearchPage(s *discordgo.Session, i *discordgo.InteractionCreate, isNext bool) {
	m.searchMu.Lock()
	state := m.searchResults[i.Member.User.ID]
	if state == nil {
		m.searchMu.Unlock()
		// Execute() is required: a component interaction that is never
		// answered makes Discord show "This interaction failed" to the user.
		_ = responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Title:       "Session Expired",
			Description: "Session expired, please run the command again",
			Feature:     m.featureName,
		}).Execute()
		return
	}
	// Clamp the page number
	maxPage := max((len(state.tracks)+SearchResultsPerPage-1)/SearchResultsPerPage, InitialPage)

	if isNext && state.page < maxPage {
		state.page++
	} else if !isNext && state.page > InitialPage {
		state.page--
	}
	// Copy the page and tracks so the value still consistent after releasing the lock
	page := state.page
	tracks := state.tracks
	m.searchMu.Unlock()

	m.searchResult(s, i, page, tracks, true)
}

func (m *MusicBot) selectSearchResult(s *discordgo.Session, i *discordgo.InteractionCreate, trackID string) {
	m.searchMu.Lock()
	state := m.searchResults[i.Member.User.ID]
	if state == nil {
		m.searchMu.Unlock()
		_ = responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Title:       "Session Expired",
			Description: "Session expired, please run the command again",
			Feature:     m.featureName,
		}).Execute()
		return
	}

	var selectedTrack *model.Track
	for _, track := range state.tracks {
		if track.ID == trackID {
			selectedTrack = &track
			break
		}
	}
	delete(m.searchResults, i.Member.User.ID)
	m.searchMu.Unlock()

	if selectedTrack == nil {
		_ = responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Title:       "Track Not Found",
			Description: "The selected track was not found in the search results. Please try again.",
			Feature:     m.featureName,
		}).Execute()
		return
	}

	// Acknowledge the component interaction before doing real work. Discord
	// requires an answer within 3 seconds; resolving + joining voice is slower
	// than that. DeferredMessageUpdate also keeps the search panel on screen
	// until we delete it below.
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	// The search panel is a single message shared by every page, so throw it
	// away once a track has been picked - leaving it would let the user click
	// a stale option from an expired cache entry.
	if err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID); err != nil {
		utils.ErrorLog.Println("Failed to delete search message:", err)
	}

	// Join only now, once the user has actually picked a track. The voice state
	// is re-read here because the user may have left the channel while
	// browsing; ensureVoiceState cannot be reused because it replies through
	// the deferred interaction, and this one was already answered with
	// DeferredMessageUpdate (a second reply fails with 40060).
	voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		m.sendPlaybackError(&trackState{channelId: &i.ChannelID}, "Join a voice channel first, then pick a track again.")
		return
	}
	if err := m.ensureVoiceJoined(s, voiceState.ChannelID); err != nil {
		utils.ErrorLog.Println("Failed to join voice channel:", err)
		m.sendPlaybackError(&trackState{channelId: &i.ChannelID}, err.Error())
		return
	}

	m.addToQueue(*selectedTrack, i)
}

/*
Create Search Result Embed and send it to the channel
*/
func (m *MusicBot) searchResult(s *discordgo.Session, i *discordgo.InteractionCreate, page int, result []model.Track, isUpdate bool) {
	var (
		title string
		body  string
	)
	pageSize := SearchResultsPerPage
	start := (page - InitialPage) * pageSize
	end := min(start+pageSize, len(result))

	pageTracks := result[start:end]

	if len(pageTracks) == 0 {
		title = "No result found"
		body = "No result found"
	} else {
		title = "Search result"
		for idx, track := range pageTracks {
			body += fmt.Sprintf(
				"%d. **%s**\n",
				start+idx+1,
				track.Title,
			)
		}
	}

	// Do NOT defer here on the !isUpdate path: musicCommandHandler has already
	// acknowledged the interaction, and a second acknowledgement fails with
	// 40060 "Interaction has already been acknowledged" - the user would see
	// nothing. The bool only tells us whether we are editing an existing
	// message (pagination button) or posting a new one.
	var err error
	response := responses.InteractionResponse(s, i.Interaction)
	if isUpdate {
		response.SetResponseTypeAsUpdate()
	}

	embed := responses.CreateMessageEmbed(
		s, title,
		body,
		m.featureName,
		responses.SetColor("5e11d9"),
	)
	response.WithEmbed(embed).WithSelectMenu(
		fmt.Sprintf("search-select-%s", i.Member.User.ID),
		"Select a track",
		func() (options []discordgo.SelectMenuOption) {
			for _, track := range pageTracks {
				options = append(options, discordgo.SelectMenuOption{
					Label: track.Title,
					Value: track.ID,
				})
			}
			return options
		},
	).AddNewRow().WithButton(
		"Previous", discordgo.PrimaryButton, fmt.Sprintf("search-previous-%s", i.Interaction.ID), page == InitialPage,
	).WithButton("Next", discordgo.PrimaryButton, fmt.Sprintf("search-next-%s", i.Interaction.ID), len(result) <= page*pageSize)

	if !isUpdate {
		err = response.SendDefer()
	} else {
		err = response.Send()
	}
	if err != nil {
		utils.ErrorLog.Println("Failed to send search result", err)
	}

	// Get Message ID after send respond
	msg, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		utils.ErrorLog.Println("Failed to get interaction response:", err)
		return
	}
	m.searchMu.Lock()
	if state := m.searchResults[i.Member.User.ID]; state != nil {
		state.messageId = msg.ID
	}
	m.searchMu.Unlock()
}

// InitializeCleanSearchCache evicts expired paginated panels. It sweeps both
// caches (search picks and /music list) from a single ticker: both have the
// same TTL and both hold a Discord message that has to be deleted, so one
// goroutine is enough and there is only one shutdown path to reason about.
func (m *MusicBot) InitializeCleanSearchCache() {
	timeout := time.Minute
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case <-m.shutdown:
			return
		case <-ticker.C:
			m.searchMu.Lock()
			for k, v := range m.searchResults {
				if time.Since(v.timestamp) > timeout {
					if v.channelId != "" && v.messageId != "" {
						err := m.session.ChannelMessageDelete(v.channelId, v.messageId)
						if err != nil {
							utils.ErrorLog.Println("Failed to remove message:", err)
						}
					}
					delete(m.searchResults, k)
				}
			}
			m.searchMu.Unlock()
			m.cleanQueueLists(timeout)
		}
	}
}
