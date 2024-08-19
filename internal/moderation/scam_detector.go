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
	SCAM_PATTERNS     = []string{
		`(?i)free discord nitro`,
		`(?i)click here`,
		`(?i)gift for you`,
	}
	SUSPECTION_URL = `(?i)discord\.gg\/[a-zA-Z0-9]+`
	ctx            = context.Background()
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
// It will check based on certain keywords and urls. If the message is scam or suspected,
// the message will be deleted and the user will be given a time out for a day
// (if suspected) and 7 days if positively scam. Admin can choose to remove
// the timeout or block it if account cannot be recovered.
func CheckScam(c *bot.Redis, s *discordgo.Session, m *discordgo.MessageCreate) bool {
	isSuspect, isContainScamLink := false, false
	var (
		timeout            time.Time
		scam_message_embed *discordgo.MessageEmbed
	)
	logChannel := getLogChannel()

	// Preprocess text before checking
	re := regexp.MustCompile(`\r?\n`)
	message := re.ReplaceAllString(m.Content, " ")
	words := strings.Split(strings.ToLower(message), " ")

	// Check message contains suspect scam
	for _, pattern := range SCAM_PATTERNS {
		match, _ := regexp.MatchString(pattern, m.Content)
		if match {
			isSuspect = true
			break
		}
	}

	for _, word := range words {
		// Check word is URL or not
		uri, err := url.ParseRequestURI(word)
		if err != nil {
			continue
		}
		domain := uri.Hostname()

		// Check domain contains in scam urls
		isScam, err := c.Client.SIsMember(ctx, "scam_dataset", domain).Result()
		if err != nil {
			log.Fatalf("Failed to checking redis: %v\n", err)
		}
		if isScam {
			isContainScamLink = true
			break
		}
	}
	if !isContainScamLink && !isSuspect {
		return false
	}
	if isContainScamLink {
		timeout = time.Now().AddDate(0, 0, 7)
		scam_message_embed = bot.CreateMessageEmbed(s,
			"Scam message has detected",
			fmt.Sprintf(
				"Scam message detected by account ``%s<@%s>`` with message content"+
					"\n\n``%s``\n\n"+
					"The account has been **timeout** for %d days."+
					"\n\nIf this is not scam, please click **'Not a scam'**", m.Author.GlobalName, m.Author.Username, m.Content, 7),
			"Anti Scam Detector",
			bot.SetColor("df0000"),
		)
	} else if isSuspect {
		timeout = time.Now().AddDate(0, 0, 1)
		scam_message_embed = bot.CreateMessageEmbed(s,
			"Scam message has detected",
			fmt.Sprintf(
				"Suspected scam message detected by account ``%s<@%s>`` with message content"+
					"\n\n``%s``\n\n"+
					"The account has been **timeout** for %d day."+
					"\n\nIf this is not scam, please click **'Not a scam'**", m.Author.GlobalName, m.Author.Username, m.Content, 1),
			"Anti Scam Detector",
			bot.SetColor("df0000"),
		)
	}
	// Send message to log, timeout user and remove spam message
	_, err := s.ChannelMessageSendComplex(logChannel, &discordgo.MessageSend{
		Embed: scam_message_embed,
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
						CustomID: fmt.Sprintf("scam-remove-timeout-%s", "155149108183695360"),
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	s.GuildMemberTimeout(m.GuildID, "155149108183695360", &timeout)
	s.ChannelMessageDelete(m.ChannelID, m.ID)
	return true
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
				log.Fatalln(err)
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
				log.Fatalln(err)
			}
			chisa.Session.GuildMemberTimeout(interaction.GuildID, userId, nil)
		},
	}
}
