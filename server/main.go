package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

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
	MsgTypeCreateGame       = "create_game"
	MsgTypeJoinGame         = "join_game"
	MsgTypeMakeMove         = "make_move"
	MsgTypeGameState        = "game_state"
	MsgTypeError            = "error"
	MsgTypeGameList         = "game_list"
	MsgTypeAnswer           = "answer"
	MsgTypeCreateRoom       = "create_room"
	MsgTypeJoinRoom         = "join_room"
	MsgTypeJoinSpectator    = "join_spectator"
	MsgTypeLeaveRoom        = "leave_room"
	MsgTypeStartGame        = "start_game"
	MsgTypeRoomState        = "room_state"
	MsgTypePlayerJoined     = "player_joined"
	MsgTypePlayerLeft       = "player_left"
	MsgTypeChatMessage      = "chat_message"
	MsgTypeQuickMatch       = "quick_match"
	MsgTypeQuickMatchFound  = "quick_match_found"
	MsgTypeLeaderboard      = "leaderboard"
	MsgTypeTimeout          = "timeout"        // Move/answer timeout
	MsgTypeGameOver         = "game_over"      // Game ended due to timeout
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// TicTacToe game state
type TicTacToeGame struct {
	Board         [9]string    `json:"board"`
	Players       [2]string    `json:"players"`
	Turn          int          `json:"turn"`
	Winner        string       `json:"winner"`
	MoveHistory   []int        `json:"move_history"`
	GameMode      string       `json:"game_mode"`       // "fading" or "speed"
	LastMoveTime  time.Time    `json:"last_move_time"`   // For speed mode
	GameStartTime time.Time    `json:"game_start_time"`  // For speed mode
}

type TicTacToeMove struct {
	GameID string `json:"game_id"`
	Player string `json:"player"`
	Index  int    `json:"index"`
}

// Jeopardy game state
type JeopardyGame struct {
	Players          []string           `json:"players"`
	Scores           map[string]int     `json:"scores"`
	CurrentQ         int                `json:"current_q"`
	Questions        []JeopardyQuestion `json:"questions"`
	GameMode         string             `json:"game_mode"`          // "speed" for speed round
	QuestionStartTime time.Time         `json:"question_start_time"` // When current question was shown
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
	rooms          map[string]*Room
	clients        map[*websocket.Conn]*Client
	leaderboard    map[string]int
	quickMatch     []QuickMatchEntry
	mu             sync.RWMutex
}

type QuickMatchEntry struct {
	playerID string
	gameType string
	conn     *websocket.Conn
}

type Client struct {
	conn     *websocket.Conn
	playerID string
	roomCode string
}

type Room struct {
	Code       string            `json:"code"`
	Host       string            `json:"host"`
	Players    []string          `json:"players"`
	Spectators []string          `json:"spectators"`
	GameType   string            `json:"game_type"`
	GameMode   string            `json:"game_mode"`
	GameID     string            `json:"game_id,omitempty"`
	Status     string            `json:"status"` // "waiting", "playing"
	Password   string            `json:"password,omitempty"`
	IsPrivate  bool              `json:"is_private"`
	CreatedAt  time.Time         `json:"created_at"`
	LastActive time.Time         `json:"last_active"`
}

