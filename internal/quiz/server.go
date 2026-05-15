package quiz

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/taufiq30s/chisa/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins
}

// StartServer starts the WebSocket + HTTP API server on 0.0.0.0:<port>.
// It fatals if the port is already in use.
// Pass the active QuizService so HTTP handlers can read/mutate session state.
func StartServer(quiz QuizService, port string) {
	addr := "0.0.0.0:" + port

	// Probe the port before binding so we can give a clear fatal message.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		utils.ErrorLog.Fatalf("Quiz server: port %s is already in use — %v\n", port, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", makeWSHandler(quiz))
	mux.HandleFunc("/api/promote", makePromoteHandler(quiz))
	mux.HandleFunc("/api/suppress", makeSuppressHandler(quiz))
	mux.HandleFunc("/api/queue", makeQueueHandler(quiz))

	utils.InfoLog.Printf("Quiz WS/HTTP server listening on %s\n", addr)
	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			utils.ErrorLog.Printf("Quiz server error: %v\n", err)
		}
	}()
}

// makeWSHandler returns an http.HandlerFunc that upgrades the connection to
// WebSocket, validates the session query param, and starts the write pump.
func makeWSHandler(quiz QuizService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session")
		if sessionID == "" || sessionID != quiz.GetSessionID() || !quiz.IsActive() {
			// Close with 4001 immediately after upgrade so the client can read the code.
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(wsCloseInvalidSession, "invalid or inactive session"))
			conn.Close()
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			utils.ErrorLog.Printf("Quiz WS upgrade error: %v\n", err)
			return
		}

		client := &Client{
			conn:      conn,
			sessionID: sessionID,
			send:      make(chan []byte, 32),
		}
		quiz.GetHub().register <- &ClientRegistration{client: client, sessionID: sessionID}

		// Send current snapshot immediately on connect.
		quiz.GetHub().BroadcastQueueUpdate(sessionID, quiz.GetSession().Snapshot())

		// Drain incoming frames (we don't expect client→server messages).
		go func() {
			defer func() {
				quiz.GetHub().unregister <- client
				conn.Close()
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		go client.writePump(quiz.GetHub())

		utils.InfoLog.Printf("Quiz WS client connected (session=%s)\n", fmt.Sprintf("%.8s…", sessionID))
	}
}
