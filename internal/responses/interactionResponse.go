package responses

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/utils"
)

type interactionResponse struct {
	session      *discordgo.Session
	interaction  *discordgo.Interaction
	responseType discordgo.InteractionResponseType
	ephemeral    bool
	data         *discordgo.MessageEmbed
	components   []discordgo.MessageComponent
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
		[]discordgo.MessageComponent{},
	}
}

func (r *interactionResponse) AddActionRow(components *discordgo.ActionsRow) {
	r.components = append(r.components, components)
}

func (r *interactionResponse) SetUpdateResponse() {
	r.responseType = discordgo.InteractionResponseUpdateMessage
}

func (response interactionResponse) Execute() {
	data := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			response.data,
		},
	}

	if len(response.components) > 0 {
		data.Components = response.components
	}

	if response.ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}

	err := response.session.InteractionRespond(
		response.interaction,
		&discordgo.InteractionResponse{
			Type: response.responseType,
			Data: data,
		},
	)

	if err != nil {
		fmt.Println("Error responding to interaction:", err)
		utils.ErrorLog.Println("Error responding to interaction:", err)
	}
}
