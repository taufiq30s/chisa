package responses

import (
	"github.com/bwmarrin/discordgo"
)

type ErrorResponseData struct {
	Feature     string
	Title       string
	Description string
	Err         error
}
type errorResponse struct {
	session     *discordgo.Session
	interaction *discordgo.InteractionCreate
	response    *discordgo.InteractionResponse
	embed       *discordgo.MessageEmbed
}

func ErrorResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, data *ErrorResponseData) *errorResponse {
	return &errorResponse{
		session:     session,
		interaction: interaction,
		response: &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		},
		embed: CreateMessageEmbed(
			session,
			data.Title,
			func() string {
				if data.Description == "" {
					return data.Err.Error()
				}
				return data.Description
			}(),
			data.Feature,
			SetColor("c30010"),
		),
	}
}
func (res *errorResponse) SetResponseTypeAsUpdate() *errorResponse {
	res.response = &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
	}
	return res
}
func (res *errorResponse) Execute() error {
	res.response.Data = &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{res.embed},
	}
	return res.session.InteractionRespond(res.interaction.Interaction, res.response)
}

func (res *errorResponse) ExecuteDefer() error {
	edit := &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{res.embed},
	}

	_, err := res.session.InteractionResponseEdit(res.interaction.Interaction, edit)
	return err
}
