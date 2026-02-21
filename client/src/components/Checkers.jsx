import { useState, useEffect } from 'react'
import './Checkers.css'

const BOARD_SIZE = 8

export default function Checkers({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ board: [], players: [], turn: 0, winner: '' })
  const [selected, setSelected] = useState(null)

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

  const handleCellClick = (row, col) => {
    if (!isMyTurn || state.winner) return
    
    const piece = state.board[row]?.[col]
    
    if (selected) {
      // Try to move
      if (isValidMove(selected.row, selected.col, row, col)) {
        onMove(selected.row, selected.col, row, col)
        setSelected(null)
      } else if (piece?.player === myIndex + 1) {
        // Select different piece
        setSelected({ row, col })
      } else {
        setSelected(null)
      }
    } else {
      // Select piece
      if (piece?.player === myIndex + 1) {
        setSelected({ row, col })
      }
    }
  }

  const isValidMove = (fromRow, fromCol, toRow, toCol) => {
    // Simple validation - actual validation is on server
    const piece = state.board[fromRow]?.[fromCol]
    if (!piece) return false
    
    const dr = toRow - fromRow
    const dc = toCol - fromCol
    
    // Regular move
    if (Math.abs(dr) === 1 && Math.abs(dc) === 1) {
      return true
    }
    // Jump
    if (Math.abs(dr) === 2 && Math.abs(dc) === 2) {
      return true
    }
    
    return false
  }

  const getWinnerMessage = () => {
    if (state.winner === playerId) return 'You Win! 🎉'
    if (state.winner) return 'You Lose 😔'
    return null
  }

  const renderPiece = (piece) => {
    if (!piece || piece.player === 0) return null
    const isMine = piece.player === myIndex + 1
    return (
      <span className={`checker-piece ${isMine ? 'mine' : 'opponent'} ${piece.king ? 'king' : ''}`}>
        {piece.king ? '♔' : '●'}
      </span>
    )
  }

  return (
    <div className="checkers">
      <div className="checkers-header">
        <h2>Checkers</h2>
        <div className="checkers-status">
          {state.winner ? (
            <span className="game-over">{getWinnerMessage()}</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? 'Your turn!' : `Waiting for ${state.players[state.turn] || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="checkers-board">
        {state.board.map((row, r) => (
          <div key={r} className="checkers-row">
            {row.map((piece, c) => (
              <button
                key={c}
                className={`checkers-cell ${(r + c) % 2 === 1 ? 'dark' : 'light'} ${selected?.row === r && selected?.col === c ? 'selected' : ''}`}
                onClick={() => handleCellClick(r, c)}
                disabled={!isMyTurn || state.winner}
              >
                {renderPiece(piece)}
              </button>
            ))}
          </div>
        ))}
      </div>

      <div className="players-info">
        {state.players.map((p, i) => (
          <div key={i} className={`player-badge ${p === playerId ? 'you' : ''} ${state.turn === i && !state.winner ? 'active' : ''}`}>
            Player {i + 1}: {p === playerId ? 'You' : p}
          </div>
        ))}
      </div>
    </div>
  )
}
