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

// Hangman game state
type HangmanGame struct {
	Players       [2]string    `json:"players"`
	Turn          int          `json:"turn"`
	Word          string       `json:"word"`
	GuessedLetters []string    `json:"guessed_letters"`
	WrongGuesses  int          `json:"wrong_guesses"`
	Winner        string       `json:"winner"`
	GameStartTime time.Time    `json:"game_start_time"`
}

type HangmanGuess struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Letter  string `json:"letter"`
}

// Memory game state
type MemoryGame struct {
	Players       []string       `json:"players"`
	Scores        map[string]int `json:"scores"`
	Cards         []MemoryCard   `json:"cards"`
	FlippedCards  []int          `json:"flipped_cards"`
	MatchedPairs  int            `json:"matched_pairs"`
	CurrentPlayer int            `json:"current_player"`
	GameStartTime time.Time      `json:"game_start_time"`
	CanFlip       bool           `json:"can_flip"`
	FirstFlip     int            `json:"first_flip"`
	GameOver      bool           `json:"game_over"`
}

type MemoryCard struct {
	Value    string `json:"value"`
	Flipped  bool   `json:"flipped"`
	Matched  bool   `json:"matched"`
}

type MemoryMove struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	CardIdx int    `json:"card_idx"`
}

// Battleship game state
type BattleshipGame struct {
	Players      [2]string           `json:"players"`
	Turn         int                 `json:"turn"`
	Grids        [2]BattleshipGrid   `json:"grids"`
	Winner       string              `json:"winner"`
	GameStartTime time.Time          `json:"game_start_time"`
	GamePhase    string              `json:"game_phase"` // "placing", "playing", "gameover"
}

type BattleshipGrid struct {
	Cells    [10][10]BattleshipCell `json:"cells"`
	Ships    []BattleshipShip       `json:"ships"`
	Shots    []BattleshipShot       `json:"shots"`
}

type BattleshipCell struct {
	HasShip bool   `json:"has_ship"`
	Hit     bool   `json:"hit"`
	Miss    bool   `json:"miss"`
}

type BattleshipShip struct {
	Type        string `json:"type"`
	Size        int    `json:"size"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Horizontal  bool   `json:"horizontal"`
	H           int    `json:"hits"`
}

type BattleshipShot struct {
	X     int  `json:"x"`
	Y     int  `json:"y"`
	Hit   bool `json:"hit"`
}

type BattleshipMove struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

// Wordle game state
type WordleGame struct {
	Players        [2]string      `json:"players"`
	Turn           int            `json:"turn"`
	TargetWord    string         `json:"target_word"`
	Guesses       []string       `json:"guesses"`
	GuessResults  [][]string     `json:"guess_results"` // "correct", "present", "absent"
	Winner        string         `json:"winner"`
	GameStartTime time.Time      `json:"game_start_time"`
	GameOver      bool            `json:"game_over"`
}

type WordleGuess struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Guess   string `json:"guess"`
}

// Trivia Quiz game state
type TriviaGame struct {
	Players          []string           `json:"players"`
	Scores           map[string]int     `json:"scores"`
	CurrentQ         int                `json:"current_q"`
	Questions        []TriviaQuestion   `json:"questions"`
	QuestionStartTime time.Time         `json:"question_start_time"`
	GameOver         bool               `json:"game_over"`
}

type TriviaQuestion struct {
	Category    string   `json:"category"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	CorrectIdx  int      `json:"correct_idx"`
}

type TriviaAnswer struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Idx     int    `json:"idx"`
}

// Rock Paper Scissors game state
type RPSGame struct {
	Players     [2]string `json:"players"`
	Turn        int       `json:"turn"`
	Moves       [2]string `json:"moves"`       // "", "rock", "paper", "scissors"
	Winner      string    `json:"winner"`
	BestOf      int       `json:"best_of"`     // 3, 5, or 7
	Scores      [2]int    `json:"scores"`
	RoundOver   bool      `json:"round_over"`
	GameOver    bool      `json:"game_over"`
	GameStartTime time.Time `json:"game_start_time"`
}

type RPSMove struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Move    string `json:"move"` // "rock", "paper", "scissors"
}

// Connect Four game state
type ConnectFourGame struct {
	Board         [6][7]string `json:"board"`
	Players       [2]string    `json:"players"`
	Turn          int          `json:"turn"`
	Winner        string       `json:"winner"`
	GameStartTime time.Time    `json:"game_start_time"`
}

type ConnectFourMove struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Column  int    `json:"column"`
}

// Checkers game state
type CheckersGame struct {
	Board         [8][8]CheckersPiece `json:"board"`
	Players       [2]string           `json:"players"`
	Turn          int                 `json:"turn"`
	Winner        string              `json:"winner"`
	GameStartTime time.Time           `json:"game_start_time"`
	ValidMoves    []CheckersMove      `json:"valid_moves"`
}

