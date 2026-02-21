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
    { id: 'speed', name: 'Speed Round', description: '10 second timer per question' },
    { id: 'teams', name: 'Teams', description: '2v2 team mode' }
  ]
}

export default function Lobby({ onCreateRoom, onJoinRoom, ws }) {
  const [joinCode, setJoinCode] = useState('')
  const [error, setError] = useState('')
  const [selectedGame, setSelectedGame] = useState('tictactoe')
  const [selectedMode, setSelectedMode] = useState('classic')
  const [showModes, setShowModes] = useState(false)

  const handleGameSelect = (game) => {
    setSelectedGame(game)
    setSelectedMode('classic')
    setShowModes(true)
  }

  const handleCreate = () => {
    setError('')
    onCreateRoom(selectedGame, selectedMode)
  }

  const handleJoin = () => {
    setError('')
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

        <div className="game-selection">
          <h3>Select Game</h3>
          <div className="game-options">
            <button
              className={`game-option ${selectedGame === 'tictactoe' ? 'selected' : ''}`}
              onClick={() => handleGameSelect('tictactoe')}
            >
              <span className="game-icon">⭕</span>
              <span className="game-name">Tic Tac Toe</span>
            </button>
            <button
              className={`game-option ${selectedGame === 'jeopardy' ? 'selected' : ''}`}
              onClick={() => handleGameSelect('jeopardy')}
            >
              <span className="game-icon">🎯</span>
              <span className="game-name">Jeopardy</span>
            </button>
          </div>
        </div>

        {showModes && (
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
