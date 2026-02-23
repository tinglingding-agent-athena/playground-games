import { useState } from 'react'
import './Lobby.css'

// Game modes configuration
const GAME_MODES = {
  tictactoe: [
    { id: 'classic', name: 'Classic', description: 'Normal rules - first to get 3 in a row wins' },
    { id: 'fading', name: 'Fading', description: 'Oldest move disappears after 4 moves - adds strategy' },
    { id: 'speed', name: 'Speed', description: '5 second timer per move - think fast!' },
    { id: 'infinite', name: 'Infinite', description: "Board doesn't reset, keep playing until someone wins" }
  ],
  jeopardy: [
    { id: 'classic', name: 'Classic', description: 'Normal Jeopardy rules' },
    { id: 'speed', name: 'Speed Round', description: '10 second timer per question' }
  ],
  hangman: [
    { id: 'classic', name: 'Classic', description: 'Guess letters, 6 wrong guesses allowed' }
  ],
  memory: [
    { id: 'classic', name: 'Classic', description: 'Find all matching pairs' }
  ],
  battleship: [
    { id: 'classic', name: 'Classic', description: 'Naval combat - sink all ships to win' }
  ],
  trivia: [
    { id: 'classic', name: 'Classic', description: 'Answer multiple choice questions' }
  ],
  rps: [
    { id: 'classic', name: 'Best of 3', description: 'Rock Paper Scissors - first to 2 wins' }
  ],
  connectfour: [
    { id: 'classic', name: 'Classic', description: 'Connect 4 discs in a row to win' }
  ],
  checkers: [
    { id: 'classic', name: 'Classic', description: 'Capture all opponent pieces to win' }
  ],
  dotsboxes: [
    { id: 'classic', name: 'Classic', description: 'Draw lines, complete boxes to score' }
  ],
  uno: [
    { id: 'classic', name: 'Classic', description: 'Match color or number, first to empty hand wins' }
  ],
  mafia: [
    { id: 'classic', name: 'Classic', description: 'Social deduction - Mafia vs Villagers' }
  ]
}

const GAMES = [
  { id: 'tictactoe', name: 'Tic Tac Toe', icon: '⭕' },
  { id: 'jeopardy', name: 'Jeopardy', icon: '🎯' },
  { id: 'hangman', name: 'Hangman', icon: '🔤' },
  { id: 'memory', name: 'Memory', icon: '🃏' },
  { id: 'battleship', name: 'Battleship', icon: '🚢' },
  { id: 'trivia', name: 'Trivia Quiz', icon: '❓' },
  { id: 'rps', name: 'Rock Paper Scissors', icon: '✊' },
  { id: 'connectfour', name: 'Connect Four', icon: '🔴' },
  { id: 'checkers', name: 'Checkers', icon: '♟️' },
  { id: 'dotsboxes', name: 'Dots and Boxes', icon: '⬜' },
  { id: 'uno', name: 'Uno', icon: '🎴' },
  { id: 'mafia', name: 'Mafia', icon: '🕵️' }
]

export default function Lobby({ playerName, setPlayerName, onCreateRoom, onJoinRoom, ws }) {
  const [joinCode, setJoinCode] = useState('')
  const [error, setError] = useState('')
  const [selectedGame, setSelectedGame] = useState('tictactoe')
  const [selectedMode, setSelectedMode] = useState('classic')
  const [showModes, setShowModes] = useState(false)

  const handlePlayerNameChange = (e) => {
    const name = e.target.value.slice(0, 20) // Limit to 20 characters
    setPlayerName(name)
    localStorage.setItem('playerName', name)
  }

  const handleGameSelect = (game) => {
    setSelectedGame(game)
    setSelectedMode('classic')
    setShowModes(true)
  }

  const handleCreate = () => {
    setError('')
    if (!playerName.trim()) {
      setError('Please enter your name')
      return
    }
    onCreateRoom(selectedGame, selectedMode)
  }

  const handleJoin = () => {
    setError('')
    if (!playerName.trim()) {
      setError('Please enter your name')
      return
    }
    const code = joinCode.trim().toUpperCase()
    if (!code) {
      setError('Please enter a room code')
      return
    }
    if (code.length < 4) {
      setError('Room code must be 4 characters')
      return
    }
    onJoinRoom(code)
  }

  return (
    <div className="lobby">
      <div className="lobby-container">
        <div className="lobby-header">
          <h1>🎮 Playground Games</h1>
          <p className="subtitle">Choose a game and create or join a room</p>
        </div>

        <div className="player-name-section">
          <label htmlFor="playerName">Your Name:</label>
          <input
            id="playerName"
            type="text"
            value={playerName}
            onChange={handlePlayerNameChange}
            placeholder="Enter your name"
            maxLength={20}
          />
        </div>

        <div className="game-selection">
          <h3>Select Game</h3>
          <div className="game-options">
            {GAMES.map((game) => (
              <button
                key={game.id}
                className={`game-option ${selectedGame === game.id ? 'selected' : ''}`}
                onClick={() => handleGameSelect(game.id)}
              >
                <span className="game-icon">{game.icon}</span>
                <span className="game-name">{game.name}</span>
              </button>
            ))}
          </div>
        </div>

        {showModes && GAME_MODES[selectedGame] && (
          <div className="mode-selection">
            <h3>Select Mode</h3>
            <div className="mode-options">
              {GAME_MODES[selectedGame].map((mode) => (
                <button
                  key={mode.id}
                  className={`mode-option ${selectedMode === mode.id ? 'selected' : ''}`}
                  onClick={() => setSelectedMode(mode.id)}
                >
                  <span className="mode-name">{mode.name}</span>
                  <span className="mode-description">{mode.description}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        <div className="lobby-actions">
          <div className="action-card create-room">
            <h3>Create Room</h3>
            <p>Start a new game room and invite friends</p>
            <button className="primary-btn" onClick={handleCreate}>
              Create Room
            </button>
          </div>

          <div className="action-card join-room">
            <h3>Join Room</h3>
            <p>Enter a room code to join an existing game</p>
            <div className="join-input-group">
              <input
                type="text"
                value={joinCode}
                onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
                placeholder="Room Code"
                maxLength={6}
              />
              <button className="secondary-btn" onClick={handleJoin}>
                Join
              </button>
            </div>
          </div>
        </div>

        {error && <div className="error-message">{error}</div>}
      </div>
    </div>
  )
}
