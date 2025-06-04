package moderation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

// Handle Send Request Verification
//
// When new member execute "/verify" command, chisa will send confirmation
// message to admin or moderator and chisa will return message if command
// was sent.
func SendRequestVerificationHandle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var (
		modChannel = getModeratorChannelId()
		logChannel = getLogChannel()
	)

	// Check member has @verified role and executed in specific channel
	// if yes, return message
	if i.ChannelID != getVerificationChannelId() {
		embed := responses.CreateMessageEmbed(
			s,
			"Wrong Channel",
			"You can only use this command in the **verification** channel.",
			featureName,
			responses.SetColor("df0000"),
		)
		responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SetEphemeral().Send()
		return
	}
	if slices.Contains(i.Member.Roles, getVerifiedRoleId()) {
		embed := responses.CreateMessageEmbed(
			s,
			"You are already verified",
			"You already have the **verified** role, which means you've completed the verification process successfully. There's no need to use the **/verify** command again.\n\nIf you believe this is a mistake or you lost access to certain channels, please contact **\"Pengasuh Anak\"**.",
			featureName,
			responses.SetColor("df0000"),
		)
		responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SetEphemeral().Send()
		return
	}

	err := SendRequestVerificationToAdmin(s, i.Member.User, modChannel)
	if err != nil {
		utils.ErrorLog.Println(err)
		s.ChannelMessageSendEmbed(logChannel, responses.CreateMessageEmbed(s,
			"Error",
			fmt.Sprintf(
				"Failed to send message with error \n``%s``",
				err),
			"Moderation",
			responses.SetColor("df0000"),
		))
		return
	}

	embed := responses.CreateMessageEmbed(
		s,
		"Verification Request Sent",
		"You request has been sent to moderator and we will process it.",
		featureName,
		responses.SetColor("0bdd47"),
	)

	responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SetEphemeral().Send()
}

func SendRequestVerificationToAdmin(s *discordgo.Session, newMember *discordgo.User, modChannel string) error {
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
func HandleVerificationAccept(s *discordgo.Session, i *discordgo.InteractionCreate, params ...any) {
	var responseEmbed *discordgo.MessageEmbed
	memberId := i.MessageComponentData().CustomID[strings.LastIndex(i.MessageComponentData().CustomID, "-")+1:]
	verifiedRoleId = getVerifiedRoleId()

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
	member, err := s.GuildMember(i.Interaction.GuildID, memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		// When new member is not found or has left the server
		// before being approved
		if strings.Contains(err.Error(), "404 Not Found") {
			responses.ErrorResponse(
				s,
				i,
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
			s,
			i,
			&responses.ErrorResponseData{
				Feature:     featureName,
				Title:       "Failed to process request",
				Description: "Sorry, this member was verified!",
			},
		).Execute()
		return
	}

	// Assign "verify" role to new member and send status response to moderator
	err = s.GuildMemberRoleAdd(i.Interaction.GuildID, memberId, verifiedRoleId)
	if err != nil {
		utils.ErrorLog.Println(err)
		responses.ErrorResponse(
			s,
			i,
			&responses.ErrorResponseData{
				Feature: featureName,
				Title:   "Failed to process request",
				Description: fmt.Sprintf(`Sorry, your request failed to process!
					Detail:
					%s`, err.Error()),
			},
		).SetResponseTypeAsUpdate().Execute()
		return
	}
	responseEmbed = responses.CreateMessageEmbed(
		s,
		"Accepted New Member Success",
		fmt.Sprintf(
			"%s has been processed to get channel access and assign “verified” role.",
			member.User.GlobalName),
		featureName,
		responses.SetColor("0bdd47"),
	)

	// Send response to moderator
	responses.InteractionResponse(s, i.Interaction).WithEmbed(responseEmbed).SetResponseTypeAsUpdate().Send()

	// Send welcome message to new member in "welcome" channel
	responseEmbed = responses.CreateMessageEmbed(
		s,
		fmt.Sprintf("Welcome to %s", s.State.Guilds[0].Name),
		fmt.Sprintf(`Hello <@%s>, welcome to %s.
				Please see the server rules at <#%s>.
				If you have any questions or suggestions, please ask \"Pengasuh Anak\"`,
			memberId, s.State.Guilds[0].Name, ruleChannelId,
		),
		featureName,
		responses.SetColor("0bdd47"),
	)

	s.ChannelMessageSendEmbed(
		welcomeChannelId,
		responseEmbed,
	)
}

// Handle when admin reject request then kick rejected new member
// from server and send DM to confirm to people who give
// him invitation link
func HandleVerificationReject(s *discordgo.Session, i *discordgo.InteractionCreate, params ...any) {
	var responseEmbed *discordgo.MessageEmbed
	memberId := i.MessageComponentData().CustomID[strings.LastIndex(i.MessageComponentData().CustomID, "-")+1:]

	// Check member exists
	member, err := s.GuildMember(i.Interaction.GuildID, memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		// When new member is not found or has left the server
		// before being approved
		if strings.Contains(err.Error(), "404 Not Found") {
			responses.ErrorResponse(
				s,
				i,
				&responses.ErrorResponseData{
					Feature:     featureName,
					Title:       "Failed to process request",
					Description: "Sorry, your request failed to process because `member id` not found!",
				},
			).SetResponseTypeAsUpdate().Execute()
		}
		return
	}

	// Kick rejected new member from server and send him DM)
	userChannel, err := s.UserChannelCreate(memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		responses.ErrorResponse(
			s,
			i,
			&responses.ErrorResponseData{
				Feature: featureName,
				Title:   "Failed to process request",
				Description: fmt.Sprintf(`Sorry, your request failed to process!
				Detail:
				%s`, err.Error()),
			},
		).SetResponseTypeAsUpdate().Execute()
		return
	}

	// Send DM to rejected member
	responseEmbed = responses.CreateMessageEmbed(
		s,
		"Verification Rejected",
		`Sorry, your request was rejected!
		Please contact the source of the invitation link for further confirmation`,
		featureName,
		responses.SetColor("df0000"),
	)
	_, err = s.ChannelMessageSendEmbed(
		userChannel.ID, responseEmbed)
	if err != nil {
		utils.ErrorLog.Println(err)
		// Send error message
		responseEmbed = responses.CreateMessageEmbed(
			s,
			"Failed to send DM",
			fmt.Sprintf(`Sorry, your request failed to process!
			Detail:
			%s`, err.Error()),
			featureName,
			responses.SetColor("df0000"),
		)
		s.ChannelMessageSendEmbed(
			i.Interaction.ChannelID,
			responseEmbed,
		)
		return
	}

	// Kick Member
	err = s.GuildMemberDeleteWithReason(
		i.Interaction.GuildID,
		memberId,
		"Request rejected by admin",
	)
	if err != nil {
		utils.ErrorLog.Println(err)
		responses.ErrorResponse(
			s,
			i,
			&responses.ErrorResponseData{
				Feature: featureName,
				Title:   "Failed to process request",
				Description: fmt.Sprintf(`Sorry, your request failed to process!
				Detail:
				%s`, err.Error()),
			},
		).SetResponseTypeAsUpdate().Execute()
		return
	}

	// Send response to admin
	responseEmbed = responses.CreateMessageEmbed(
		s,
		"Rejected New Member Success",
		fmt.Sprintf(
			"%s has been rejected and kicked from server.",
			member.User.GlobalName),
		featureName,
		responses.SetColor("df0000"),
	)
	responses.InteractionResponse(s, i.Interaction).WithEmbed(responseEmbed).SetResponseTypeAsUpdate().Send()
}
