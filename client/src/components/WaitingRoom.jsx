import { useState } from 'react'
import './WaitingRoom.css'

export default function WaitingRoom({ room, playerId, onStartGame, onLeaveRoom }) {
  const [error, setError] = useState('')
  const isHost = room.host === playerId

  const handleStart = () => {
    setError('')
    if (room.players.length < 1) {
      setError('Need at least 1 player to start')
      return
    }
    onStartGame()
  }

  const gameTypeNames = {
    tictactoe: 'Tic Tac Toe',
    jeopardy: 'Jeopardy'
  }

  const gameModeNames = {
    tictactoe: {
      classic: 'Classic',
      fading: 'Fading',
      speed: 'Speed',
      infinite: 'Infinite'
    },
    jeopardy: {
      classic: 'Classic',
      speed: 'Speed Round',
      teams: 'Teams'
    }
  }

  const getModeName = () => {
    const modes = gameModeNames[room.game_type]
    if (modes && room.game_mode) {
      return modes[room.game_mode] || room.game_mode
    }
    return 'Classic'
  }

  return (
    <div className="waiting-room">
      <div className="waiting-container">
        <div className="room-info">
          <div className="room-code-display">
            <span className="label">Room Code</span>
            <span className="code">{room.code}</span>
          </div>
          <div className="game-type">
            {gameTypeNames[room.game_type] || room.game_type}
            <span className="mode-tag">{getModeName()}</span>
          </div>
        </div>

        <div className="players-section">
          <h3>Players ({room.players.length})</h3>
          <div className="players-list">
            {room.players.map((player, index) => (
              <div key={index} className={`player-item ${player === playerId ? 'you' : ''} ${player === room.host ? 'host' : ''}`}>
                <span className="player-avatar">
                  {player === playerId ? '👤' : '👥'}
                </span>
                <span className="player-name">
                  {room.player_names?.[player] || player}
                  {player === playerId && ' (You)'}
                  {player === room.host && ' ⭐ Host'}
                </span>
              </div>
            ))}
            {room.players.length < 2 && (
              <div className="waiting-message">
                Waiting for more players to join...
              </div>
            )}
          </div>
        </div>

        {error && <div className="error-message">{error}</div>}

        <div className="waiting-actions">
          {isHost ? (
            <button className="start-btn" onClick={handleStart}>
              Start Game
            </button>
          ) : (
            <div className="waiting-status">
              Waiting for host to start the game...
            </div>
          )}
          <button className="leave-btn" onClick={onLeaveRoom}>
            Leave Room
          </button>
        </div>

        <div className="share-info">
          <p>Share this code with friends to join:</p>
          <div className="share-code">{room.code}</div>
        </div>
      </div>
    </div>
  )
}
