package music

import (
	"github.com/bwmarrin/discordgo"
)

func (m *MusicBot) SearchNextPage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.moveSearchPage(s, i, true)
}

func (m *MusicBot) SearchPreviousPage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.moveSearchPage(s, i, false)
}

func (m *MusicBot) SearchSelect(s *discordgo.Session, i *discordgo.InteractionCreate, trackID string) {
	m.selectSearchResult(s, i, trackID)
}
