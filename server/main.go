package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Game message types
const (
	MsgTypeCreateGame    = "create_game"
	MsgTypeJoinGame      = "join_game"
	MsgTypeMakeMove      = "make_move"
	MsgTypeGameState     = "game_state"
	MsgTypeError         = "error"
	MsgTypeGameList      = "game_list"
	MsgTypeAnswer        = "answer"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// TicTacToe game state
type TicTacToeGame struct {
	Board   [9]string `json:"board"`
	Players [2]string `json:"players"`
	Turn    int        `json:"turn"`
	Winner  string     `json:"winner"`
}

type TicTacToeMove struct {
	GameID string `json:"game_id"`
	Player string `json:"player"`
	Index  int    `json:"index"`
}

// Jeopardy game state
type JeopardyGame struct {
	Players   []string       `json:"players"`
	Scores    map[string]int `json:"scores"`
	CurrentQ  int            `json:"current_q"`
	Questions []JeopardyQuestion
}

type JeopardyQuestion struct {
	Category string `json:"category"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Value    int    `json:"value"`
}

type JeopardyAnswer struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Answer  string `json:"answer"`
}

// Hub maintains active games and connections
type Hub struct {
	tictactoeGames map[string]*TicTacToeGame
	jeopardyGames  map[string]*JeopardyGame
	clients       map[*websocket.Conn]string
	mu            sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		tictactoeGames: make(map[string]*TicTacToeGame),
		jeopardyGames:  make(map[string]*JeopardyGame),
		clients:        make(map[*websocket.Conn]string),
	}
}

var hub = newHub()

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	hub.mu.Lock()
	hub.clients[conn] = ""
	hub.mu.Unlock()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		handleMessage(conn, &msg)
	}

	hub.mu.Lock()
	delete(hub.clients, conn)
	hub.mu.Unlock()
}

func handleMessage(conn *websocket.Conn, msg *Message) {
	switch msg.Type {
	case MsgTypeCreateGame:
		payload := msg.Payload.(map[string]interface{})
		gameType := payload["game_type"].(string)
		playerID := payload["player_id"].(string)

		if gameType == "tictactoe" {
			gameID := generateGameID()
			game := &TicTacToeGame{
				Board:   [9]string{},
				Players: [2]string{playerID, ""},
				Turn:    0,
				Winner:  "",
			}
			hub.mu.Lock()
			hub.tictactoeGames[gameID] = game
			hub.mu.Unlock()

			sendMessage(conn, MsgTypeGameState, map[string]interface{}{
				"game_id": gameID,
				"game":    game,
			})
		} else if gameType == "jeopardy" {
			gameID := generateGameID()
			game := &JeopardyGame{
				Players:   []string{playerID},
				Scores:    map[string]int{playerID: 0},
				CurrentQ:  0,
				Questions: getJeopardyQuestions(),
			}
			hub.mu.Lock()
			hub.jeopardyGames[gameID] = game
			hub.mu.Unlock()

			sendMessage(conn, MsgTypeGameState, map[string]interface{}{
				"game_id": gameID,
				"game":    game,
			})
		}

	case MsgTypeJoinGame:
		payload := msg.Payload.(map[string]interface{})
		gameType := payload["game_type"].(string)
		gameID := payload["game_id"].(string)
		playerID := payload["player_id"].(string)

		if gameType == "tictactoe" {
			hub.mu.RLock()
			game, exists := hub.tictactoeGames[gameID]
			hub.mu.RUnlock()

			if !exists {
				sendMessage(conn, MsgTypeError, "Game not found")
				return
			}

			if game.Players[1] == "" {
				game.Players[1] = playerID
				sendMessage(conn, MsgTypeGameState, map[string]interface{}{
					"game_id": gameID,
					"game":    game,
				})
			} else {
				sendMessage(conn, MsgTypeError, "Game is full")
			}
		} else if gameType == "jeopardy" {
			hub.mu.RLock()
			game, exists := hub.jeopardyGames[gameID]
			hub.mu.RUnlock()

			if !exists {
				sendMessage(conn, MsgTypeError, "Game not found")
				return
			}

			game.Players = append(game.Players, playerID)
			game.Scores[playerID] = 0

			sendMessage(conn, MsgTypeGameState, map[string]interface{}{
				"game_id": gameID,
				"game":    game,
			})
		}

	case MsgTypeMakeMove:
		payload := msg.Payload.(map[string]interface{})
		gameID := payload["game_id"].(string)
		playerID := payload["player_id"].(string)
		index := int(payload["index"].(float64))

		hub.mu.RLock()
		game, exists := hub.tictactoeGames[gameID]
		hub.mu.RUnlock()

		if !exists {
			sendMessage(conn, MsgTypeError, "Game not found")
			return
		}

		playerIndex := -1
		for i, p := range game.Players {
			if p == playerID {
				playerIndex = i
				break
			}
		}

		if playerIndex == -1 || playerIndex != game.Turn {
			sendMessage(conn, MsgTypeError, "Not your turn")
			return
		}

		if game.Board[index] != "" {
			sendMessage(conn, MsgTypeError, "Cell already taken")
			return
		}

		symbols := []string{"X", "O"}
		game.Board[index] = symbols[playerIndex]

		// Check for winner
		winPatterns := [][]int{
			{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
			{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
			{0, 4, 8}, {2, 4, 6},             // diagonals
		}

		for _, pattern := range winPatterns {
			if game.Board[pattern[0]] != "" &&
				game.Board[pattern[0]] == game.Board[pattern[1]] &&
				game.Board[pattern[1]] == game.Board[pattern[2]] {
				game.Winner = playerID
			}
		}

		// Check for draw
		if game.Winner == "" {
			draw := true
			for _, cell := range game.Board {
				if cell == "" {
					draw = false
					break
				}
			}
			if draw {
				game.Winner = "draw"
			}
		}

		game.Turn = 1 - game.Turn

		sendMessage(conn, MsgTypeGameState, map[string]interface{}{
			"game_id": gameID,
			"game":    game,
		})

	case MsgTypeAnswer:
		payload := msg.Payload.(map[string]interface{})
		gameID := payload["game_id"].(string)
		playerID := payload["player_id"].(string)
		answer := payload["answer"].(string)

		hub.mu.RLock()
		game, exists := hub.jeopardyGames[gameID]
		hub.mu.RUnlock()

		if !exists {
			sendMessage(conn, MsgTypeError, "Game not found")
			return
		}

		if game.CurrentQ >= len(game.Questions) {
			sendMessage(conn, MsgTypeError, "No more questions")
			return
		}

		currentQ := game.Questions[game.CurrentQ]
		correct := answer == currentQ.Answer

		if correct {
			game.Scores[playerID] += currentQ.Value
		}

		game.CurrentQ++

		sendMessage(conn, MsgTypeGameState, map[string]interface{}{
			"game_id":  gameID,
			"game":     game,
			"correct":  correct,
			"answer":   currentQ.Answer,
			"question": currentQ,
		})
	}
}

func sendMessage(conn *websocket.Conn, msgType string, payload interface{}) {
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	conn.WriteJSON(msg)
}

func generateGameID() string {
	return "game_" + randomString(8)
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

func getJeopardyQuestions() []JeopardyQuestion {
	return []JeopardyQuestion{
		{Category: "Science", Question: "What is the chemical symbol for gold?", Answer: "Au", Value: 100},
		{Category: "Science", Question: "What planet is known as the Red Planet?", Answer: "Mars", Value: 100},
		{Category: "History", Question: "In what year did World War II end?", Answer: "1945", Value: 200},
		{Category: "History", Question: "Who was the first President of the United States?", Answer: "George Washington", Value: 200},
		{Category: "Geography", Question: "What is the capital of Japan?", Answer: "Tokyo", Value: 300},
		{Category: "Geography", Question: "What is the largest ocean on Earth?", Answer: "Pacific", Value: 300},
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
