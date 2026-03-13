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

	if len(result) >= 20 {
		length = 20
		tracks = result[:20]
	} else {
		length = len(result)
		tracks = result
	}

	if length > 0 {
		m.searchResults[i.Member.User.ID] = &searchResultState{
			tracks:    tracks,
			page:      1,
			length:    length,
			timestamp: time.Now(),
			channelId: i.ChannelID,
		}
	}
	m.searchResult(s, i, 1, tracks, false)
}

func (m *MusicBot) moveSearchPage(s *discordgo.Session, i *discordgo.InteractionCreate, isNext bool) {
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
	state := m.searchResults[i.Member.User.ID]
	if state == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Title:       "Session Expired",
			Description: "Session expired, please run the command again",
			Feature:     m.featureName,
		})
		return
	}

	m.Load(s, i, trackID)
	s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	delete(m.searchResults, i.Member.User.ID)
}

/*
Create Search Result Embed and send it to the channel
*/
func (m *MusicBot) searchResult(s *discordgo.Session, i *discordgo.InteractionCreate, page int, result []lavalink.Track, isUpdate bool) {
	var (
		title string
		body  string
		limit int = 5
	)
	if len(result) < limit {
		limit = len(result)
	}

	if limit == 0 {
		title = "No result found"
		body = "No result found"
	} else {
		title = "Search result"
		for _, track := range result[(page-1)*limit : (page)*limit] {
			body += fmt.Sprintf(
				"%d. **%s - %s** `%02d:%02d`\n",
				(page-1)*limit+1,
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
			fmt.Println("Failed to defer interaction", err)
			utils.ErrorLog.Println("Failed to defer interaction", err)
			return
		}
		time.Sleep(500 * time.Millisecond)
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
			for _, track := range result[(page-1)*limit : (page)*limit] {
				options = append(options, discordgo.SelectMenuOption{
					Label: track.Info.Title,
					Value: track.Info.Identifier,
				})
			}
			return options
		},
	).AddNewRow().WithButton(
		"Previous", discordgo.PrimaryButton, fmt.Sprintf("search-previous-%s", i.Interaction.ID), page == 1,
	).WithButton("Next", discordgo.PrimaryButton, fmt.Sprintf("search-next-%s", i.Interaction.ID), len(result) <= page*limit)

	if !isUpdate {
		err = response.SendDefer()
	} else {
		err = response.Send()
	}
	if err != nil {
		fmt.Println("Failed to send search result", err)
		utils.ErrorLog.Println("Failed to send search result", err)
	}

	// Get Message ID after send respond
	msg, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		fmt.Println(err)
		return
	}
	m.searchResults[i.Member.User.ID].messageId = msg.ID
}

// Clean seach cache every 5 minutes and
// age of the cache is more than 5 minutes
func (m *MusicBot) InitializeCleanSearchCache() {
	timeout, _ := time.ParseDuration("1m")
	for {
		time.Sleep(timeout)
		for k, v := range m.searchResults {
			if time.Since(v.timestamp) > timeout {
				err := m.session.ChannelMessageDelete(v.channelId, v.messageId)
				if err != nil {
					fmt.Println("Failed to remove message:", err)
					utils.ErrorLog.Println("Failed to remove message:", err)
				}
				delete(m.searchResults, k)
			}
		}
	}
}