type CheckersPiece struct {
	Player   int    `json:"player"` // 0 = none, 1 = player1, 2 = player2
	King     bool   `json:"king"`
}

type CheckersMove struct {
	GameID    string `json:"game_id"`
	Player    string `json:"player"`
	FromRow   int    `json:"from_row"`
	FromCol   int    `json:"from_col"`
	ToRow     int    `json:"to_row"`
	ToCol     int    `json:"to_col"`
}

// 2048 game state (single player)
type Game2048 struct {
	PlayerID    string   `json:"player_id"`
	Board       [4][4]int `json:"board"`
	Score       int      `json:"score"`
	BestScore   int      `json:"best_score"`
	GameOver    bool     `json:"game_over"`
	Won         bool     `json:"won"`
	GameStartTime time.Time `json:"game_start_time"`
}

type Move2048 struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Dir     string `json:"dir"` // "up", "down", "left", "right"
}

// Dots and Boxes game state
type DotsBoxesGame struct {
	Players     [2]string            `json:"players"`
	Turn        int                  `json:"turn"`
	Board       DotsBoxesBoard       `json:"board"`
	Scores      [2]int               `json:"scores"`
	Boxes       []DotsBoxesBox       `json:"boxes"` // [][2]int for each box owner
	GameOver    bool                 `json:"game_over"`
	GameStartTime time.Time          `json:"game_start_time"`
}

type DotsBoxesBoard struct {
	Horizontal [6][7]bool `json:"horizontal"` // rows x cols
	Vertical   [7][6]bool `json:"vertical"`   // rows x cols
}

type DotsBoxesBox struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type DotsBoxesMove struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Type    string `json:"type"` // "horizontal" or "vertical"
	Row     int    `json:"row"`
	Col     int    `json:"col"`
}

