package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

const (
	verifyActionAccept = "accept"
	verifyActionReject = "reject"

	verifyDecisionKeyPrefix = "verify:decision:"
	verifyDecisionTTL       = 24 * time.Hour

	// 16 shards: enough to keep unrelated members from serialising, small
	// enough that memory stays bounded.
	verificationLockShards = 16
)

// verificationLocks serialises decisions for the same member. discordgo
// dispatches every event in its own goroutine, so two moderators clicking
// Accept and Reject at the same moment would otherwise both mutate the member.
var verificationLocks [verificationLockShards]sync.Mutex

func lockVerification(memberID string) func() {
	h := fnv.New32a()
	_, _ = h.Write([]byte(memberID))
	mu := &verificationLocks[h.Sum32()%uint32(len(verificationLocks))]
	mu.Lock()
	return mu.Unlock
}

// verificationDecision is the decision a moderator made, written to Redis
// before it is applied so a shutdown or crash cannot silently drop it.
type verificationDecision struct {
	Action      string `json:"action"`
	MemberID    string `json:"member_id"`
	ModeratorID string `json:"moderator_id"`
	GuildID     string `json:"guild_id"`
	At          int64  `json:"at"`
}

func savePendingDecision(decision verificationDecision) {
	if redisClient == nil {
		return
	}
	decision.At = time.Now().Unix()
	payload, err := json.Marshal(decision)
	if err != nil {
		utils.ErrorLog.Println(err)
		return
	}
	if err := redisClient.Set(context.Background(), verifyDecisionKeyPrefix+decision.MemberID, payload, verifyDecisionTTL).Err(); err != nil {
		utils.ErrorLog.Println(err)
	}
}

func clearPendingDecision(memberID string) {
	if redisClient == nil {
		return
	}
	if err := redisClient.Del(context.Background(), verifyDecisionKeyPrefix+memberID).Err(); err != nil {
		utils.ErrorLog.Println(err)
	}
}

// ReplayPendingVerifications applies decisions that were stored but never
// completed, typically because the bot stopped between the click and the
// Discord call. Call once the session is connected.
// ponytail: skips the welcome message and the moderator response, since both
// were tied to the interaction token that died with the previous process.
func ReplayPendingVerifications(s *discordgo.Session) {
	if redisClient == nil || s == nil {
		return
	}
	ctx := context.Background()
	// ponytail: KEYS instead of SCAN - the number of pending decisions is
	// bounded by the members awaiting verification. Switch to SCAN if that
	// ever stops being true.
	keys, err := redisClient.Keys(ctx, verifyDecisionKeyPrefix+"*").Result()
	if err != nil {
		utils.ErrorLog.Println(err)
		return
	}
	if len(keys) == 0 {
		return
	}
	utils.InfoLog.Printf("Replaying %d pending verification decision(s)\n", len(keys))

	for _, key := range keys {
		payload, err := redisClient.Get(ctx, key).Bytes()
		if err != nil {
			utils.ErrorLog.Println(err)
			continue
		}
		var decision verificationDecision
		if err := json.Unmarshal(payload, &decision); err != nil {
			utils.ErrorLog.Println(err)
			clearPendingDecision(strings.TrimPrefix(key, verifyDecisionKeyPrefix))
			continue
		}
		if err := applyDecision(s, decision); err != nil {
			// Kept for the next startup; the TTL bounds how long it retries.
			utils.ErrorLog.Printf("Failed to replay %s for %s: %v\n", decision.Action, decision.MemberID, err)
			continue
		}
		utils.InfoLog.Printf("Replayed %s for %s\n", decision.Action, decision.MemberID)
		clearPendingDecision(decision.MemberID)
	}
}

