package quiz

const (
	// WebSocket close codes
	wsCloseInvalidSession = 4001 // session ID unknown or inactive
	wsCloseSessionEnded   = 4002 // moderator stopped the session

	// Environment variable name for the WS/HTTP server port
	QuizWSPortEnvKey  = "QUIZ_WS_PORT"
	DefaultQuizWSPort = "8080"

	// HTTP header used for API authentication
	authHeader = "X-Session-Token"
)
