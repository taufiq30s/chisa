package bot

import "github.com/bwmarrin/discordgo"

type ErrorResponseData struct {
	Feature     string
	Title       string
	Description string
	Err         error
}
type errorResponse struct {
	session     *discordgo.Session
	interaction *discordgo.InteractionCreate
	data        *ErrorResponseData
}

func ErrorResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, data *ErrorResponseData) errorResponse {
	return errorResponse{
		session, interaction, data,
	}
}
func (res errorResponse) Execute() {
	res.session.InteractionRespond(
		res.interaction.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					CreateMessageEmbed(
						res.session,
						res.data.Title,
						func() string {
							if res.data.Description == "" {
								return res.data.Err.Error()
							}
							return res.data.Description
						}(),
						res.data.Feature,
						SetColor("c30010")),
				},
			},
		},
	)
}