// applyDecision performs the Discord mutation for a decision. A member that is
// already gone or already verified counts as applied, so the replay stays
// idempotent.
func applyDecision(s *discordgo.Session, decision verificationDecision) error {
	unlock := lockVerification(decision.MemberID)
	defer unlock()

	member, err := s.GuildMember(decision.GuildID, decision.MemberID)
	if err != nil {
		if strings.Contains(err.Error(), HTTPNotFoundError) {
			return nil
		}
		return err
	}

	switch decision.Action {
	case verifyActionAccept:
		if slices.Contains(member.Roles, getVerifiedRoleId()) {
			return nil
		}
		return s.GuildMemberRoleAdd(decision.GuildID, decision.MemberID, getVerifiedRoleId())
	case verifyActionReject:
		err := s.GuildMemberDeleteWithReason(decision.GuildID, decision.MemberID, "Request rejected by admin")
		if err != nil && strings.Contains(err.Error(), HTTPNotFoundError) {
			return nil
		}
		return err
	default:
		return fmt.Errorf("unknown verification action %q", decision.Action)
	}
}

// moderatorID reports who clicked the button, when Discord sent it.
func moderatorID(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

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
		_, _ = s.ChannelMessageSendEmbed(logChannel, responses.CreateMessageEmbed(s,
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

// verificationButtons builds the Accept/Reject row for a member's request.
// disabled is set once a moderator has decided, so a second click cannot kick an
// already verified member (or vice versa).
func verificationButtons(memberID string, disabled bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Emoji:    &discordgo.ComponentEmoji{Name: "✅"},
					Label:    "Accept",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("acc-req-accept-%s", memberID),
					Disabled: disabled,
				},
				discordgo.Button{
					Emoji:    &discordgo.ComponentEmoji{Name: "❎"},
					Label:    "Reject",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("acc-req-reject-%s", memberID),
					Disabled: disabled,
				},
			},
		},
	}
}

// respondVerification decides the request. The interaction is acknowledged
// first: the flow makes several REST calls and Discord fails the interaction
// after 3 seconds, regardless of how long ago the request was posted.
func respondVerification(s *discordgo.Session, i *discordgo.InteractionCreate, memberId string, embed *discordgo.MessageEmbed) {
	if err := responses.InteractionResponse(s, i.Interaction).
		WithEmbed(embed).
		WithButton("Accept", discordgo.SuccessButton, fmt.Sprintf("acc-req-accept-%s", memberId), true).
		WithButton("Reject", discordgo.DangerButton, fmt.Sprintf("acc-req-reject-%s", memberId), true).
		SendDefer(); err != nil {
		utils.ErrorLog.Println(err)
	}
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
		Embed:      requestMessageEmbed,
		Components: verificationButtons(newMember.ID, false),
	})
	return err
}

// Handle when admin accept request by add "verify" role and send
// Welcome message to "welcome" channel.
func HandleVerificationAccept(s *discordgo.Session, i *discordgo.InteractionCreate, params ...any) {
	memberId := i.MessageComponentData().CustomID[strings.LastIndex(i.MessageComponentData().CustomID, "-")+StringOffsetAfterLastIndex:]
	verifiedRoleId := getVerifiedRoleId()

	// Ack before any REST call: Discord fails the interaction after 3 seconds,
	// no matter how old the request message is.
	if err := responses.InteractionResponse(s, i.Interaction).DeferUpdate(); err != nil {
		utils.ErrorLog.Println(err)
		return
	}

	// Lock after the ack so the 3 second window never waits on another click.
	unlock := lockVerification(memberId)
	defer unlock()

	fail := func(description string) {
		respondVerification(s, i, memberId, responses.CreateMessageEmbed(
			s, "Failed to process request", description, featureName, responses.SetColor("df0000")))
	}

	// Check member exists
	member, err := s.GuildMember(i.GuildID, memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		// When new member is not found or has left the server
		// before being approved
		fail("Sorry, your request failed to process because `member id` not found!")
		return
	}

	if slices.Contains(member.Roles, verifiedRoleId) {
		fail("Sorry, this member was verified!")
		return
	}

	savePendingDecision(verificationDecision{
		Action:      verifyActionAccept,
		MemberID:    memberId,
		ModeratorID: moderatorID(i),
		GuildID:     i.GuildID,
	})

	// Assign "verify" role to new member and send status response to moderator
	if err := s.GuildMemberRoleAdd(i.GuildID, memberId, verifiedRoleId); err != nil {
		utils.ErrorLog.Println(err)
		fail(fmt.Sprintf("Sorry, your request failed to process!\nDetail:\n%s", err.Error()))
		return
	}
	clearPendingDecision(memberId)

	respondVerification(s, i, memberId, responses.CreateMessageEmbed(
		s,
		"Accepted New Member Success",
		fmt.Sprintf(
			"%s has been processed to get channel access and assign “verified” role.",
			member.User.GlobalName),
		featureName,
		responses.SetColor("0bdd47"),
	))

	// Send welcome message to new member in "welcome" channel
	guildName := ""
	if guild, err := s.Guild(i.GuildID); err == nil {
		guildName = guild.Name
	}
	_, _ = s.ChannelMessageSendEmbed(
		cfg.WelcomeChannelID,
		responses.CreateMessageEmbed(
			s,
			fmt.Sprintf("Welcome to %s", guildName),
			fmt.Sprintf(`Hello <@%s>, welcome to %s.
				Please see the server rules at <#%s>.
				If you have any questions or suggestions, please ask \"Pengasuh Anak\"`,
				memberId, guildName, cfg.RulesChannelID,
			),
			featureName,
			responses.SetColor("0bdd47"),
		),
	)
}

