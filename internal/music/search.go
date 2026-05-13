package music

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

type searchResultState struct {
	tracks    []lavalink.Track
	page      int
	length    int
	timestamp time.Time
	channelId string
	messageId string
}

func (m *MusicBot) search(s *discordgo.Session, i *discordgo.InteractionCreate, result []lavalink.Track) {
	var (
		length int
		tracks []lavalink.Track
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
	defer m.searchMu.Unlock()
	state := m.searchResults[i.Member.User.ID]
	if state == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Title:       "Session Expired",
			Description: "Session expired, please run the command again",
			Feature:     m.featureName,
		})
		return
	}
	if isNext {
		state.page++
	} else {
		state.page--
	}
	m.searchResult(s, i, state.page, state.tracks, true)
}

func (m *MusicBot) selectSearchResult(s *discordgo.Session, i *discordgo.InteractionCreate, trackID string) {
	m.searchMu.Lock()
	state := m.searchResults[i.Member.User.ID]
	if state == nil {
		m.searchMu.Unlock()
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Title:       "Session Expired",
			Description: "Session expired, please run the command again",
			Feature:     m.featureName,
		})
		return
	}
	delete(m.searchResults, i.Member.User.ID)
	m.searchMu.Unlock()

	m.Load(s, i, trackID)
	s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
}

/*
Create Search Result Embed and send it to the channel
*/
func (m *MusicBot) searchResult(s *discordgo.Session, i *discordgo.InteractionCreate, page int, result []lavalink.Track, isUpdate bool) {
	var (
		title string
		body  string
		limit int = SearchResultsPerPage
	)
	if len(result) < limit {
		limit = len(result)
	}

	if limit == InitialQueueCapacity {
		title = "No result found"
		body = "No result found"
	} else {
		title = "Search result"
		for _, track := range result[(page-QueueOffsetForLength)*limit : (page)*limit] {
			body += fmt.Sprintf(
				"%d. **%s - %s** `%02d:%02d`\n",
				(page-QueueOffsetForLength)*limit+QueueOffsetForLength,
				track.Info.Author,
				track.Info.Title,
				int(track.Info.Length)/(1000*60),
				int(track.Info.Length)/1000%60,
			)
		}
	}

	var err error
	response := responses.InteractionResponse(s, i.Interaction)
	if !isUpdate {
		err = response.Defer()
		if err != nil {
			utils.ErrorLog.Println("Failed to defer interaction", err)
			return
		}
		time.Sleep(InteractionResponseDelay)
	} else {
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
			for _, track := range result[(page-QueueOffsetForLength)*limit : (page)*limit] {
				options = append(options, discordgo.SelectMenuOption{
					Label: track.Info.Title,
					Value: track.Info.Identifier,
				})
			}
			return options
		},
	).AddNewRow().WithButton(
		"Previous", discordgo.PrimaryButton, fmt.Sprintf("search-previous-%s", i.Interaction.ID), page == InitialPage,
	).WithButton("Next", discordgo.PrimaryButton, fmt.Sprintf("search-next-%s", i.Interaction.ID), len(result) <= page*limit)

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

// InitializeCleanSearchCache starts a background goroutine that periodically
// evicts expired search sessions (TTL = 1 minute) and removes their Discord messages.
func (m *MusicBot) InitializeCleanSearchCache() {
	timeout := time.Minute
	for {
		time.Sleep(timeout)
		m.searchMu.Lock()
		for k, v := range m.searchResults {
			if time.Since(v.timestamp) > timeout {
				err := m.session.ChannelMessageDelete(v.channelId, v.messageId)
				if err != nil {
					utils.ErrorLog.Println("Failed to remove message:", err)
				}
				delete(m.searchResults, k)
			}
		}
		m.searchMu.Unlock()
	}
}
