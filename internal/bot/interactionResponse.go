package bot

import "github.com/bwmarrin/discordgo"

type interactionResponse struct {
	Session      *discordgo.Session
	Interaction  *discordgo.Interaction
	ResponseType discordgo.InteractionResponseType
	Data         *discordgo.MessageEmbed
}

func InteractionResponse(
	session *discordgo.Session,
	interaction *discordgo.Interaction,
	responseType discordgo.InteractionResponseType,
	data *discordgo.MessageEmbed) interactionResponse {
	return interactionResponse{
		session,
		interaction,
		responseType,
		data,
	}
}

func (response interactionResponse) Execute() {
	response.Session.InteractionRespond(
		response.Interaction,
		&discordgo.InteractionResponse{
			Type: response.ResponseType,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					response.Data,
				},
			},
		},
	)
}
