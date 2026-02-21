import { useState, useEffect } from 'react'
import './ConnectFour.css'

const ROWS = 6
const COLS = 7

export default function ConnectFour({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ board: [], players: [], turn: 0, winner: '' })

  useEffect(() => {
    if (game) {
      setState({
        board: game.board || [],
        players: game.players || [],
        turn: game.turn || 0,
        winner: game.winner || ''
      })
    }
  }, [game])

  const isMyTurn = state.players[state.turn] === playerId
  const myIndex = state.players.indexOf(playerId)

  const handleColumnClick = (col) => {
    if (!isMyTurn || state.winner) return
    
    // Check if column is full
    if (state.board[0]?.[col]) return
    
    onMove(col)
  }

  const getWinnerMessage = () => {
    if (state.winner === playerId) return 'You Win! 🎉'
    if (state.winner === 'draw') return "It's a Draw!"
    if (state.winner) return 'You Lose 😔'
    return null
  }

  const getCellContent = (row, col) => {
    const cell = state.board[row]?.[col]
    if (!cell) return null
    return cell
  }

  return (
    <div className="connectfour">
      <div className="connectfour-header">
        <h2>Connect Four</h2>
        <div className="connectfour-status">
          {state.winner ? (
            <span className="game-over">{getWinnerMessage()}</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? 'Your turn!' : `Waiting for ${state.players[state.turn] || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="connectfour-board">
        {Array.from({ length: ROWS }).map((_, row) => (
          <div key={row} className="cf-row">
            {Array.from({ length: COLS }).map((_, col) => {
              const content = getCellContent(row, col)
              return (
                <div key={col} className="cf-cell">
                  {content && <span className={`cf-disc ${content === '🔴' ? 'red' : 'yellow'}`}>{content}</span>}
                </div>
              )
            })}
          </div>
        ))}
        <div className="cf-columns">
          {Array.from({ length: COLS }).map((_, col) => (
            <button
              key={col}
              className="cf-col-btn"
              onClick={() => handleColumnClick(col)}
              disabled={!isMyTurn || state.winner}
            >
              ▼
            </button>
          ))}
        </div>
      </div>

      <div className="players-info">
        {state.players.map((p, i) => (
          <div key={i} className={`player-badge ${p === playerId ? 'you' : ''} ${state.turn === i && !state.winner ? 'active' : ''}`}>
            {i === 0 ? '🔴' : '🟡'} {p === playerId ? 'You' : p}
          </div>
        ))}
      </div>
    </div>
  )
}
