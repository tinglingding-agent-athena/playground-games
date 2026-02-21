import { useState } from 'react'
import './Lobby.css'

export default function Lobby({ onCreateRoom, onJoinRoom, ws }) {
  const [joinCode, setJoinCode] = useState('')
  const [error, setError] = useState('')
  const [selectedGame, setSelectedGame] = useState('tictactoe')

  const handleCreate = () => {
    setError('')
    onCreateRoom(selectedGame)
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
              onClick={() => setSelectedGame('tictactoe')}
            >
              <span className="game-icon">⭕</span>
              <span className="game-name">Tic Tac Toe</span>
            </button>
            <button
              className={`game-option ${selectedGame === 'jeopardy' ? 'selected' : ''}`}
              onClick={() => setSelectedGame('jeopardy')}
            >
              <span className="game-icon">🎯</span>
              <span className="game-name">Jeopardy</span>
            </button>
          </div>
        </div>

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
