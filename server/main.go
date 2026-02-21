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

// Uno game state
type UnoGame struct {
	Players       []string          `json:"players"`
	Deck          []UnoCard         `json:"deck"`
	Hands         map[string][]UnoCard `json:"hands"`
	CurrentPlayer int                `json:"current_player"`
	CurrentCard   UnoCard            `json:"current_card"`
	Direction     int                `json:"direction"` // 1 = clockwise, -1 = counter-clockwise
	Winner        string             `json:"winner"`
	GameOver      bool               `json:"game_over"`
	GameStartTime time.Time          `json:"game_start_time"`
}

type UnoCard struct {
	Color  string `json:"color"`  // "red", "yellow", "green", "blue", "wild"
	Value  string `json:"value"`  // "0-9", "skip", "reverse", "draw2", "wild", "wild4"
}

type UnoMove struct {
	GameID   string `json:"game_id"`
	Player   string `json:"player"`
	CardIdx  int    `json:"card_idx"`
	ChosenColor string `json:"chosen_color"` // For wild cards
}

// Mafia game state
type MafiaGame struct {
	Players         []string        `json:"players"`
	Roles           map[string]string `json:"roles"` // playerID -> role
	Phase           string          `json:"phase"`   // "night", "day", "lynch", "gameover"
	DayNumber       int             `json:"day_number"`
	AlivePlayers    []string        `json:"alive_players"`
	NightActions    map[string]NightAction `json:"night_actions"`
	VoteCounts      map[string]int `json:"vote_counts"`
	Votes           map[string]string `json:"votes"` // voter -> target
	KillTarget      string          `json:"kill_target"`
	SaveTarget      string          `json:"save_target"`
	Investigation   string          `json:"investigation"` // Result of detective's investigation
	LynchedPlayer   string          `json:"lynched_player"`
	Winner          string          `json:"winner"`
	GameStartTime   time.Time       `json:"game_start_time"`
	GameOver        bool            `json:"game_over"`
}

type NightAction struct {
	Target string `json:"target"`
	Result string `json:"result"`
}

type MafiaAction struct {
	GameID  string `json:"game_id"`
	Player  string `json:"player"`
	Action  string `json:"action"` // "kill", "save", "investigate", "vote"
	Target  string `json:"target"`
}

