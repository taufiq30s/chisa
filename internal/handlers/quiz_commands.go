package handlers

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/quiz"
	"github.com/taufiq30s/chisa/internal/responses"
)

const quizFeatureName = "Interactive Quiz"

var quizCommandHandler = func(chisa *bot.Bot, i *discordgo.InteractionCreate) {
	// Guard: quiz service must be initialized
	if chisa.Quiz == nil {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature:     quizFeatureName,
			Title:       "Service Unavailable",
			Description: "Quiz service is not initialized. Please contact the administrator.",
		}).Execute()
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return
	}

	switch options[OptionIndexFirst].Name {
	case "start":
		handleQuizStart(chisa, i, options[OptionIndexFirst])
	case "stop":
		handleQuizStop(chisa, i)
	}
}

func handleQuizStart(chisa *bot.Bot, i *discordgo.InteractionCreate, sub *discordgo.ApplicationCommandInteractionDataOption) {
	// Permission check: caller must have MANAGE_CHANNELS
	if !hasManageChannels(chisa.Session, i) {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature:     quizFeatureName,
			Title:       "Permission Denied",
			Description: "You need the **Manage Channels** permission to start a quiz session.",
		}).Execute()
		return
	}

	// Guard: session already active
	if chisa.Quiz.IsActive() {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature:     quizFeatureName,
			Title:       "Session Already Active",
			Description: "A quiz session is already running. Use `/interactive-quiz stop` first.",
		}).Execute()
		return
	}

	// Resolve the channel option
	channelOption := sub.Options[OptionIndexFirst].ChannelValue(chisa.Session)
	if channelOption == nil {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature:     quizFeatureName,
			Title:       "Invalid Channel",
			Description: "Could not resolve the provided channel.",
		}).Execute()
		return
	}

	// Validate it's a Stage channel
	if channelOption.Type != discordgo.ChannelTypeGuildStageVoice {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature:     quizFeatureName,
			Title:       "Not a Stage Channel",
			Description: fmt.Sprintf("**%s** is not a Stage channel. Please select a Stage channel.", channelOption.Name),
		}).Execute()
		return
	}

	// Start session
	sessionID, wsEndpoint, err := chisa.Quiz.Start(chisa.Session, i.GuildID, channelOption.ID)
	if err != nil {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature: quizFeatureName,
			Title:   "Failed to Start Session",
			Err:     err,
		}).Execute()
		return
	}

	// Re-hydrate queue from existing raise-hands
	rehydrateQueue(chisa, channelOption.ID, i.GuildID)

	responses.InteractionResponse(chisa.Session, i.Interaction).WithEmbed(
		responses.CreateMessageEmbed(chisa.Session,
			"Quiz Session Started",
			"",
			quizFeatureName,
			responses.SetColor("0bdd47"),
			responses.SetFields([]*discordgo.MessageEmbedField{
				{Name: "Channel", Value: fmt.Sprintf("<#%s>", channelOption.ID), Inline: true},
				{Name: "WS Session ID", Value: fmt.Sprintf("`%s`", sessionID), Inline: false},
				{Name: "WS Endpoint", Value: fmt.Sprintf("`%s`", wsEndpoint), Inline: false},
			}),
		),
	).Send()
}

func handleQuizStop(chisa *bot.Bot, i *discordgo.InteractionCreate) {
	// Permission check
	if !hasManageChannels(chisa.Session, i) {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature:     quizFeatureName,
			Title:       "Permission Denied",
			Description: "You need the **Manage Channels** permission to stop a quiz session.",
		}).Execute()
		return
	}

	// Guard: no active session
	if !chisa.Quiz.IsActive() {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature:     quizFeatureName,
			Title:       "No Active Session",
			Description: "There is no quiz session currently running.",
		}).Execute()
		return
	}

	totalPromoted, channelID, err := chisa.Quiz.Stop()
	if err != nil {
		responses.ErrorResponse(chisa.Session, i, &responses.ErrorResponseData{
			Feature: quizFeatureName,
			Title:   "Failed to Stop Session",
			Err:     err,
		}).Execute()
		return
	}

	responses.InteractionResponse(chisa.Session, i.Interaction).WithEmbed(
		responses.CreateMessageEmbed(chisa.Session,
			"Quiz Session Stopped",
			"",
			quizFeatureName,
			responses.SetColor("c30010"),
			responses.SetFields([]*discordgo.MessageEmbedField{
				{Name: "Channel", Value: fmt.Sprintf("<#%s>", channelID), Inline: true},
				{Name: "Participants Served", Value: fmt.Sprintf("%d", totalPromoted), Inline: true},
			}),
		),
	).Send()
}

// hasManageChannels checks whether the interaction member has MANAGE_CHANNELS.
func hasManageChannels(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil {
		return false
	}
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		return false
	}
	return perms&discordgo.PermissionManageChannels != 0
}

// rehydrateQueue seeds the active session queue from users already raising
// their hands in the Stage channel at the time the session starts.
func rehydrateQueue(chisa *bot.Bot, channelID, guildID string) {
	g, err := chisa.Session.State.Guild(guildID)
	if err != nil {
		return
	}
	for _, vs := range g.VoiceStates {
		if vs.ChannelID != channelID {
			continue
		}
		if vs.RequestToSpeakTimestamp == nil {
			continue
		}
		if vs.UserID == chisa.Session.State.User.ID {
			continue
		}
		// if chisa.Quiz.GetSession().IsPromoted(vs.UserID) {
		// 	continue
		// }
		member, err := chisa.Session.GuildMember(guildID, vs.UserID)
		username := vs.UserID
		avatar := ""
		if err == nil && member.User != nil {
			username = member.User.Username
			avatar = member.User.AvatarURL("128")
		}
		chisa.Quiz.GetSession().AddToQueue(quizEntry(vs.UserID, username, avatar, *vs.RequestToSpeakTimestamp))
	}
}

func quizEntry(userID, username, avatar string, ts time.Time) quiz.QueueEntry {
	return quiz.QueueEntry{
		UserID:    userID,
		Username:  username,
		Avatar:    avatar,
		Timestamp: ts,
	}
}

var quizCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "interactive-quiz",
		Description: "Manage an interactive quiz session on a Stage channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "start",
				Description: "Start a quiz session on a Stage channel",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionChannel,
						Name:         "channel",
						Description:  "The Stage channel to use",
						Required:     true,
						ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildStageVoice},
					},
				},
			},
			{
				Name:        "stop",
				Description: "Stop the current quiz session",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},
}
