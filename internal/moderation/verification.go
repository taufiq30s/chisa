package moderation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

// Constant variable that store handlers
var (
	VerificationCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "verify",
			Description: "Send verification request for admin review",
		},
	}
	VerificationCommandHandlers = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		if cmd := interaction.ApplicationCommandData(); cmd.Name == "verify" {
			sendRequestVerificationHandle(chisa, interaction)
		}
	}
)

// Handle New Membership Verification Response
func VerificationResponseButtonHandle() map[string]func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
	return map[string]func(chisa *bot.Bot, interaction *discordgo.InteractionCreate){
		"acc-req-accept": handleVerificationAccept,
		"acc-req-reject": handleVerificationReject,
	}
}

// Handle Send Request Verification
//
// When new member execute "/verify" command, chisa will send confirmation
// message to admin or moderator and chisa will return message if command
// was sent.
func sendRequestVerificationHandle(chisa *bot.Bot, i *discordgo.InteractionCreate) {
	var (
		modChannel = getModeratorChannelId()
		logChannel = getLogChannel()
	)

	err := sendRequestVerificationToAdmin(chisa.Session, i.Member.User, modChannel)
	if err != nil {
		utils.ErrorLog.Println(err)
		chisa.Session.ChannelMessageSendEmbed(logChannel, responses.CreateMessageEmbed(chisa.Session,
			"Error",
			fmt.Sprintf(
				"Failed to send message with error \n``%s``",
				err),
			"Moderation",
			responses.SetColor("df0000"),
		))
	}

	responses.InteractionResponse(chisa.Session, i.Interaction, discordgo.InteractionResponseChannelMessageWithSource, true, responses.CreateMessageEmbed(
		chisa.Session,
		"Verification Request Sent",
		"You request has been sent to moderator and we will process it.",
		featureName,
		responses.SetColor("0bdd47"),
	)).Execute()
}

func sendRequestVerificationToAdmin(s *discordgo.Session, newMember *discordgo.User, modChannel string) error {
	requestMessageEmbed := responses.CreateMessageEmbed(
		s,
		featureName,
		fmt.Sprintf(
			"A user with the name %s (username: %s) sends a request to verification.\nDo you want to accept it?",
			newMember.GlobalName, newMember.Username,
		),
		featureName,
		responses.SetColor("0bdd47"),
	)
	_, err := s.ChannelMessageSendComplex(modChannel, &discordgo.MessageSend{
		Embed: requestMessageEmbed,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Emoji: &discordgo.ComponentEmoji{
							Name: "✅",
						},
						Label:    "Accept",
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("acc-req-accept-%s", newMember.ID),
					},
					discordgo.Button{
						Emoji: &discordgo.ComponentEmoji{
							Name: "❎",
						},
						Label:    "Reject",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("acc-req-reject-%s", newMember.ID),
					},
				},
			},
		},
	})
	return err
}

// Handle when admin accept request by add "verify" role and send
// Welcome message to "welcome" channel.
func handleVerificationAccept(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
	var responseEmbed *discordgo.MessageEmbed
	memberId := interaction.MessageComponentData().CustomID[strings.LastIndex(interaction.MessageComponentData().CustomID, "-")+1:]

	verifiedRoleId, err := utils.GetEnv("AKASHIC_VERIFIED_ROLE_ID")
	if err != nil {
		utils.ErrorLog.Println(err)
		return
	}

	welcomeChannelId, err := utils.GetEnv("AKASHIC_WELCOME_CHANNEL_ID")
	if err != nil {
		utils.ErrorLog.Println(err)
		return
	}

	ruleChannelId, err := utils.GetEnv("AKASHIC_RULE_CHANNEL_ID")
	if err != nil {
		utils.ErrorLog.Println(err)
		return
	}

	// Check member exists
	member, err := chisa.Session.GuildMember(interaction.Interaction.GuildID, memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		// When new member is not found or has left the server
		// before being approved
		if strings.Contains(err.Error(), "404 Not Found") {
			responses.ErrorResponse(
				chisa.Session,
				interaction,
				&responses.ErrorResponseData{
					Feature:     featureName,
					Title:       "Failed to process request",
					Description: "Sorry, your request failed to process because `member id` not found!",
				},
			).Execute()
		}
		return
	}

	if slices.Contains(member.Roles, verifiedRoleId) {
		responses.ErrorResponse(
			chisa.Session,
			interaction,
			&responses.ErrorResponseData{
				Feature:     featureName,
				Title:       "Failed to process request",
				Description: "Sorry, this member was verified!",
			},
		).Execute()
		return
	}

	// Assign "verify" role to new member and send status response to moderator
	go func() {
		err := chisa.Session.GuildMemberRoleAdd(interaction.Interaction.GuildID, memberId, verifiedRoleId)
		if err != nil {
			utils.ErrorLog.Println(err)
			responses.ErrorResponse(
				chisa.Session,
				interaction,
				&responses.ErrorResponseData{
					Feature: featureName,
					Title:   "Failed to process request",
					Description: fmt.Sprintf(`Sorry, your request failed to process!
					Detail:
					%s`, err.Error()),
				},
			).Execute()
		}
		responseEmbed = responses.CreateMessageEmbed(
			chisa.Session,
			"Accepted New Member Success",
			fmt.Sprintf(
				"%s has been processed to get channel access and assign “verified” role.",
				member.User.GlobalName),
			featureName,
			responses.SetColor("0bdd47"),
		)

		// Send response to moderator
		responses.InteractionResponse(
			chisa.Session,
			interaction.Interaction,
			discordgo.InteractionResponseChannelMessageWithSource,
			false,
			responseEmbed,
		).Execute()
	}()

	// Send welcome message to new member in "welcome" channel
	go func() {
		responseEmbed = responses.CreateMessageEmbed(
			chisa.Session,
			fmt.Sprintf("Welcome to %s", chisa.Session.State.Guilds[0].Name),
			fmt.Sprintf(`Hello <@%s>, welcome to %s.
					Please see the server rules at <#%s>.
					If you have any questions or suggestions, please ask \"Pengasuh Anak\"`,
				memberId, chisa.Session.State.Guilds[0].Name, ruleChannelId,
			),
			featureName,
			responses.SetColor("0bdd47"),
		)

		chisa.Session.ChannelMessageSendEmbed(
			welcomeChannelId,
			responseEmbed,
		)
	}()
}

