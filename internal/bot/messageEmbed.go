package bot

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

/*
Create a message embed with the given title, description, feature name, and options.
parameters:
Client: The Discord client.
Title: The title of the embed.
Description: The description of the embed.
FeatureName: The name of the feature.
Options: A list of functions that modify the embed.
Returns:
The created message embed.

Example usage:
embed := CreateMessageEmbed(client, "Title", "Description", "FeatureName", SetColor("#FFFFFF"))

	client.SendInteractionResponse(interaction, &discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		}
	)

This Example creates an embed with the title "Title", description "Description", feature name "FeatureName", and color "#FFFFFF".
*/
func CreateMessageEmbed(
	session *discordgo.Session,
	title string, description string,
	featureName string,
	options ...func(*discordgo.MessageEmbed),
) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    featureName,
			IconURL: session.State.User.AvatarURL(""),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Chisa Version: 0.0.1 ",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	for _, option := range options {
		option(embed)
	}
	return embed
}

func convertHexToInt(color string) int {
	color = strings.Replace(color, "#", "", -1)
	color = strings.ToLower(color)
	base := 1
	result := 0
	for i := len(color) - 1; i >= 0; i-- {
		if color[i] >= '0' && color[i] <= '9' {
			result += int(color[i]-'0') * base
		} else if color[i] >= 'a' && color[i] <= 'f' {
			result += int(color[i]-'a'+10) * base
		}
		base *= 16
	}
	return result
}

func SetColor(color string) func(*discordgo.MessageEmbed) {
	return func(embed *discordgo.MessageEmbed) {
		embed.Color = convertHexToInt(color)
	}
}

func SetUrl(url string) func(*discordgo.MessageEmbed) {
	return func(embed *discordgo.MessageEmbed) {
		embed.URL = url
	}
}

func SetFields(fields []*discordgo.MessageEmbedField) func(*discordgo.MessageEmbed) {
	return func(embed *discordgo.MessageEmbed) {
		embed.Fields = fields
	}
}

func SetThumbnailUrl(url string) func(*discordgo.MessageEmbed) {
	return func(embed *discordgo.MessageEmbed) {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: url,
		}
	}
}

func SetImageUrl(url string) func(*discordgo.MessageEmbed) {
	return func(embed *discordgo.MessageEmbed) {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: url,
		}
	}
}
