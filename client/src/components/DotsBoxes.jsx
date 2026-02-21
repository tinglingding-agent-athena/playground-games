import { useState, useEffect } from 'react'
import './DotsBoxes.css'

const GRID_SIZE = 6

export default function DotsBoxes({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ players: [], turn: 0, board: { horizontal: [], vertical: [] }, scores: [0, 0], gameOver: false })

  useEffect(() => {
    if (game) {
      setState({
        players: game.players || [],
        turn: game.turn || 0,
        board: game.board || { horizontal: [], vertical: [] },
        scores: game.scores || [0, 0],
        gameOver: game.game_over || false
      })
    }
  }, [game])

  const isMyTurn = state.players[state.turn] === playerId
  const myIndex = state.players.indexOf(playerId)

  const handleLineClick = (type, row, col) => {
    if (!isMyTurn || state.gameOver) return
    
    const board = state.board
    if (type === 'horizontal') {
      if (board.horizontal?.[row]?.[col]) return
    } else {
      if (board.vertical?.[row]?.[col]) return
    }
    
    onMove(type, row, col)
  }

  const isLineDrawn = (type, row, col) => {
    const board = state.board
    if (type === 'horizontal') {
      return board.horizontal?.[row]?.[col] === true
    }
    return board.vertical?.[row]?.[col] === true
  }

  const isBoxFilled = (row, col) => {
    const board = state.board
    return board.horizontal?.[row]?.[col] && 
           board.horizontal?.[row+1]?.[col] && 
           board.vertical?.[row]?.[col] && 
           board.vertical?.[row]?.[col+1]
  }

  const getBoxOwner = (row, col) => {
    // Check if box is filled and determine who owns it
    // For simplicity, we just show if it's filled
    return isBoxFilled(row, col)
  }

  return (
    <div className="dotsboxes">
      <div className="dotsboxes-header">
        <h2>Dots and Boxes</h2>
        <div className="dotsboxes-status">
          {state.gameOver ? (
            <span className="game-over">Game Over!</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? 'Your turn!' : `Waiting for ${state.players[state.turn] || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="dotsboxes-scores">
        {state.players.map((p, i) => (
          <div key={i} className={`score-item ${p === playerId ? 'you' : ''} ${state.turn === i && !state.gameOver ? 'active' : ''}`}>
            <span>{p === playerId ? 'You' : p}</span>
            <span className="score">{state.scores[i]}</span>
          </div>
        ))}
      </div>

      <div className="dotsboxes-board">
        {Array.from({ length: GRID_SIZE }).map((_, row) => (
          <div key={row} className="db-row">
            {/* Horizontal lines */}
            <div className="db-line-row">
              {Array.from({ length: GRID_SIZE }).map((_, col) => (
                <div key={col} className="db-cell-top">
                  <button
                    className={`db-line horizontal ${isLineDrawn('horizontal', row, col) ? 'drawn' : ''}`}
                    onClick={() => handleLineClick('horizontal', row, col)}
                    disabled={!isMyTurn || state.gameOver}
                  />
                  {col < GRID_SIZE - 1 && (
                    <div className={`db-box ${getBoxOwner(row, col) ? 'filled' : ''}`} />
                  )}
                </div>
              ))}
            </div>
            {/* Vertical lines */}
            {row < GRID_SIZE - 1 && (
              <div className="db-line-row">
                {Array.from({ length: GRID_SIZE }).map((_, col) => (
                  <div key={col} className="db-cell-side">
                    <button
                      className={`db-line vertical ${isLineDrawn('vertical', row, col) ? 'drawn' : ''}`}
                      onClick={() => handleLineClick('vertical', row, col)}
                      disabled={!isMyTurn || state.gameOver}
                    />
                    {col < GRID_SIZE - 1 && <div className="db-empty" />}
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