// Hub maintains active games and connections
type Hub struct {
	tictactoeGames  map[string]*TicTacToeGame
	jeopardyGames   map[string]*JeopardyGame
	hangmanGames    map[string]*HangmanGame
	memoryGames    map[string]*MemoryGame
	battleshipGames map[string]*BattleshipGame
	wordleGames     map[string]*WordleGame
	triviaGames    map[string]*TriviaGame
	rpsGames       map[string]*RPSGame
	connectFourGames map[string]*ConnectFourGame
	checkersGames  map[string]*CheckersGame
	game2048       map[string]*Game2048
	dotsBoxesGames map[string]*DotsBoxesGame
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
		tictactoeGames:   make(map[string]*TicTacToeGame),
		jeopardyGames:    make(map[string]*JeopardyGame),
		hangmanGames:     make(map[string]*HangmanGame),
		memoryGames:     make(map[string]*MemoryGame),
		battleshipGames: make(map[string]*BattleshipGame),
		wordleGames:      make(map[string]*WordleGame),
		triviaGames:     make(map[string]*TriviaGame),
		rpsGames:        make(map[string]*RPSGame),
		connectFourGames: make(map[string]*ConnectFourGame),
		checkersGames:   make(map[string]*CheckersGame),
		game2048:        make(map[string]*Game2048),
		dotsBoxesGames:  make(map[string]*DotsBoxesGame),
		rooms:           make(map[string]*Room),
		clients:         make(map[*websocket.Conn]*Client),
		leaderboard:     make(map[string]int),
		quickMatch:      []QuickMatchEntry{},
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
		handleMakeMove(conn, msg)

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
				switch room.GameType {
				case "tictactoe":
					game = hub.tictactoeGames[gameID]
				case "jeopardy":
					game = hub.jeopardyGames[gameID]
				case "hangman":
					game = hub.hangmanGames[gameID]
				case "memory":
					game = hub.memoryGames[gameID]
				case "battleship":
					game = hub.battleshipGames[gameID]
				case "wordle":
					game = hub.wordleGames[gameID]
				case "trivia":
					game = hub.triviaGames[gameID]
				case "rps":
					game = hub.rpsGames[gameID]
				case "connectfour":
					game = hub.connectFourGames[gameID]
				case "checkers":
					game = hub.checkersGames[gameID]
				case "2048":
					game = hub.game2048[gameID]
				case "dotsboxes":
					game = hub.dotsBoxesGames[gameID]
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
	} else if room.GameType == "hangman" {
		words := []string{"GALAXY", "PLANET", "ORBIT", "COMET", "ASTRO", "NEBULA", "STARS", "MOON", "SPACE", "ROCKET"}
		word := words[rand.Intn(len(words))]
		game := &HangmanGame{
			Players:        [2]string{},
			Turn:           0,
			Word:           word,
			GuessedLetters: []string{},
			WrongGuesses:   0,
			Winner:         "",
			GameStartTime:  time.Now(),
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.hangmanGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "memory" {
		players := room.Players
		scores := make(map[string]int)
		for _, p := range players {
			scores[p] = 0
		}
		emojis := []string{"🚀", "🌟", "🎮", "🎲", "🎯", "🏆", "🎪", "🎭"}
		cards := []MemoryCard{}
		for _, emoji := range emojis {
			cards = append(cards, MemoryCard{Value: emoji, Flipped: false, Matched: false})
			cards = append(cards, MemoryCard{Value: emoji, Flipped: false, Matched: false})
		}
		// Shuffle
		for i := len(cards) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			cards[i], cards[j] = cards[j], cards[i]
		}
		game := &MemoryGame{
			Players:       players,
			Scores:        scores,
			Cards:         cards,
			FlippedCards:  []int{},
			MatchedPairs:  0,
			CurrentPlayer: 0,
			GameStartTime: time.Now(),
			CanFlip:       true,
			FirstFlip:     -1,
			GameOver:      false,
		}

		hub.mu.Lock()
		hub.memoryGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "battleship" {
		game := &BattleshipGame{
			Players:      [2]string{},
			Turn:         0,
			Grids:        [2]BattleshipGrid{},
			Winner:       "",
			GameStartTime: time.Now(),
			GamePhase:    "placing",
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.battleshipGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "wordle" {
		words := []string{"WORLD", "SPACE", "GAMES", "PLAY", "WIN", "GUESS", "SMART", "QUICK", "LOGIC", "BRAIN"}
		word := words[rand.Intn(len(words))]
		game := &WordleGame{
			Players:       [2]string{},
			Turn:          0,
			TargetWord:    word,
			Guesses:       []string{},
			GuessResults:  [][]string{},
			Winner:        "",
			GameStartTime: time.Now(),
			GameOver:      false,
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.wordleGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "trivia" {
		players := room.Players
		scores := make(map[string]int)
		for _, p := range players {
			scores[p] = 0
		}
		game := &TriviaGame{
			Players:          players,
			Scores:           scores,
			CurrentQ:         0,
			Questions:        getTriviaQuestions(),
			QuestionStartTime: time.Now(),
			GameOver:         false,
		}

		hub.mu.Lock()
		hub.triviaGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "rps" {
		game := &RPSGame{
			Players:       [2]string{},
			Turn:          0,
			Moves:         [2]string{},
			Winner:        "",
			BestOf:        3,
			Scores:        [2]int{0, 0},
			RoundOver:     false,
			GameOver:      false,
			GameStartTime: time.Now(),
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.rpsGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "connectfour" {
		game := &ConnectFourGame{
			Board:         [6][7]string{},
			Players:       [2]string{},
			Turn:          0,
			Winner:        "",
			GameStartTime: time.Now(),
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.connectFourGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "checkers" {
		board := [8][8]CheckersPiece{}
		// Initialize pieces
		for row := 0; row < 8; row++ {
			for col := 0; col < 8; col++ {
				if (row+col)%2 == 1 {
					if row < 3 {
						board[row][col] = CheckersPiece{Player: 2, King: false}
					} else if row > 4 {
						board[row][col] = CheckersPiece{Player: 1, King: false}
					}
				}
			}
		}
		game := &CheckersGame{
			Board:         board,
			Players:       [2]string{},
			Turn:          0,
			Winner:        "",
			GameStartTime: time.Now(),
			ValidMoves:    []CheckersMove{},
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.checkersGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "2048" {
		board := [4][4]int{}
		// Add two random tiles
		addRandomTile2048(&board)
		addRandomTile2048(&board)
		game := &Game2048{
			PlayerID:     room.Players[0],
			Board:        board,
			Score:        0,
			BestScore:    0,
			GameOver:     false,
			Won:          false,
			GameStartTime: time.Now(),
		}

		hub.mu.Lock()
		hub.game2048[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "dotsboxes" {
		game := &DotsBoxesGame{
			Players:     [2]string{},
			Turn:        0,
			Board:       DotsBoxesBoard{},
			Scores:      [2]int{0, 0},
			Boxes:       []DotsBoxesBox{},
			GameOver:    false,
			GameStartTime: time.Now(),
		}
		if len(room.Players) >= 1 {
			game.Players[0] = room.Players[0]
		}
		if len(room.Players) >= 2 {
			game.Players[1] = room.Players[1]
		}

		hub.mu.Lock()
		hub.dotsBoxesGames[gameID] = game
		hub.mu.Unlock()
	}

	return nil
}

func handleMakeMove(conn *websocket.Conn, msg *Message) {
	payload := msg.Payload.(map[string]interface{})
	gameID := payload["game_id"].(string)
	playerID := payload["player_id"].(string)

	// Determine game type by checking each game map
	gameType := ""

	hub.mu.RLock()
	for _, r := range hub.rooms {
		if r.GameID == gameID {
			gameType = r.GameType
			break
		}
	}
	hub.mu.RUnlock()

	if gameType == "" {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	switch gameType {
	case "tictactoe":
		handleTicTacToeMove(conn, gameID, playerID, payload)
	case "hangman":
		handleHangmanMove(conn, gameID, playerID, payload)
	case "memory":
		handleMemoryMove(conn, gameID, playerID, payload)
	case "battleship":
		handleBattleshipMove(conn, gameID, playerID, payload)
	case "wordle":
		handleWordleMove(conn, gameID, playerID, payload)
	case "trivia":
		handleTriviaAnswer(conn, gameID, playerID, payload)
	case "rps":
		handleRPSMove(conn, gameID, playerID, payload)
	case "connectfour":
		handleConnectFourMove(conn, gameID, playerID, payload)
	case "checkers":
		handleCheckersMove(conn, gameID, playerID, payload)
	case "2048":
		handle2048Move(conn, gameID, playerID, payload)
	case "dotsboxes":
		handleDotsBoxesMove(conn, gameID, playerID, payload)
	default:
		sendMessage(conn, MsgTypeError, "Unknown game type")
	}
}

func handleTicTacToeMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	index := int(payload["index"].(float64))

	hub.mu.RLock()
	game, exists := hub.tictactoeGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.Winner != "" {
		sendMessage(conn, MsgTypeError, "Game already over")
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
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
		{0, 4, 8}, {2, 4, 6},
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

	broadcastGameState(gameID, "tictactoe", game)
}

func handleHangmanMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	letter := strings.ToUpper(payload["letter"].(string))

	hub.mu.RLock()
	game, exists := hub.hangmanGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.Winner != "" {
		sendMessage(conn, MsgTypeError, "Game already over")
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

	// Check if letter already guessed
	for _, l := range game.GuessedLetters {
		if l == letter {
			sendMessage(conn, MsgTypeError, "Letter already guessed")
			return
		}
	}

	game.GuessedLetters = append(game.GuessedLetters, letter)

	// Check if letter is in word
	found := false
	for _, c := range game.Word {
		if string(c) == letter {
			found = true
			break
		}
	}

	if !found {
		game.WrongGuesses++
		// Check if player lost (6 wrong guesses max)
		if game.WrongGuesses >= 6 {
			game.Winner = "lose"
		}
	}

	// Check if word is complete
	if game.Winner == "" {
		complete := true
		for _, c := range game.Word {
			found := false
			for _, l := range game.GuessedLetters {
				if string(c) == l {
					found = true
					break
				}
			}
			if !found {
				complete = false
				break
			}
		}
		if complete {
			game.Winner = playerID
		}
	}

	if game.Winner == "" {
		game.Turn = 1 - game.Turn
	}

	broadcastGameState(gameID, "hangman", game)
}

func handleMemoryMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	cardIdx := int(payload["card_idx"].(float64))

	hub.mu.RLock()
	game, exists := hub.memoryGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
		return
	}

	if !game.CanFlip {
		sendMessage(conn, MsgTypeError, "Wait for animation")
		return
	}

	playerIndex := -1
	for i, p := range game.Players {
		if p == playerID {
			playerIndex = i
			break
		}
	}

	if playerIndex == -1 || playerIndex != game.CurrentPlayer {
		sendMessage(conn, MsgTypeError, "Not your turn")
		return
	}

	if cardIdx < 0 || cardIdx >= len(game.Cards) {
		sendMessage(conn, MsgTypeError, "Invalid card")
		return
	}

	if game.Cards[cardIdx].Flipped || game.Cards[cardIdx].Matched {
		sendMessage(conn, MsgTypeError, "Card already flipped")
		return
	}

	game.Cards[cardIdx].Flipped = true
	game.FlippedCards = append(game.FlippedCards, cardIdx)

	if game.FirstFlip == -1 {
		game.FirstFlip = cardIdx
		game.CanFlip = false
	} else {
		// Second card flipped
		firstCard := game.Cards[game.FirstFlip]
		secondCard := game.Cards[cardIdx]

		if firstCard.Value == secondCard.Value {
			// Match!
			game.Cards[game.FirstFlip].Matched = true
			game.Cards[cardIdx].Matched = true
			game.MatchedPairs++
			game.Scores[playerID]++

			// Clear flipped cards
			game.FlippedCards = []int{}
			game.FirstFlip = -1
			game.CanFlip = true

			// Check if game over
			if game.MatchedPairs >= len(game.Cards)/2 {
				game.GameOver = true
			}
		} else {
			// No match - turn passes
			game.CanFlip = false
			game.CurrentPlayer = 1 - game.CurrentPlayer
			// Reset flipped after delay (handled client-side)
		}
	}

	broadcastGameState(gameID, "memory", game)
}

func handleBattleshipMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	x := int(payload["x"].(float64))
	y := int(payload["y"].(float64))

	hub.mu.RLock()
	game, exists := hub.battleshipGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GamePhase != "playing" {
		sendMessage(conn, MsgTypeError, "Not in playing phase")
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

	// Fire at opponent's grid
	opponentIndex := 1 - playerIndex
	grid := &game.Grids[opponentIndex]

	if x < 0 || x >= 10 || y < 0 || y >= 10 {
		sendMessage(conn, MsgTypeError, "Invalid coordinates")
		return
	}

	if grid.Cells[y][x].Hit || grid.Cells[y][x].Miss {
		sendMessage(conn, MsgTypeError, "Already fired here")
		return
	}

	shot := BattleshipShot{X: x, Y: y, Hit: grid.Cells[y][x].HasShip}
	grid.Shots = append(grid.Shots, shot)

	if grid.Cells[y][x].HasShip {
		grid.Cells[y][x].Hit = true
		// Check if all ships sunk
		allSunk := true
		for _, ship := range grid.Ships {
			if ship.H < ship.Size {
				allSunk = false
				break
			}
		}
		if allSunk {
			game.Winner = playerID
			game.GamePhase = "gameover"
		}
	} else {
		grid.Cells[y][x].Miss = true
	}

	if game.Winner == "" {
		game.Turn = 1 - game.Turn
	}

	broadcastGameState(gameID, "battleship", game)
}

func handleWordleMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	guess := strings.ToUpper(payload["guess"].(string))

	hub.mu.RLock()
	game, exists := hub.wordleGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
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

	// Validate guess length
	if len(guess) != 5 {
		sendMessage(conn, MsgTypeError, "Guess must be 5 letters")
		return
	}

	// Check if already guessed
	for _, g := range game.Guesses {
		if g == guess {
			sendMessage(conn, MsgTypeError, "Already guessed")
			return
		}
	}

	// Check if valid word (simplified - just check length)
	game.Guesses = append(game.Guesses, guess)

	// Calculate result
	result := make([]string, 5)
	wordChars := []rune(game.TargetWord)
	guessChars := []rune(guess)
	used := make([]bool, 5)

	// First pass: correct position
	for i := 0; i < 5; i++ {
		if guessChars[i] == wordChars[i] {
			result[i] = "correct"
			used[i] = true
		}
	}

	// Second pass: wrong position
	for i := 0; i < 5; i++ {
		if result[i] == "correct" {
			continue
		}
		found := false
		for j := 0; j < 5; j++ {
			if !used[j] && guessChars[i] == wordChars[j] {
				result[i] = "present"
				used[j] = true
				found = true
				break
			}
		}
		if !found {
			result[i] = "absent"
		}
	}

	game.GuessResults = append(game.GuessResults, result)

	// Check win
	if guess == game.TargetWord {
		game.Winner = playerID
		game.GameOver = true
	}

	// Check loss (6 guesses used)
	if game.Winner == "" && len(game.Guesses) >= 6 {
		game.GameOver = true
		game.Winner = "lose"
	}

	if game.Winner == "" {
		game.Turn = 1 - game.Turn
	}

	broadcastGameState(gameID, "wordle", game)
}

func handleTriviaAnswer(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	idx := int(payload["idx"].(float64))

	hub.mu.RLock()
	game, exists := hub.triviaGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
		return
	}

	if game.CurrentQ >= len(game.Questions) {
		sendMessage(conn, MsgTypeError, "No more questions")
		return
	}

	currentQ := game.Questions[game.CurrentQ]
	correct := idx == currentQ.CorrectIdx

	if correct {
		game.Scores[playerID] += 100
	}

	game.CurrentQ++

	// Check if game over
	if game.CurrentQ >= len(game.Questions) {
		game.GameOver = true
	}

	broadcastGameState(gameID, "trivia", game)
}

func handleRPSMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	move := payload["move"].(string)

	hub.mu.RLock()
	game, exists := hub.rpsGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
		return
	}

	playerIndex := -1
	for i, p := range game.Players {
		if p == playerID {
			playerIndex = i
			break
		}
	}

	if playerIndex == -1 {
		sendMessage(conn, MsgTypeError, "Not a player")
		return
	}

	if game.Moves[playerIndex] != "" {
		sendMessage(conn, MsgTypeError, "Already played this round")
		return
	}

	game.Moves[playerIndex] = move

	// Check if both played
	if game.Moves[0] != "" && game.Moves[1] != "" {
		game.RoundOver = true

		// Determine winner
		m0, m1 := game.Moves[0], game.Moves[1]

		if m0 == m1 {
			// Tie - no points
		} else if (m0 == "rock" && m1 == "scissors") ||
			(m0 == "paper" && m1 == "rock") ||
			(m0 == "scissors" && m1 == "paper") {
			game.Scores[0]++
		} else {
			game.Scores[1]++
		}

		// Check if game over (reached best of)
		if game.Scores[0] > game.BestOf/2 || game.Scores[1] > game.BestOf/2 {
			game.GameOver = true
			if game.Scores[0] > game.Scores[1] {
				game.Winner = game.Players[0]
			} else if game.Scores[1] > game.Scores[0] {
				game.Winner = game.Players[1]
			} else {
				game.Winner = "draw"
			}
		}
	}

	broadcastGameState(gameID, "rps", game)
}

func handleConnectFourMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	col := int(payload["column"].(float64))

	hub.mu.RLock()
	game, exists := hub.connectFourGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.Winner != "" {
		sendMessage(conn, MsgTypeError, "Game already over")
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

	if col < 0 || col >= 7 {
		sendMessage(conn, MsgTypeError, "Invalid column")
		return
	}

	// Find lowest empty row
	row := -1
	for r := 5; r >= 0; r-- {
		if game.Board[r][col] == "" {
			row = r
			break
		}
	}

	if row == -1 {
		sendMessage(conn, MsgTypeError, "Column full")
		return
	}

	symbols := []string{"🔴", "🟡"}
	game.Board[row][col] = symbols[playerIndex]

	// Check for winner
	winner := checkConnectFourWinner(game.Board, col, row)
	if winner != "" {
		game.Winner = playerID
	}

	// Check for draw (board full)
	if game.Winner == "" {
		full := true
		for r := 0; r < 6; r++ {
			for c := 0; c < 7; c++ {
				if game.Board[r][c] == "" {
					full = false
					break
				}
			}
		}
		if full {
			game.Winner = "draw"
		}
	}

	if game.Winner == "" {
		game.Turn = 1 - game.Turn
	}

	broadcastGameState(gameID, "connectfour", game)
}

func handleCheckersMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	fromRow := int(payload["from_row"].(float64))
	fromCol := int(payload["from_col"].(float64))
	toRow := int(payload["to_row"].(float64))
	toCol := int(payload["to_col"].(float64))

	hub.mu.RLock()
	game, exists := hub.checkersGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.Winner != "" {
		sendMessage(conn, MsgTypeError, "Game already over")
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

	// Validate move
	if fromRow < 0 || fromRow >= 8 || fromCol < 0 || fromCol >= 8 ||
		toRow < 0 || toRow >= 8 || toCol < 0 || toCol >= 8 {
		sendMessage(conn, MsgTypeError, "Invalid coordinates")
		return
	}

	piece := game.Board[fromRow][fromCol]
	if piece.Player != playerIndex+1 {
		sendMessage(conn, MsgTypeError, "No piece there")
		return
	}

	// Simple move validation
	dr := toRow - fromRow
	dc := toCol - fromCol

	if abs(dr) != 1 || abs(dc) != 1 {
		// Could be a jump
		if abs(dr) == 2 && abs(dc) == 2 {
			midR := (fromRow + toRow) / 2
			midC := (fromCol + toCol) / 2
			midPiece := game.Board[midR][midC]
			if midPiece.Player == 0 || midPiece.Player == playerIndex+1 {
				sendMessage(conn, MsgTypeError, "Invalid jump")
				return
			}
			// Capture!
			game.Board[midR][midC] = CheckersPiece{}
		} else {
			sendMessage(conn, MsgTypeError, "Invalid move")
			return
		}
	}

	// Check direction
	if !piece.King {
		if playerIndex == 0 && dr > 0 {
			sendMessage(conn, MsgTypeError, "Can only move forward")
			return
		}
		if playerIndex == 1 && dr < 0 {
			sendMessage(conn, MsgTypeError, "Can only move forward")
			return
		}
	}

	// Move piece
	game.Board[toRow][toCol] = piece
	game.Board[fromRow][fromCol] = CheckersPiece{}

	// Check for king
	if playerIndex == 0 && toRow == 7 {
		game.Board[toRow][toCol].King = true
	} else if playerIndex == 1 && toRow == 0 {
		game.Board[toRow][toCol].King = true
	}

	// Check for win (all opponent pieces captured)
	oppCount := 0
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if game.Board[r][c].Player == playerIndex+1 {
				oppCount++
			}
		}
	}
	if oppCount == 0 {
		game.Winner = playerID
	}

	if game.Winner == "" {
		game.Turn = 1 - game.Turn
		game.ValidMoves = getCheckersValidMoves(game.Board, game.Turn+1)
	}

	broadcastGameState(gameID, "checkers", game)
}

func handle2048Move(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	dir := payload["dir"].(string)

	hub.mu.RLock()
	game, exists := hub.game2048[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
		return
	}

	if game.PlayerID != playerID {
		sendMessage(conn, MsgTypeError, "Not your game")
		return
	}

	// Make move
	moved := move2048(&game.Board, dir)
	if moved {
		addRandomTile2048(&game.Board)
	}

	// Check game over
	if isGameOver2048(game.Board) {
		game.GameOver = true
	}

	// Update score
	score := 0
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			score += game.Board[r][c]
		}
	}
	game.Score = score

	if game.Score > game.BestScore {
		game.BestScore = game.Score
	}

	broadcastGameState(gameID, "2048", game)
}

func handleDotsBoxesMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	moveType := payload["type"].(string)
	row := int(payload["row"].(float64))
	col := int(payload["col"].(float64))

	hub.mu.RLock()
	game, exists := hub.dotsBoxesGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
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

	// Validate move
	if moveType == "horizontal" {
		if row < 0 || row > 5 || col < 0 || col > 6 {
			sendMessage(conn, MsgTypeError, "Invalid position")
			return
		}
		if game.Board.Horizontal[row][col] {
			sendMessage(conn, MsgTypeError, "Already drawn")
			return
		}
		game.Board.Horizontal[row][col] = true
	} else if moveType == "vertical" {
		if row < 0 || row > 6 || col < 0 || col > 5 {
			sendMessage(conn, MsgTypeError, "Invalid position")
			return
		}
		if game.Board.Vertical[row][col] {
			sendMessage(conn, MsgTypeError, "Already drawn")
			return
		}
		game.Board.Vertical[row][col] = true
	} else {
		sendMessage(conn, MsgTypeError, "Invalid move type")
		return
	}

	// Check for completed boxes
	completed := 0
	// Check horizontal lines for boxes above
	for r := 0; r < 5; r++ {
		for c := 0; c < 6; c++ {
			if game.Board.Horizontal[r][c] && game.Board.Horizontal[r+1][c] && 
			   game.Board.Vertical[r][c] && game.Board.Vertical[r][c+1] {
				boxOwned := false
				for _, box := range game.Boxes {
					if box.Row == r && box.Col == c {
						boxOwned = true
						break
					}
				}
				if !boxOwned {
					game.Boxes = append(game.Boxes, DotsBoxesBox{Row: r, Col: c})
					completed++
				}
			}
		}
	}

	if completed > 0 {
		game.Scores[playerIndex] += completed
		// Player gets another turn
	} else {
		game.Turn = 1 - game.Turn
	}

	// Check game over (all boxes filled = 36 boxes for 6x6 grid)
	if len(game.Boxes) >= 36 {
		game.GameOver = true
	}

	broadcastGameState(gameID, "dotsboxes", game)
}

// Helper functions

func broadcastGameState(gameID string, gameType string, game interface{}) {
	var roomCode string
	hub.mu.RLock()
	for code, room := range hub.rooms {
		if room.GameID == gameID {
			roomCode = code
			break
		}
	}
	hub.mu.RUnlock()

	if roomCode == "" {
		return
	}

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	for c := range hub.clients {
		if c != nil {
			sendMessage(c, MsgTypeGameState, map[string]interface{}{
				"game_id": gameID,
				"game":    game,
			})
		}
	}
}

func move2048(board *[4][4]int, dir string) bool {
	moved := false
	
	rotateBoard := func(b *[4][4]int) [4][4]int {
		newBoard := [4][4]int{}
		for r := 0; r < 4; r++ {
			for c := 0; c < 4; c++ {
				newBoard[c][3-r] = (*b)[r][c]
			}
		}
		return newBoard
	}

	rotateBoardBack := func(b *[4][4]int) [4][4]int {
		newBoard := [4][4]int{}
		for r := 0; r < 4; r++ {
			for c := 0; c < 4; c++ {
				newBoard[3-c][r] = (*b)[r][c]
			}
		}
		return newBoard
	}

	slideLeft := func(b *[4][4]int) bool {
		m := false
		for r := 0; r < 4; r++ {
			// Remove zeros
			row := []int{}
			for c := 0; c < 4; c++ {
				if (*b)[r][c] != 0 {
					row = append(row, (*b)[r][c])
				}
			}
			// Merge
			newRow := []int{}
			i := 0
			for i < len(row) {
				if i+1 < len(row) && row[i] == row[i+1] {
					newRow = append(newRow, row[i]*2)
					i += 2
				} else {
					newRow = append(newRow, row[i])
					i++
				}
			}
			// Pad with zeros
			for len(newRow) < 4 {
				newRow = append(newRow, 0)
			}
			// Check if changed
			for c := 0; c < 4; c++ {
				if (*b)[r][c] != newRow[c] {
					m = true
				}
				(*b)[r][c] = newRow[c]
			}
		}
		return m
	}

	// Rotate board to always slide left
	temp := *board
	switch dir {
	case "left":
		moved = slideLeft(board)
	case "right":
		temp = rotateBoard(board)
		temp = rotateBoard(&temp)
		moved = slideLeft(&temp)
		*board = rotateBoardBack(&temp)
		*board = rotateBoardBack(board)
	case "up":
		temp = rotateBoard(board)
		moved = slideLeft(&temp)
		*board = rotateBoardBack(&temp)
	case "down":
		temp = rotateBoard(board)
		temp = rotateBoard(&temp)
		temp = rotateBoard(&temp)
		moved = slideLeft(&temp)
		*board = rotateBoardBack(&temp)
	}

	return moved
}

func isGameOver2048(board [4][4]int) bool {
	// Check for empty cells
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if board[r][c] == 0 {
				return false
			}
		}
	}

	// Check for possible merges
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if c < 3 && board[r][c] == board[r][c+1] {
				return false
			}
			if r < 3 && board[r][c] == board[r+1][c] {
				return false
			}
		}
	}

	return true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
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

// Helper functions for new games

func getTriviaQuestions() []TriviaQuestion {
	return []TriviaQuestion{
		{Category: "Science", Question: "What is H2O?", Options: []string{"Gold", "Water", "Silver", "Oxygen"}, CorrectIdx: 1},
		{Category: "Science", Question: "How many planets in solar system?", Options: []string{"7", "8", "9", "10"}, CorrectIdx: 1},
		{Category: "History", Question: "Year US independence?", Options: []string{"1776", "1789", "1492", "1812"}, CorrectIdx: 0},
		{Category: "History", Question: "First US President?", Options: []string{"Lincoln", "Jefferson", "Washington", "Adams"}, CorrectIdx: 2},
		{Category: "Geography", Question: "Capital of France?", Options: []string{"London", "Berlin", "Madrid", "Paris"}, CorrectIdx: 3},
		{Category: "Geography", Question: "Largest ocean?", Options: []string{"Atlantic", "Indian", "Arctic", "Pacific"}, CorrectIdx: 3},
	}
}

func addRandomTile2048(board *[4][4]int) {
	empty := [][2]int{}
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if (*board)[r][c] == 0 {
				empty = append(empty, [2]int{r, c})
			}
		}
	}
	if len(empty) > 0 {
		pos := empty[rand.Intn(len(empty))]
		if rand.Float32() < 0.9 {
			(*board)[pos[0]][pos[1]] = 2
		} else {
			(*board)[pos[0]][pos[1]] = 4
		}
	}
}

