package responses

import (
	"github.com/bwmarrin/discordgo"
)

type interactionResponse struct {
	session             *discordgo.Session
	interaction         *discordgo.Interaction
	response            *discordgo.InteractionResponse
	ephemeral           bool
	embed               *discordgo.MessageEmbed
	components          []discordgo.MessageComponent
	temporaryComponents []discordgo.MessageComponent
	content             string
	err                 error
}

func InteractionResponse(session *discordgo.Session, interaction *discordgo.Interaction) *interactionResponse {
	return &interactionResponse{
		session:     session,
		interaction: interaction,
		response: &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		},
	}
}

func (r *interactionResponse) AddNewRow() *interactionResponse {
	r.components = append(r.components, discordgo.ActionsRow{
		Components: r.temporaryComponents,
	})
	r.temporaryComponents = nil
	return r
}

func (r *interactionResponse) WithContent(content string) *interactionResponse {
	r.content = content
	return r
}

func (r *interactionResponse) WithEmbed(embed *discordgo.MessageEmbed) *interactionResponse {
	r.embed = embed
	return r
}

func (r *interactionResponse) WithError(err error) {
	r.err = err
}

func (r *interactionResponse) WithButton(label string, style discordgo.ButtonStyle, customId string, isDisable bool) *interactionResponse {
	button := discordgo.Button{
		Label:    label,
		Style:    style,
		CustomID: customId,
		Disabled: isDisable,
	}

	r.temporaryComponents = append(r.temporaryComponents, button)
	return r
}

func (r *interactionResponse) WithSelectMenu(customId string, placeholder string, options func() (options []discordgo.SelectMenuOption)) *interactionResponse {
	selectMenu := discordgo.SelectMenu{
		CustomID:    customId,
		Placeholder: placeholder,
		Options:     options(),
	}

	r.temporaryComponents = append(r.temporaryComponents, selectMenu)
	return r
}

func (r *interactionResponse) SetEphemeral() *interactionResponse {
	r.ephemeral = true
	return r
}

func (r *interactionResponse) SetResponseTypeAsUpdate() *interactionResponse {
	r.response = &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
	}
	return r
}

func (r *interactionResponse) Defer() error {
	err := r.session.InteractionRespond(r.interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	return err
}

func (r *interactionResponse) Send() error {
	if r.err != nil && r.embed != nil {
		r.embed.Color = 0xc30010
		if r.embed.Description == "" {
			r.embed.Description = r.err.Error()
		}
	}

	// Initialize and set the response data
	if r.response.Data == nil {
		r.response.Data = &discordgo.InteractionResponseData{}
	}
	r.response.Data.Content = r.content
	r.response.Data.Embeds = []*discordgo.MessageEmbed{r.embed}

	if len(r.temporaryComponents) > 0 {
		r.AddNewRow()
	}

	if len(r.components) > 0 {
		r.response.Data.Components = r.components
	}

	return r.session.InteractionRespond(r.interaction, r.response)
}

func (r *interactionResponse) SendDefer() error {
	if r.err != nil && r.embed != nil {
		r.embed.Color = 0xc30010
		if r.embed.Description == "" {
			r.embed.Description = r.err.Error()
		}
	}

	if len(r.temporaryComponents) > 0 {
		r.AddNewRow()
	}

	edit := &discordgo.WebhookEdit{
		Content:    &r.content,
		Components: &r.components,
	}

	if r.embed != nil {
		edit.Embeds = &[]*discordgo.MessageEmbed{r.embed}
	}

	_, err := r.session.InteractionResponseEdit(r.interaction, edit)
	return err
}

// func (response interactionResponse) Execute() {
// 	data := &discordgo.InteractionResponseData{
// 		Embeds: []*discordgo.MessageEmbed{
// 			response.data,
// 		},
// 	}

// 	if len(response.components) > 0 {
// 		data.Components = response.components
// 	}

// 	if response.ephemeral {
// 		data.Flags = discordgo.MessageFlagsEphemeral
// 	}

// 	err := response.session.InteractionRespond(
// 		response.interaction,
// 		&discordgo.InteractionResponse{
// 			Type: response.responseType,
// 			Data: data,
// 		},
// 	)

// 	if err != nil {
// 		fmt.Println("Error responding to interaction:", err)
// 		utils.ErrorLog.Println("Error responding to interaction:", err)
// 	}
// }
