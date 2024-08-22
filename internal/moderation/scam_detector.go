package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/utils"
)

var (
	DATABSE_SCAM_URLS = "https://raw.githubusercontent.com/Discord-AntiScam/scam-links/main/list.json"
	ctx               = context.Background()
)

func getLogChannel() string {
	logChannel, err := utils.GetEnv("CHISA_LOG_CHANNEL_ID")
	if err != nil {
		log.Fatalln(err)
	}
	return logChannel
}

// Update Scam Links
// This feature using dataset from The DSP Project
//
// You can check their repository:
//
// https://github.com/Discord-AntiScam/scam-links
func UpdateDataset(client *redis.Client) error {
	resp, err := http.Get(DATABSE_SCAM_URLS)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to GET dataset: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var urls []string
	if err = json.Unmarshal(body, &urls); err != nil {
		return err
	}

	client.Del(ctx, "scam_dataset")
	client.SAdd(ctx, "scam_dataset", urls)

	return nil
}

// Check message content is spam or not.
//
// It will check based on certain keywords and urls.
//
// Return:
//
//   - 0 if not scam
//
//   - 1 if suspect scam (using @everyone or @here)
//
//   - 2 if positively scam
func CheckScam(c *redis.Client, s *discordgo.Session, m *discordgo.MessageCreate) uint8 {
	isSuspect, isContainScamLink := false, false

	// Preprocess text before checking
	re := regexp.MustCompile(`\r?\n`)
	message := re.ReplaceAllString(m.Content, " ")
	words := strings.ToLower(message)

	// Check message mention @everyone or @here and get urls
	isSuspect = strings.Contains(words, "@everyone") || strings.Contains(words, "@here")
	urlPattern := regexp.MustCompile(`(http|https):\/\/[^\s]+`)
	links := urlPattern.FindAllString(message, -1)
	for _, link := range links {
		// Check word is URL or not
		uri, err := url.ParseRequestURI(link)
		if err != nil {
			continue
		}
		domain := uri.Hostname()

		// Check domain contains in scam urls
		isScam, err := c.SIsMember(ctx, "scam_dataset", domain).Result()
		if err != nil {
			log.Fatalf("Failed to checking redis: %v\n", err)
		}
		if isScam {
			isContainScamLink = true
			break
		}
	}
	if isSuspect {
		return 1
	} else if isContainScamLink {
		return 2
	} else {
		return 0
	}
}

// Handle scam meesage
//
// The message will be deleted and the user will be given a time out for a day
// (if suspected) and 7 days if positively scam. Admin can choose to remove
// the timeout or block it if account cannot be recovered.
func HandleScamMessage(s *discordgo.Session, m *discordgo.MessageCreate, code uint8) {
	var (
		timeoutDay       int
		status           string
		scamMessageEmbed *discordgo.MessageEmbed
		title            = "Scam message has detected"
		color            = bot.SetColor("df0000")
		footer           = "Anti Scam Detector"
		titleError       = "Error"
	)
	logChannel := getLogChannel()

	// Create Message Embed
	switch code {
	case 1:
		timeoutDay = 1
		status = "Suspected scam"
	case 2:
		timeoutDay = 7
		status = "Scam"
	default:
		timeoutDay = 0
	}

	if timeoutDay == 0 {
		return
	}
	timeout := time.Now().AddDate(0, 0, timeoutDay)
	scamMessageEmbed = bot.CreateMessageEmbed(s,
		title,
		fmt.Sprintf(
			"%s message detected by account ``%s<@%s>`` with message content"+
				"\n\n``%s``\n\n"+
				"The account has been **timeout** for %d day."+
				"\n\nIf this is not scam, please click **'Not a scam'**",
			status, m.Author.GlobalName, m.Author.Username, m.Content, timeoutDay),
		footer,
		color,
	)
	// Send message to log, timeout user and remove spam message
	_, err := s.ChannelMessageSendComplex(logChannel, &discordgo.MessageSend{
		Embed: scamMessageEmbed,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Emoji: &discordgo.ComponentEmoji{
							Name: "⛔",
						},
						Label:    "Ban User",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("scam-ban-%s", m.Author.ID),
					},
					discordgo.Button{
						Emoji: &discordgo.ComponentEmoji{
							Name: "✅",
						},
						Label:    "Not a Scam",
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("scam-remove-timeout-%s", m.Author.ID),
					},
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
		s.ChannelMessageSendEmbed(logChannel, bot.CreateMessageEmbed(s,
			titleError,
			fmt.Sprintf(
				"Failed to send message with error \n``%s``",
				err),
			footer,
			color,
		))
	}
	err = s.GuildMemberTimeout(m.GuildID, m.Author.ID, &timeout)
	if err != nil {
		log.Println(err)
		s.ChannelMessageSendEmbed(logChannel, bot.CreateMessageEmbed(s,
			titleError,
			fmt.Sprintf(
				"Failed to timeout with error \n``%s``",
				err),
			footer,
			color,
		))
	}
	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Println(err)
		s.ChannelMessageSendEmbed(logChannel, bot.CreateMessageEmbed(s,
			titleError,
			fmt.Sprintf(
				"Failed to remove message with error \n``%s``",
				err),
			footer,
			color,
		))
	}
}

// Get Scam Detector button trigger
func GetScamButtonHandlers() map[string]func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
	return map[string]func(chisa *bot.Bot, interaction *discordgo.InteractionCreate){
		"scam-ban": func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
			userId := interaction.MessageComponentData().CustomID[strings.LastIndex(interaction.MessageComponentData().CustomID, "-")+1:]
			err := chisa.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						bot.CreateMessageEmbed(chisa.Session,
							"Ban Successful",
							fmt.Sprintf(
								"<@%s> has been banned.", userId),
							"Moderation",
							bot.SetColor("0bdd47"),
						),
					},
				},
			})
			if err != nil {
				log.Println(err)
			}
			chisa.Session.GuildBanCreateWithReason(interaction.GuildID, userId, "Compromise account/indicated scam", 0)
		},
		"scam-remove-timeout": func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
			userId := interaction.MessageComponentData().CustomID[strings.LastIndex(interaction.MessageComponentData().CustomID, "-")+1:]
			err := chisa.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						bot.CreateMessageEmbed(chisa.Session,
							"Remove timeout successful",
							"Timeout removed.",
							"Moderation",
							bot.SetColor("0bdd47"),
						),
					},
				},
			})
			if err != nil {
				log.Println(err)
			}
			chisa.Session.GuildMemberTimeout(interaction.GuildID, userId, nil)
		},
	}
}