// Hub maintains active games and connections
type Hub struct {
	tictactoeGames  map[string]*TicTacToeGame
	jeopardyGames   map[string]*JeopardyGame
	hangmanGames    map[string]*HangmanGame
	memoryGames    map[string]*MemoryGame
	battleshipGames map[string]*BattleshipGame
	triviaGames    map[string]*TriviaGame
	rpsGames       map[string]*RPSGame
	connectFourGames map[string]*ConnectFourGame
	checkersGames  map[string]*CheckersGame
	dotsBoxesGames map[string]*DotsBoxesGame
	unoGames       map[string]*UnoGame
	mafiaGames     map[string]*MafiaGame
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
		triviaGames:     make(map[string]*TriviaGame),
		rpsGames:        make(map[string]*RPSGame),
		connectFourGames: make(map[string]*ConnectFourGame),
		checkersGames:   make(map[string]*CheckersGame),
		dotsBoxesGames:  make(map[string]*DotsBoxesGame),
		unoGames:        make(map[string]*UnoGame),
		mafiaGames:      make(map[string]*MafiaGame),
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
				case "trivia":
					game = hub.triviaGames[gameID]
				case "rps":
					game = hub.rpsGames[gameID]
				case "connectfour":
					game = hub.connectFourGames[gameID]
				case "checkers":
					game = hub.checkersGames[gameID]
				case "dotsboxes":
					game = hub.dotsBoxesGames[gameID]
				case "uno":
					game = hub.unoGames[gameID]
				case "mafia":
					game = hub.mafiaGames[gameID]
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
	} else if room.GameType == "uno" {
		game := createUnoGame(room.Players)
		hub.mu.Lock()
		hub.unoGames[gameID] = game
		hub.mu.Unlock()
	} else if room.GameType == "mafia" {
		game := createMafiaGame(room.Players)
		hub.mu.Lock()
		hub.mafiaGames[gameID] = game
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
	case "trivia":
		handleTriviaAnswer(conn, gameID, playerID, payload)
	case "rps":
		handleRPSMove(conn, gameID, playerID, payload)
	case "connectfour":
		handleConnectFourMove(conn, gameID, playerID, payload)
	case "checkers":
		handleCheckersMove(conn, gameID, playerID, payload)
	case "dotsboxes":
		handleDotsBoxesMove(conn, gameID, playerID, payload)
	case "uno":
		handleUnoMove(conn, gameID, playerID, payload)
	case "mafia":
		handleMafiaAction(conn, gameID, playerID, payload)
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

// Uno game functions

func createUnoGame(players []string) *UnoGame {
	// Create deck
	colors := []string{"red", "yellow", "green", "blue"}
	
	deck := []UnoCard{}
	
	// Add number cards (two of each 1-9, one of 0 for each color)
	for _, color := range colors {
		deck = append(deck, UnoCard{Color: color, Value: "0"})
		for _, v := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"} {
			deck = append(deck, UnoCard{Color: color, Value: v})
			deck = append(deck, UnoCard{Color: color, Value: v})
		}
		for _, v := range []string{"skip", "reverse", "draw2"} {
			deck = append(deck, UnoCard{Color: color, Value: v})
			deck = append(deck, UnoCard{Color: color, Value: v})
		}
	}
	
	// Add wild cards
	for i := 0; i < 4; i++ {
		deck = append(deck, UnoCard{Color: "wild", Value: "wild"})
		deck = append(deck, UnoCard{Color: "wild", Value: "wild4"})
	}
	
	// Shuffle deck
	for i := len(deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
	
	// Deal cards to players
	hands := make(map[string][]UnoCard)
	for _, player := range players {
		hands[player] = []UnoCard{}
		for i := 0; i < 7; i++ {
			if len(deck) > 0 {
				card := deck[len(deck)-1]
				deck = deck[:len(deck)-1]
				hands[player] = append(hands[player], card)
			}
		}
	}
	
	// Draw first card
	currentCard := deck[len(deck)-1]
	deck = deck[:len(deck)-1]
	
	// Handle wild first card
	if currentCard.Color == "wild" {
		currentCard.Color = colors[rand.Intn(len(colors))]
	}
	
	game := &UnoGame{
		Players:       players,
		Deck:          deck,
		Hands:         hands,
		CurrentPlayer: 0,
		CurrentCard:   currentCard,
		Direction:     1,
		Winner:        "",
		GameOver:      false,
		GameStartTime: time.Now(),
	}
	
	return game
}

func handleUnoMove(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	cardIdx := int(payload["card_idx"].(float64))
	chosenColor := ""
	if c, ok := payload["chosen_color"].(string); ok {
		chosenColor = c
	}

	hub.mu.RLock()
	game, exists := hub.unoGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
		return
	}

	// Find player index
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

	hand := game.Hands[playerID]
	if cardIdx < 0 || cardIdx >= len(hand) {
		sendMessage(conn, MsgTypeError, "Invalid card index")
		return
	}

	card := hand[cardIdx]
	currentCard := game.CurrentCard

	// Validate move
	valid := false
	if card.Color == "wild" {
		valid = true
	} else if card.Color == currentCard.Color {
		valid = true
	} else if card.Value == currentCard.Value {
		valid = true
	}

	if !valid {
		sendMessage(conn, MsgTypeError, "Invalid move - card doesn't match")
		return
	}

	// Play the card
	game.Hands[playerID] = append(hand[:cardIdx], hand[cardIdx+1:]...)
	game.CurrentCard = card

	// Handle wild card color choice
	if card.Color == "wild" {
		if chosenColor == "" {
			chosenColor = "red" // Default
		}
		game.CurrentCard.Color = chosenColor
	}

	// Check for winner
	if len(game.Hands[playerID]) == 0 {
		game.Winner = playerID
		game.GameOver = true
		broadcastGameState(gameID, "uno", game)
		return
	}

	// Handle special cards
	if card.Value == "reverse" {
		game.Direction *= -1
		if len(game.Players) == 2 {
			// In 2-player, reverse acts like skip
			game.CurrentPlayer = (game.CurrentPlayer + game.Direction + len(game.Players)) % len(game.Players)
		}
	} else if card.Value == "skip" {
		game.CurrentPlayer = (game.CurrentPlayer + game.Direction + len(game.Players)) % len(game.Players)
	} else if card.Value == "draw2" {
		// Next player draws 2 and misses turn
		nextPlayer := (game.CurrentPlayer + game.Direction + len(game.Players)) % len(game.Players)
		nextPlayerID := game.Players[nextPlayer]
		for i := 0; i < 2 && len(game.Deck) > 0; i++ {
			drawCard := game.Deck[len(game.Deck)-1]
			game.Deck = game.Deck[:len(game.Deck)-1]
			game.Hands[nextPlayerID] = append(game.Hands[nextPlayerID], drawCard)
		}
	} else if card.Value == "wild4" {
		// Next player draws 4 and misses turn
		nextPlayer := (game.CurrentPlayer + game.Direction + len(game.Players)) % len(game.Players)
		nextPlayerID := game.Players[nextPlayer]
		for i := 0; i < 4 && len(game.Deck) > 0; i++ {
			drawCard := game.Deck[len(game.Deck)-1]
			game.Deck = game.Deck[:len(game.Deck)-1]
			game.Hands[nextPlayerID] = append(game.Hands[nextPlayerID], drawCard)
		}
	}

	// Move to next player
	game.CurrentPlayer = (game.CurrentPlayer + game.Direction + len(game.Players)) % len(game.Players)

	// Reshuffle if deck is low
	if len(game.Deck) < 5 {
		// For simplicity, just draw more cards as needed
	}

	broadcastGameState(gameID, "uno", game)
}

