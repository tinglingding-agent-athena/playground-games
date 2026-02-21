import { useState, useEffect } from 'react'
import './Game2048.css'

export default function Game2048({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ board: [], score: 0, bestScore: 0, gameOver: false, won: false })

  useEffect(() => {
    if (game) {
      setState({
        board: game.board || [],
        score: game.score || 0,
        bestScore: game.best_score || 0,
        gameOver: game.game_over || false,
        won: game.won || false
      })
    }
  }, [game])

  const handleKeyDown = (e) => {
    if (state.gameOver) return
    
    const keyMap = {
      ArrowUp: 'up',
      ArrowDown: 'down',
      ArrowLeft: 'left',
      ArrowRight: 'right'
    }
    
    if (keyMap[e.key]) {
      e.preventDefault()
      onMove(keyMap[e.key])
    }
  }

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [state.gameOver])

  const getTileColor = (value) => {
    const colors = {
      2: '#eee4da',
      4: '#ede0c8',
      8: '#f2b179',
      16: '#f59563',
      32: '#f67c5f',
      64: '#f65e3b',
      128: '#edcf72',
      256: '#edcc61',
      512: '#edc850',
      1024: '#edc53f',
      2048: '#edc22e'
    }
    return colors[value] || '#3c3a32'
  }

  const getTileTextColor = (value) => {
    return value <= 4 ? '#776e65' : '#f9f6f2'
  }

  return (
    <div className="game2048">
      <div className="game2048-header">
        <h2>2048</h2>
        <div className="game2048-scores">
          <div className="score-box">
            <span className="score-label">Score</span>
            <span className="score-value">{state.score}</span>
          </div>
          <div className="score-box best">
            <span className="score-label">Best</span>
            <span className="score-value">{state.bestScore}</span>
          </div>
        </div>
      </div>

      {state.gameOver && (
        <div className="game-over-overlay">
          <div className="game-over-message">
            {state.won ? 'You reached 2048! 🎉' : 'Game Over!'}
          </div>
        </div>
      )}

      <div className="game2048-board">
        {state.board.map((row, r) => (
          <div key={r} className="board-row">
            {row.map((cell, c) => (
              <div
                key={c}
                className={`tile ${cell !== 0 ? 'filled' : ''}`}
                style={{
                  backgroundColor: cell ? getTileColor(cell) : '#cdc1b4',
                  color: cell ? getTileTextColor(cell) : 'transparent'
                }}
              >
                {cell || ''}
              </div>
            ))}
          </div>
        ))}
      </div>

      <div className="game2048-controls">
        <p>Use arrow keys to move tiles</p>
        <div className="arrow-keys">
          <button onClick={() => onMove('up')}>↑</button>
          <div className="arrow-row">
            <button onClick={() => onMove('left')}>←</button>
            <button onClick={() => onMove('down')}>↓</button>
            <button onClick={() => onMove('right')}>→</button>
          </div>
        </div>
      </div>
    </div>
  )
}