func checkConnectFourWinner(board [6][7]string, col int, row int) string {
	// Check vertical
	if row >= 3 {
		if board[row-1][col] != "" && board[row-1][col] == board[row-2][col] && board[row-2][col] == board[row-3][col] {
			return board[row][col]
		}
	}
	// Check horizontal
	for c := 0; c <= col-3; c++ {
		if board[row][c] != "" && board[row][c] == board[row][c+1] && board[row][c+1] == board[row][c+2] && board[row][c+2] == board[row][c+3] {
			return board[row][c]
		}
	}
	// Check diagonal down-right
	for r, c := row-3, col-3; r+3 <= row && c+3 <= col; r, c = r+1, c+1 {
		if board[r][c] != "" && board[r][c] == board[r+1][c+1] && board[r+1][c+1] == board[r+2][c+2] && board[r+2][c+2] == board[r+3][c+3] {
			return board[r][c]
		}
	}
	// Check diagonal up-right
	for r, c := row+3, col-3; r-3 >= row && c+3 <= col; r, c = r-1, c+1 {
		if board[r][c] != "" && board[r][c] == board[r-1][c+1] && board[r-1][c+1] == board[r-2][c+2] && board[r-2][c+2] == board[r-3][c+3] {
			return board[r][c]
		}
	}
	return ""
}