// Mafia game functions

func createMafiaGame(players []string) *MafiaGame {
	// Assign roles
	roles := make(map[string]string)
	numPlayers := len(players)
	
	// Shuffle players
	shuffled := make([]string, len(players))
	copy(shuffled, players)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	
	// Assign mafia (1 if 3-5 players, 2 if 6-8, 3 if 9+)
	numMafia := 1
	if numPlayers >= 6 {
		numMafia = 2
	}
	if numPlayers >= 9 {
		numMafia = 3
	}
	
	// Assign roles
	mafiaCount := 0
	detectiveIdx := rand.Intn(numPlayers)
	doctorIdx := (detectiveIdx + 1) % numPlayers
	
	for i, player := range shuffled {
		if mafiaCount < numMafia {
			roles[player] = "mafia"
			mafiaCount++
		} else if i == detectiveIdx {
			roles[player] = "detective"
		} else if i == doctorIdx {
			roles[player] = "doctor"
		} else {
			roles[player] = "villager"
		}
	}
	
	// Create list of alive players
	alivePlayers := make([]string, len(players))
	copy(alivePlayers, players)
	
	game := &MafiaGame{
		Players:         players,
		Roles:           roles,
		Phase:           "night",
		DayNumber:       1,
		AlivePlayers:    alivePlayers,
		NightActions:    make(map[string]NightAction),
		VoteCounts:      make(map[string]int),
		Votes:           make(map[string]string),
		KillTarget:      "",
		SaveTarget:      "",
		Investigation:  "",
		LynchedPlayer:   "",
		Winner:          "",
		GameStartTime:   time.Now(),
		GameOver:        false,
	}
	
	return game
}