func newHub() *Hub {
	return &Hub{
		tictactoeGames: make(map[string]*TicTacToeGame),
		jeopardyGames:  make(map[string]*JeopardyGame),
		rooms:          make(map[string]*Room),
		clients:        make(map[*websocket.Conn]*Client),
		leaderboard:    make(map[string]int),
		quickMatch:     []QuickMatchEntry{},
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
	hub.clients[conn] = &Client{conn: conn, playerID: "", roomCode: ""}
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
	if client, exists := hub.clients[conn]; exists {
		// Remove player from room if in one
		if client.roomCode != "" {
			leaveRoom(client.playerID, client.roomCode)
		}
		delete(hub.clients, conn)
	}
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

		// Check for speed mode timeout (5 seconds per move)
		if game.GameMode == "speed" && !game.LastMoveTime.IsZero() {
			elapsed := time.Since(game.LastMoveTime)
			if elapsed > 5*time.Second {
				// Timeout - other player wins
				game.Winner = game.Players[1-game.Turn]
				hub.mu.RLock()
				for c := range hub.clients {
					sendMessage(c, MsgTypeGameOver, map[string]interface{}{
						"game_id": gameID,
						"game":    game,
						"reason":  "timeout",
						"winner":  game.Winner,
					})
				}
				hub.mu.RUnlock()
				return
			}
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

		// Fading Mode: Track move history and remove oldest after 4 moves
		if game.GameMode == "fading" {
			game.MoveHistory = append(game.MoveHistory, index)
			if len(game.MoveHistory) > 4 {
				// Remove the oldest move (first in the list)
				oldestMove := game.MoveHistory[0]
				game.MoveHistory = game.MoveHistory[1:]
				game.Board[oldestMove] = "" // Clear the oldest move from board
			}
		}

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

		// Update last move time for speed mode
		game.LastMoveTime = time.Now()

		game.Turn = 1 - game.Turn

		sendMessage(conn, MsgTypeGameState, map[string]interface{}{
			"game_id": gameID,
			"game":    game,
		})

	case MsgTypeCreateRoom:
		payload := msg.Payload.(map[string]interface{})
		gameType := payload["game_type"].(string)
		playerID := payload["player_id"].(string)
		gameMode := ""
		if gm, ok := payload["game_mode"].(string); ok {
			gameMode = gm
		}
		password := ""
		if pwd, ok := payload["password"].(string); ok {
			password = pwd
		}

		room := createRoom(playerID, gameType, gameMode, password)

		// Update client state
		hub.mu.Lock()
		if client, exists := hub.clients[conn]; exists {
			client.playerID = playerID
			client.roomCode = room.Code
		}
		hub.mu.Unlock()

		sendMessage(conn, MsgTypeRoomState, map[string]interface{}{
			"room": room,
		})

	case MsgTypeJoinRoom:
		payload := msg.Payload.(map[string]interface{})
		code := payload["code"].(string)
		playerID := payload["player_id"].(string)
		password := ""
		if pwd, ok := payload["password"].(string); ok {
			password = pwd
		}

		room, err := joinRoom(playerID, code, password)
		if err != nil {
			sendMessage(conn, MsgTypeError, err.Error())
			return
		}

		// Update client state
		hub.mu.Lock()
		if client, exists := hub.clients[conn]; exists {
			client.playerID = playerID
			client.roomCode = room.Code
		}
		hub.mu.Unlock()

		sendMessage(conn, MsgTypeRoomState, map[string]interface{}{
			"room": room,
		})

		// Broadcast to other players in room
		broadcastToRoom(room.Code, MsgTypePlayerJoined, map[string]interface{}{
			"player_id": playerID,
			"room":      room,
		})

	case MsgTypeLeaveRoom:
		payload := msg.Payload.(map[string]interface{})
		code := payload["code"].(string)
		playerID := payload["player_id"].(string)

		hub.mu.Lock()
		if client, exists := hub.clients[conn]; exists {
			client.roomCode = ""
		}
		hub.mu.Unlock()

		leaveRoom(playerID, code)

		sendMessage(conn, MsgTypeRoomState, map[string]interface{}{
			"room": nil,
		})

	case MsgTypeStartGame:
		payload := msg.Payload.(map[string]interface{})
		code := payload["code"].(string)
		playerID := payload["player_id"].(string)

		hub.mu.RLock()
		room, exists := hub.rooms[code]
		hub.mu.RUnlock()

		if !exists {
			sendMessage(conn, MsgTypeError, "Room not found")
			return
		}

		if room.Host != playerID {
			sendMessage(conn, MsgTypeError, "Only host can start the game")
			return
		}

		if len(room.Players) < 1 {
			sendMessage(conn, MsgTypeError, "Need at least 1 player")
			return
		}

		err := startGame(room)
		if err != nil {
			sendMessage(conn, MsgTypeError, err.Error())
			return
		}

		// Broadcast game start to all players
		hub.mu.RLock()
		for c, client := range hub.clients {
			if client.roomCode == code {
				// Get the game for this room
				gameID := room.GameID
				var game interface{}
				if room.GameType == "tictactoe" {
					game = hub.tictactoeGames[gameID]
				} else if room.GameType == "jeopardy" {
					game = hub.jeopardyGames[gameID]
				}
				sendMessage(c, MsgTypeGameState, map[string]interface{}{
					"game_id": gameID,
					"game":    game,
					"room":    room,
				})
			}
		}
		hub.mu.RUnlock()

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

		// Check for speed round timeout (10 seconds to answer)
		if game.GameMode == "speed" && !game.QuestionStartTime.IsZero() {
			elapsed := time.Since(game.QuestionStartTime)
			if elapsed > 10*time.Second {
				// Timeout - move to next question, no points
				game.CurrentQ++
				
				// Reset question timer for next question
				game.QuestionStartTime = time.Now()
				
				// Broadcast timeout to all players
				hub.mu.RLock()
				for c := range hub.clients {
					sendMessage(c, MsgTypeTimeout, map[string]interface{}{
						"game_id":  gameID,
						"game":     game,
						"reason":   "answer_timeout",
						"player":   playerID,
						"timeout":  true,
					})
				}
				hub.mu.RUnlock()
				return
			}
		}

		currentQ := game.Questions[game.CurrentQ]
		correct := strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(currentQ.Answer))

		if correct {
			game.Scores[playerID] += currentQ.Value
		}

		game.CurrentQ++

		// Reset question timer for speed mode after answering
		if game.GameMode == "speed" {
			game.QuestionStartTime = time.Now()
		}

		sendMessage(conn, MsgTypeGameState, map[string]interface{}{
			"game_id":  gameID,
			"game":     game,
			"correct":  correct,
			"answer":   currentQ.Answer,
			"question": currentQ,
		})

		// Update leaderboard for correct answers
		if correct {
			hub.mu.Lock()
			hub.leaderboard[playerID] += currentQ.Value
			hub.mu.Unlock()
		}

	case MsgTypeChatMessage:
		payload := msg.Payload.(map[string]interface{})
		roomCode := payload["room_code"].(string)
		playerID := payload["player_id"].(string)
		text := payload["text"].(string)

		// Broadcast chat message to all in room (including spectators)
		broadcastToRoom(roomCode, MsgTypeChatMessage, map[string]interface{}{
			"player_id": playerID,
			"text":      text,
			"timestamp": time.Now().Unix(),
		})

	case MsgTypeQuickMatch:
		payload := msg.Payload.(map[string]interface{})
		playerID := payload["player_id"].(string)
		gameType := payload["game_type"].(string)

		handleQuickMatch(conn, playerID, gameType)

	case MsgTypeLeaderboard:
		// Return top 10 players
		hub.mu.RLock()
		type leaderboardEntry struct {
			playerID string
			score    int
		}
		var entries []leaderboardEntry
		for pid, score := range hub.leaderboard {
			entries = append(entries, leaderboardEntry{playerID: pid, score: score})
		}
		hub.mu.RUnlock()

		// Sort by score descending
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[j].score > entries[i].score {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Take top 10
		if len(entries) > 10 {
			entries = entries[:10]
		}

		sendMessage(conn, MsgTypeLeaderboard, entries)
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
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateRoomCode() string {
	return strings.ToUpper(randomString(6))
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

// Room handling functions
func createRoom(playerID, gameType, gameMode, password string) *Room {
	code := generateRoomCode()
	isPrivate := password != ""
	room := &Room{
		Code:       code,
		Host:       playerID,
		Players:    []string{playerID},
		Spectators: []string{},
		GameType:   gameType,
		GameMode:   gameMode,
		Status:     "waiting",
		Password:   password,
		IsPrivate:  isPrivate,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	hub.mu.Lock()
	hub.rooms[code] = room
	hub.mu.Unlock()
	return room
}

func joinRoom(playerID, code, password string) (*Room, error) {
	// Validate room code format (6 uppercase chars)
	code = strings.ToUpper(code)
	if len(code) != 6 || code != strings.ToUpper(code) {
		return nil, fmt.Errorf("invalid room code format")
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()

	room, exists := hub.rooms[code]
	if !exists || room == nil {
		return nil, fmt.Errorf("room not found")
	}

	// Check password for private rooms
	if room.IsPrivate && room.Password != password {
		return nil, fmt.Errorf("invalid room password")
	}

	// Check if player already in room (as player)
	for _, p := range room.Players {
		if p == playerID {
			return room, nil
		}
	}

	// Check if player already in room (as spectator)
	for _, s := range room.Spectators {
		if s == playerID {
			return room, nil
		}
	}

	if room.Status == "playing" {
		// Can only join as spectator during gameplay
		room.Spectators = append(room.Spectators, playerID)
	} else {
		// Check player limit (max 8 players)
		if len(room.Players) >= 8 {
			return nil, fmt.Errorf("room is full (max 8 players)")
		}
		room.Players = append(room.Players, playerID)
	}

	room.LastActive = time.Now()
	return room, nil
}

func leaveRoom(playerID, code string) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	room, exists := hub.rooms[code]
	if !exists {
		return
	}

	// Remove player from players list
	newPlayers := []string{}
	for _, p := range room.Players {
		if p != playerID {
			newPlayers = append(newPlayers, p)
		}
	}

	// Remove player from spectators list
	newSpectators := []string{}
	for _, s := range room.Spectators {
		if s != playerID {
			newSpectators = append(newSpectators, s)
		}
	}

	if len(newPlayers) == 0 && len(newSpectators) == 0 || playerID == room.Host {
		// Delete room if empty or host left
		delete(hub.rooms, code)
	} else {
		room.Players = newPlayers
		room.Spectators = newSpectators
		room.LastActive = time.Now()
		// If host left, assign new host
		if playerID == room.Host && len(room.Players) > 0 {
			room.Host = room.Players[0]
		}
	}
}

func startGame(room *Room) error {
	if len(room.Players) < 1 {
		return fmt.Errorf("need at least 1 player")
	}

	gameID := generateGameID()
	room.GameID = gameID
	room.Status = "playing"
	room.LastActive = time.Now()

	if room.GameType == "tictactoe" {
		game := &TicTacToeGame{
			Board:         [9]string{},
			Players:       [2]string{},
			Turn:          0,
			Winner:        "",
			MoveHistory:   []int{},
			GameMode:      room.GameMode,
			LastMoveTime:  time.Time{},
			GameStartTime: time.Now(),
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.tictactoeGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "jeopardy" {
		// Collect all player IDs
		players := room.Players
		scores := make(map[string]int)

		for _, p := range players {
			scores[p] = 0
		}

		game := &JeopardyGame{
			Players:          players,
			Scores:           scores,
			CurrentQ:         0,
			Questions:        getJeopardyQuestions(),
			GameMode:         room.GameMode,
			QuestionStartTime: time.Time{},
		}

		// Set question start time for speed mode
		if room.GameMode == "speed" {
			game.QuestionStartTime = time.Now()
		}

		hub.mu.Lock()
		hub.jeopardyGames[gameID] = game
		hub.mu.Unlock()
	}

	return nil
}

func broadcastToRoom(code string, msgType string, payload interface{}) {
	hub.mu.RLock()
	room, exists := hub.rooms[code]
	hub.mu.RUnlock()

	if !exists {
		return
	}

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	for conn, client := range hub.clients {
		if client.roomCode == code {
			sendMessage(conn, msgType, payload)
		}
	}

	// Update room last active
	if room != nil {
		room.LastActive = time.Now()
	}
}

func handleQuickMatch(conn *websocket.Conn, playerID, gameType string) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Check if player is already in quick match queue
	for _, entry := range hub.quickMatch {
		if entry.playerID == playerID {
			sendMessage(conn, MsgTypeError, "Already in quick match queue")
			return
		}
	}

	// Add to queue
	hub.quickMatch = append(hub.quickMatch, QuickMatchEntry{
		playerID: playerID,
		gameType: gameType,
		conn:     conn,
	})

	// Try to find a match
	if len(hub.quickMatch) >= 2 {
		// Find two players with same game type
		for i := 0; i < len(hub.quickMatch)-1; i++ {
			for j := i + 1; j < len(hub.quickMatch); j++ {
				if hub.quickMatch[i].gameType == hub.quickMatch[j].gameType {
					// Found a match!
					player1 := hub.quickMatch[i]
					player2 := hub.quickMatch[j]

					// Create a room for them
					room := &Room{
						Code:       generateRoomCode(),
						Host:       player1.playerID,
						Players:    []string{player1.playerID, player2.playerID},
						Spectators: []string{},
						GameType:   player1.gameType,
						Status:     "waiting",
						CreatedAt:  time.Now(),
						LastActive: time.Now(),
					}

					hub.rooms[room.Code] = room

					// Notify both players
					sendMessage(player1.conn, MsgTypeQuickMatchFound, map[string]interface{}{
						"room":    room,
						"opponent": player2.playerID,
					})
					sendMessage(player2.conn, MsgTypeQuickMatchFound, map[string]interface{}{
						"room":    room,
						"opponent": player1.playerID,
					})

					// Remove both from queue
					hub.quickMatch = append(hub.quickMatch[:i], hub.quickMatch[i+1:]...)
					if j > i {
						hub.quickMatch = append(hub.quickMatch[:j-1], hub.quickMatch[j:]...)
					}

					return
				}
			}
		}
	}

	// No match found yet, tell player they're waiting
	sendMessage(conn, MsgTypeQuickMatch, map[string]interface{}{
		"status": "waiting",
	})
}

// Clean up rooms older than 30 minutes
func cleanupRooms() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		hub.mu.Lock()
		for code, room := range hub.rooms {
			if time.Since(room.LastActive) > 30*time.Minute {
				delete(hub.rooms, code)
				log.Printf("Room %s timed out and was deleted", code)
			}
		}
		hub.mu.Unlock()
	}
}

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Start room cleanup goroutine
	go cleanupRooms()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
