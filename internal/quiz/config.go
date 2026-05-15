package quiz

const (
	featureName = "Interactive Quiz"

	// Embed colors
	colorStart = "0bdd47" // green — session started
	colorStop  = "c30010" // red   — session stopped
	colorInfo  = "5e11d9" // purple — informational

	// WebSocket close codes
	wsCloseInvalidSession = 4001 // session ID unknown or inactive
	wsCloseSessionEnded   = 4002 // moderator stopped the session

	// Environment variable name for the WS/HTTP server port
	QuizWSPortEnvKey     = "QUIZ_WS_PORT"
	DefaultQuizWSPort    = "8080"

	// HTTP header used for API authentication
	authHeader = "X-Session-Token"

	// Discord permission bit for MANAGE_CHANNELS
	permManageChannels = 0x0000000000000010
)