func handleMafiaAction(conn *websocket.Conn, gameID string, playerID string, payload map[string]interface{}) {
	action := payload["action"].(string)
	target := ""
	if t, ok := payload["target"].(string); ok {
		target = t
	}

	hub.mu.RLock()
	game, exists := hub.mafiaGames[gameID]
	hub.mu.RUnlock()

	if !exists {
		sendMessage(conn, MsgTypeError, "Game not found")
		return
	}

	if game.GameOver {
		sendMessage(conn, MsgTypeError, "Game already over")
		return
	}

	role := game.Roles[playerID]

	// Check if player is alive
	isAlive := false
	for _, p := range game.AlivePlayers {
		if p == playerID {
			isAlive = true
			break
		}
	}
	if !isAlive {
		sendMessage(conn, MsgTypeError, "You are dead")
		return
	}

	switch game.Phase {
	case "night":
		handleMafiaNightAction(conn, gameID, game, playerID, role, action, target)
	case "day":
		handleMafiaDayAction(conn, gameID, game, playerID, role, action, target)
	case "lynch":
		handleMafiaLynchAction(conn, gameID, game, playerID, role, action, target)
	default:
		sendMessage(conn, MsgTypeError, "Invalid game phase")
	}
}

func handleMafiaNightAction(conn *websocket.Conn, gameID string, game *MafiaGame, playerID, role, action, target string) {
	// Validate target is alive
	validTarget := false
	for _, p := range game.AlivePlayers {
		if p == target {
			validTarget = true
			break
		}
	}
	if target != "" && !validTarget {
		sendMessage(conn, MsgTypeError, "Invalid target - player not alive")
		return
	}

	switch role {
	case "mafia":
		if action != "kill" {
			sendMessage(conn, MsgTypeError, "Mafia can only kill at night")
			return
		}
		// Mafia agrees on kill target
		game.NightActions[playerID] = NightAction{Target: target, Result: "kill"}
		
		// Check if all mafia have voted
		mafiaCount := 0
		mafiaVotes := 0
		for _, p := range game.AlivePlayers {
			if game.Roles[p] == "mafia" {
				mafiaCount++
				if _, ok := game.NightActions[p]; ok {
					mafiaVotes++
				}
			}
		}
		if mafiaVotes == mafiaCount {
			// Determine kill target (majority vote)
			voteCounts := make(map[string]int)
			for _, a := range game.NightActions {
				if a.Result == "kill" && a.Target != "" {
					voteCounts[a.Target]++
				}
			}
			maxVotes := 0
			for t, v := range voteCounts {
				if v > maxVotes {
					maxVotes = v
					game.KillTarget = t
				}
			}
		}
		
	case "detective":
		if action != "investigate" {
			sendMessage(conn, MsgTypeError, "Detective can only investigate at night")
			return
		}
		// One investigation per night
		game.NightActions[playerID] = NightAction{Target: target, Result: "investigate"}
		if target != "" {
			targetRole := game.Roles[target]
			if targetRole == "mafia" {
				game.Investigation = target + " is MAFIA!"
			} else {
				game.Investigation = target + " is innocent."
			}
		}
		
	case "doctor":
		if action != "save" {
			sendMessage(conn, MsgTypeError, "Doctor can only save at night")
			return
		}
		game.NightActions[playerID] = NightAction{Target: target, Result: "save"}
		game.SaveTarget = target
		
	case "villager":
		sendMessage(conn, MsgTypeError, "Villagers have no night action")
		return
	}

	// Check if all actions are complete
	actionsComplete := true
	for _, p := range game.AlivePlayers {
		role := game.Roles[p]
		if role == "mafia" || role == "detective" || role == "doctor" {
			if _, ok := game.NightActions[p]; !ok {
				actionsComplete = false
				break
			}
		}
	}

	if actionsComplete {
		// Process night results
		processMafiaNightResults(game)
	}

	broadcastGameState(gameID, "mafia", game)
}