func getCheckersValidMoves(board [8][8]CheckersPiece, player int) []CheckersMove {
	moves := []CheckersMove{}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if board[r][c].Player == player {
				// Check diagonal moves
				dirs := []int{-1, 1}
				if board[r][c].King {
					dirs = []int{-1, 1}
				} else if player == 1 {
					dirs = []int{1}
				} else {
					dirs = []int{-1}
				}
				for _, dc := range []int{-1, 1} {
					for _, dr := range dirs {
						nr, nc := r+dr, c+dc
						if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
							if board[nr][nc].Player == 0 {
								moves = append(moves, CheckersMove{FromRow: r, FromCol: c, ToRow: nr, ToCol: nc})
							} else if board[nr][nc].Player != player {
								// Jump
								jr, jc := nr+dr, nc+dc
								if jr >= 0 && jr < 8 && jc >= 0 && jc < 8 && board[jr][jc].Player == 0 {
									moves = append(moves, CheckersMove{FromRow: r, FromCol: c, ToRow: jr, ToCol: jc})
								}
							}
						}
					}
				}
			}
		}
	}
	return moves
}

func checkDotsBoxesComplete(board DotsBoxesBoard, row, col int) bool {
	// Check if a box is complete (all 4 sides filled)
	if row < 6 && col < 6 {
		if board.Horizontal[row][col] && board.Horizontal[row+1][col] && board.Vertical[row][col] && board.Vertical[row][col+1] {
			return true
		}
	}
	return false
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
