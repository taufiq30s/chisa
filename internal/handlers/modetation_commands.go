package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/moderation"
)

var (
	VerificationCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "verify",
			Description: "Send verification request for admin review",
		},
	}
	VerificationCommandHandlers = func(chisa *bot.Bot, i *discordgo.InteractionCreate) {
		if bot := i.ApplicationCommandData(); bot.Name == "verify" {
			moderation.SendRequestVerificationHandle(chisa.Session, i)
		}
	}
	ScamButtonResponseHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"scam-ban":            moderation.BanScammerHandler,
		"scam-remove-timeout": moderation.RemoveSuspectHandler,
	}
	VerificationButtonResponseHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"acc-req-accept": moderation.HandleVerificationAccept,
		"acc-req-reject": moderation.HandleVerificationReject,
	}
)