func processMafiaNightResults(game *MafiaGame) {
	// Check if doctor saved the kill target
	if game.SaveTarget == game.KillTarget {
		game.KillTarget = "" // Saved
	}

	// Kill the target if not saved
	if game.KillTarget != "" {
		// Remove from alive players
		newAlive := []string{}
		for _, p := range game.AlivePlayers {
			if p != game.KillTarget {
				newAlive = append(newAlive, p)
			}
		}
		game.AlivePlayers = newAlive
	}

	// Check win conditions
	checkMafiaWinConditions(game)

	if !game.GameOver {
		// Transition to day
		game.Phase = "day"
		game.DayNumber++
		game.NightActions = make(map[string]NightAction)
		game.KillTarget = ""
		game.SaveTarget = ""
		game.Votes = make(map[string]string)
		game.VoteCounts = make(map[string]int)
	}
}

func handleMafiaDayAction(conn *websocket.Conn, gameID string, game *MafiaGame, playerID, role, action, target string) {
	// In day phase, players can discuss but can't vote yet
	// They can only suggest suspects
	sendMessage(conn, MsgTypeGameState, map[string]interface{}{
		"game_id": gameID,
		"game":    game,
		"message": "Discussion phase - use /lynch to start voting",
	})
}

func handleMafiaLynchAction(conn *websocket.Conn, gameID string, game *MafiaGame, playerID, role, action, target string) {
	if action != "vote" {
		sendMessage(conn, MsgTypeError, "Must vote to lynch during voting phase")
		return
	}

	// Validate target is alive
	validTarget := false
	for _, p := range game.AlivePlayers {
		if p == target {
			validTarget = true
			break
		}
	}
	if !validTarget {
		sendMessage(conn, MsgTypeError, "Invalid target - player not alive")
		return
	}

	// Record vote
	game.Votes[playerID] = target
	game.VoteCounts[target]++

	// Check if all alive players have voted
	votesNeeded := len(game.AlivePlayers)
	if len(game.Votes) >= votesNeeded {
		// Find player with most votes
		maxVotes := 0
		lynchTarget := ""
		for t, v := range game.VoteCounts {
			if v > maxVotes {
				maxVotes = v
				lynchTarget = t
			}
		}

		// Check for tie
		tieCount := 0
		for _, v := range game.VoteCounts {
			if v == maxVotes {
				tieCount++
			}
		}

		if tieCount > 1 {
			// Tie - no lynching
			game.LynchedPlayer = ""
			game.Phase = "night"
			game.Votes = make(map[string]string)
			game.VoteCounts = make(map[string]int)
		} else {
			// Lynched!
			game.LynchedPlayer = lynchTarget
			newAlive := []string{}
			for _, p := range game.AlivePlayers {
				if p != lynchTarget {
					newAlive = append(newAlive, p)
				}
			}
			game.AlivePlayers = newAlive

			// Check win conditions
			checkMafiaWinConditions(game)

			if !game.GameOver {
				game.Phase = "night"
				game.Votes = make(map[string]string)
				game.VoteCounts = make(map[string]int)
			}
		}
	}

	broadcastGameState(gameID, "mafia", game)
}

func checkMafiaWinConditions(game *MafiaGame) {
	// Count alive mafia and villagers
	mafiaAlive := 0
	villagersAlive := 0
	
	for _, p := range game.AlivePlayers {
		if game.Roles[p] == "mafia" {
			mafiaAlive++
		} else {
			villagersAlive++
		}
	}

	if mafiaAlive == 0 {
		// Villagers win
		game.Winner = "villagers"
		game.GameOver = true
	} else if mafiaAlive >= villagersAlive {
		// Mafia wins
		game.Winner = "mafia"
		game.GameOver = true
	}
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
