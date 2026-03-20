package moderation

import (
	"github.com/taufiq30s/chisa/utils"
)

type Config struct {
	VerificationChannelID string
	ModerationChannelID   string
	LogChannelID          string
	VerifiedRoleID        string
}

func LoadConfig() (*Config, error) {
	verificationChannelID, err := utils.GetEnv("VERIFICATION_CHANNEL_ID")
	if err != nil {
		return nil, err
	}
	moderationChannelID, err := utils.GetEnv("MODERATION_CHANNEL_ID")
	if err != nil {
		return nil, err
	}
	logChannelID, err := utils.GetEnv("LOG_CHANNEL_ID")
	if err != nil {
		return nil, err
	}
	verifiedRoleID, err := utils.GetEnv("VERIFIED_ROLE_ID")
	if err != nil {
		return nil, err
	}

	return &Config{
		VerificationChannelID: verificationChannelID,
		ModerationChannelID:   moderationChannelID,
		LogChannelID:          logChannelID,
		VerifiedRoleID:        verifiedRoleID,
	}, nil
}

// Magic number constants extracted for better code maintainability
const (
	// HTTP Status Codes
	// HTTPStatusOK represents the success response code from HTTP requests
	HTTPStatusOK = 200

	// Scam Detection Codes
	// ScamCodeNotScam indicates that the message is not a scam
	ScamCodeNotScam = 0
	// ScamCodeSuspect indicates that the message is suspected to be a scam (contains @everyone or @here)
	ScamCodeSuspect = 1
	// ScamCodePositive indicates that the message is positively identified as a scam (contains known scam links)
	ScamCodePositive = 2

	// Timeout Days Configuration
	// TimeoutDaysNone means no timeout is applied
	TimeoutDaysNone = 0
	// TimeoutDaysSuspect is the number of days to timeout a user for suspected scam (1 day)
	TimeoutDaysSuspect = 1
	// TimeoutDaysPositive is the number of days to timeout a user for confirmed scam (7 days)
	TimeoutDaysPositive = 7

	// String Processing Constants
	// StringOffsetAfterLastIndex is used when extracting userId from CustomID after LastIndex
	// This is needed because LastIndex returns the starting position, we need the character after
	StringOffsetAfterLastIndex = 1

	// Date Offset Constants
	// DateOffsetYears is used when no years offset is needed in AddDate
	DateOffsetYears = 0
	// DateOffsetMonths is used when no months offset is needed in AddDate
	DateOffsetMonths = 0

	// FindAll Constants
	// FindAllMatches is used with regex FindAllString to find all matches (no limit)
	FindAllMatches = -1

	// Guild Ban Constants
	// GuildBanDeleteMessageDays is the number of days of messages to delete when banning a user
	// 0 means don't delete any messages
	GuildBanDeleteMessageDays = 0

	// HTTP Error Patterns
	// HTTPNotFoundError is the error pattern to detect 404 Not Found errors
	HTTPNotFoundError = "404 Not Found"

	// Array Indices
	// FirstGuildIndex is the index of the first guild in State.Guilds array
	FirstGuildIndex = 0
)
