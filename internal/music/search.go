package music

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/taufiq30s/chisa/internal/responses"
)

type searchResultState struct {
	tracks    []lavalink.Track
	page      int
	length    int
	timestamp time.Time
	channelId *string
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
			channelId: &i.ChannelID,
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

	m.Play(s, i, trackID)
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
		for _, track := range result[(page-1)*limit+1 : (page)*limit+1] {
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

	resp := responses.InteractionResponse(s, i.Interaction,
		discordgo.InteractionResponseChannelMessageWithSource,
		false,
		responses.CreateMessageEmbed(
			s, title,
			body,
			m.featureName,
			responses.SetColor("5e11d9"),
		),
	)
	resp.AddActionRow(&discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    fmt.Sprintf("search-select-%s", i.Member.User.ID),
				Placeholder: "Select a track",
				Options: func() (options []discordgo.SelectMenuOption) {
					for _, track := range result[(page-1)*limit+1 : (page)*limit+1] {
						options = append(options, discordgo.SelectMenuOption{
							Label: track.Info.Title,
							Value: track.Info.Identifier,
						})
					}
					return options
				}(),
			},
		},
	})
	resp.AddActionRow(&discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Disabled: page == 1,
				Label:    "Previous",
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("search-previous-%s", i.Interaction.ID),
			},
			discordgo.Button{
				Disabled: len(result) <= page*limit,
				Label:    "Next",
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("search-next-%s", i.Interaction.ID),
			},
		},
	})
	if isUpdate {
		resp.SetUpdateResponse()
	}
	resp.Execute()

	// Set message id after 10 seconds
	// go func() {
	// 	time.Sleep(10 * time.Second)
	// 	m.searchResults[i.Member.User.ID].messageId = i.Message.ID
	// }()
}

// Clean seach cache every 5 minutes and
// age of the cache is more than 5 minutes
func (m *MusicBot) InitializeCleanSearchCache() {
	for {
		time.Sleep(5 * time.Minute)
		for k, v := range m.searchResults {
			if time.Since(v.timestamp) > 5*time.Minute {
				m.session.ChannelMessageDelete(*v.channelId, v.messageId)
				delete(m.searchResults, k)
			}
		}
	}
}
