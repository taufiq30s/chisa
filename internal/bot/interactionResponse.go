package bot

import "github.com/bwmarrin/discordgo"

type interactionResponse struct {
	Session      *discordgo.Session
	Interaction  *discordgo.Interaction
	ResponseType discordgo.InteractionResponseType
	Ephemeral    bool
	Data         *discordgo.MessageEmbed
}

func InteractionResponse(
	session *discordgo.Session,
	interaction *discordgo.Interaction,
	responseType discordgo.InteractionResponseType,
	isEphemeral bool,
	data *discordgo.MessageEmbed) interactionResponse {
	return interactionResponse{
		session,
		interaction,
		responseType,
		isEphemeral,
		data,
	}
}

func (response interactionResponse) Execute() {
	data := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			response.Data,
		},
	}

	if response.Ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}

	response.Session.InteractionRespond(
		response.Interaction,
		&discordgo.InteractionResponse{
			Type: response.ResponseType,
			Data: data,
		},
	)
}