// Handle when admin reject request then kick rejected new member
// from server and send DM to confirm to people who give
// him invitation link
func HandleVerificationReject(s *discordgo.Session, i *discordgo.InteractionCreate, params ...any) {
	memberId := i.MessageComponentData().CustomID[strings.LastIndex(i.MessageComponentData().CustomID, "-")+StringOffsetAfterLastIndex:]

	// Ack before any REST call: Discord fails the interaction after 3 seconds.
	if err := responses.InteractionResponse(s, i.Interaction).DeferUpdate(); err != nil {
		utils.ErrorLog.Println(err)
		return
	}

	// Lock after the ack so the 3 second window never waits on another click.
	unlock := lockVerification(memberId)
	defer unlock()

	fail := func(description string) {
		respondVerification(s, i, memberId, responses.CreateMessageEmbed(
			s, "Failed to process request", description, featureName, responses.SetColor("df0000")))
	}

	// Check member exists
	member, err := s.GuildMember(i.GuildID, memberId)
	if err != nil {
		utils.ErrorLog.Println(err)
		// When new member is not found or has left the server
		// before being approved
		fail("Sorry, your request failed to process because `member id` not found!")
		return
	}

	// Never kick a member who was already accepted: a stale Reject click would
	// otherwise undo an accepted verification.
	if slices.Contains(member.Roles, getVerifiedRoleId()) {
		fail("Sorry, this member was verified!")
		return
	}

	savePendingDecision(verificationDecision{
		Action:      verifyActionReject,
		MemberID:    memberId,
		ModeratorID: moderatorID(i),
		GuildID:     i.GuildID,
	})

	// Kick first: a member who blocked DMs still has to be removed, and a
	// failed kick must still be reported through the interaction.
	if err := s.GuildMemberDeleteWithReason(i.GuildID, memberId, "Request rejected by admin"); err != nil {
		utils.ErrorLog.Println(err)
		fail(fmt.Sprintf("Sorry, your request failed to process!\nDetail:\n%s", err.Error()))
		return
	}
	clearPendingDecision(memberId)

	// Tell the rejected member why, best effort - the decision is already applied.
	if userChannel, err := s.UserChannelCreate(memberId); err == nil {
		_, _ = s.ChannelMessageSendEmbed(
			userChannel.ID,
			responses.CreateMessageEmbed(
				s,
				"Verification Rejected",
				`Sorry, your request was rejected!
		Please contact the source of the invitation link for further confirmation`,
				featureName,
				responses.SetColor("df0000"),
			),
		)
	} else {
		utils.ErrorLog.Println(err)
	}

	respondVerification(s, i, memberId, responses.CreateMessageEmbed(
		s,
		"Rejected New Member Success",
		fmt.Sprintf(
			"%s has been rejected and kicked from server.",
			member.User.GlobalName),
		featureName,
		responses.SetColor("df0000"),
	))
}