// Handle when admin reject request then kick rejected new member
// from server and send DM to confirm to people who give
// him invitation link
func handleVerificationReject(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
	var responseEmbed *discordgo.MessageEmbed
	memberId := interaction.MessageComponentData().CustomID[strings.LastIndex(interaction.MessageComponentData().CustomID, "-")+1:]

	// Check member exists
	member, err := chisa.Session.GuildMember(interaction.Interaction.GuildID, memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		// When new member is not found or has left the server
		// before being approved
		if strings.Contains(err.Error(), "404 Not Found") {
			responses.ErrorResponse(
				chisa.Session,
				interaction,
				&responses.ErrorResponseData{
					Feature:     featureName,
					Title:       "Failed to process request",
					Description: "Sorry, your request failed to process because `member id` not found!",
				},
			).Execute()
		}
		return
	}

	// Kick rejected new member from server and send him DM)
	userChannel, err := chisa.Session.UserChannelCreate(memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		responses.ErrorResponse(
			chisa.Session,
			interaction,
			&responses.ErrorResponseData{
				Feature: featureName,
				Title:   "Failed to process request",
				Description: fmt.Sprintf(`Sorry, your request failed to process!
				Detail:
				%s`, err.Error()),
			},
		).Execute()
	}

	// Kick Member
	go func() {
		err := chisa.Session.GuildMemberDeleteWithReason(
			interaction.Interaction.GuildID,
			memberId,
			"Request rejected by admin",
		)
		if err != nil {
			utils.ErrorLog.Println(err)
			responses.ErrorResponse(
				chisa.Session,
				interaction,
				&responses.ErrorResponseData{
					Feature: featureName,
					Title:   "Failed to process request",
					Description: fmt.Sprintf(`Sorry, your request failed to process!
					Detail:
					%s`, err.Error()),
				},
			).Execute()
		}

		// Send response to admin
		responseEmbed = responses.CreateMessageEmbed(
			chisa.Session,
			"Rejected New Member Success",
			fmt.Sprintf(
				"%s has been rejected and kicked from server.",
				member.User.GlobalName),
			featureName,
			responses.SetColor("df0000"),
		)
		responses.InteractionResponse(
			chisa.Session,
			interaction.Interaction,
			discordgo.InteractionResponseChannelMessageWithSource,
			false,
			responseEmbed,
		).Execute()
	}()

	// Send DM to rejected member
	go func() {
		responseEmbed = responses.CreateMessageEmbed(
			chisa.Session,
			"Verification Rejected",
			`Sorry, your request was rejected!
			Please contact the source of the invitation link for further confirmation`,
			featureName,
			responses.SetColor("df0000"),
		)
		_, err = chisa.Session.ChannelMessageSendEmbed(
			userChannel.ID, responseEmbed)
		if err != nil {
			utils.ErrorLog.Println(err)
			// Send error message
			responseEmbed = responses.CreateMessageEmbed(
				chisa.Session,
				"Failed to send DM",
				fmt.Sprintf(`Sorry, your request failed to process!
				Detail:
				%s`, err.Error()),
				featureName,
				responses.SetColor("df0000"),
			)
			chisa.Session.ChannelMessageSendEmbed(
				interaction.Interaction.ChannelID,
				responseEmbed,
			)
		}
	}()
}
