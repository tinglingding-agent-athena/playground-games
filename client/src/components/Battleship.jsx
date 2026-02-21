import { useState, useEffect } from 'react'
import './Battleship.css'

export default function Battleship({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ players: [], turn: 0, grids: [], winner: '', gamePhase: 'placing' })

  useEffect(() => {
    if (game) {
      setState({
        players: game.players || [],
        turn: game.turn || 0,
        grids: game.grids || [],
        winner: game.winner || '',
        gamePhase: game.game_phase || 'placing'
      })
    }
  }, [game])

  const myIndex = state.players.indexOf(playerId)
  const isMyTurn = state.turn === myIndex && myIndex !== -1

  const handleCellClick = (x, y) => {
    if (!isMyTurn || state.gamePhase !== 'playing' || state.winner) return
    
    // Check if already fired here
    const opponentGrid = state.grids[1 - myIndex]
    if (!opponentGrid) return
    
    const cell = opponentGrid.cells?.[y]?.[x]
    if (cell?.hit || cell?.miss) return
    
    onMove(x, y)
  }

  const getMyGrid = () => {
    if (myIndex === -1 || !state.grids[myIndex]) return null
    return state.grids[myIndex]
  }

  const getOpponentGrid = () => {
    if (myIndex === -1 || !state.grids[1 - myIndex]) return null
    return state.grids[1 - myIndex]
  }

  const getWinnerMessage = () => {
    if (state.winner === playerId) return 'You Win! 🎉'
    if (state.winner && state.winner !== playerId) return 'You Lose 😔'
    return null
  }

  const renderCell = (cell, isMyGrid, x, y) => {
    if (!cell) return 'empty'
    if (isMyGrid) {
      if (cell.has_ship && cell.hit) return 'ship-hit'
      if (cell.has_ship) return 'ship'
      if (cell.miss) return 'miss'
      return 'empty'
    } else {
      if (cell.hit) return 'hit'
      if (cell.miss) return 'miss'
      return 'unknown'
    }
  }

  return (
    <div className="battleship">
      <div className="battleship-header">
        <h2>Battleship</h2>
        <div className="battleship-status">
          {state.winner ? (
            <span className="game-over">{getWinnerMessage()}</span>
          ) : state.gamePhase === 'placing' ? (
            <span className="waiting">Placing ships...</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? 'Your turn - Fire!' : `Waiting for ${state.players[state.turn] || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="battleship-boards">
        <div className="board-section">
          <h3>Your Fleet</h3>
          <div className="battleship-grid">
            {getMyGrid()?.cells?.map((row, y) => (
              <div key={y} className="grid-row">
                {row.map((cell, x) => (
                  <div key={x} className={`grid-cell ${renderCell(cell, true, x, y)}`}>
                    {cell?.has_ship && '🚢'}
                    {cell?.hit && '💥'}
                    {cell?.miss && '💧'}
                  </div>
                ))}
              </div>
            ))}
          </div>
        </div>

        <div className="board-section">
          <h3>Enemy Waters</h3>
          <div className="battleship-grid">
            {getOpponentGrid()?.cells?.map((row, y) => (
              <div key={y} className="grid-row">
                {row.map((cell, x) => (
                  <button
                    key={x}
                    className={`grid-cell clickable ${renderCell(cell, false, x, y)}`}
                    onClick={() => handleCellClick(x, y)}
                    disabled={!isMyTurn || state.gamePhase !== 'playing' || state.winner}
                  >
                    {cell?.hit && '💥'}
                    {cell?.miss && '💧'}
                  </button>
                ))}
              </div>
            ))}
          </div>
        </div>
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
