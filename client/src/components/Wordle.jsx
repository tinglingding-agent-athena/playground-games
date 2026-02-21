import { useState, useEffect } from 'react'
import './Wordle.css'

const WORD_LENGTH = 5
const MAX_GUESSES = 6

export default function Wordle({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ players: [], turn: 0, guesses: [], guessResults: [], winner: '', gameOver: false })
  const [guess, setGuess] = useState('')

  useEffect(() => {
    if (game) {
      setState({
        players: game.players || [],
        turn: game.turn || 0,
        guesses: game.guesses || [],
        guessResults: game.guess_results || [],
        winner: game.winner || '',
        gameOver: game.game_over || false
      })
    }
  }, [game])

  const isMyTurn = state.players[state.turn] === playerId

  const handleGuess = () => {
    if (!isMyTurn || state.gameOver || guess.length !== WORD_LENGTH) return
    onMove(guess.toUpperCase())
    setGuess('')
  }

  const getWinnerMessage = () => {
    if (state.winner === playerId) return 'You Win! 🎉'
    if (state.winner === 'lose') return 'Game Over - You Lose 😔'
    return null
  }

  return (
    <div className="wordle">
      <div className="wordle-header">
        <h2>Wordle</h2>
        <div className="wordle-status">
          {state.gameOver ? (
            <span className="game-over">{getWinnerMessage()}</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? 'Your turn!' : `Waiting for ${state.players[state.turn] || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="wordle-board">
        {Array.from({ length: MAX_GUESSES }).map((_, row) => (
          <div key={row} className="wordle-row">
            {Array.from({ length: WORD_LENGTH }).map((_, col) => {
              const letter = state.guesses[row]?.[col] || ''
              const result = state.guessResults[row]?.[col] || ''
              return (
                <div key={col} className={`wordle-cell ${result}`}>
                  {letter}
                </div>
              )
            })}
          </div>
        ))}
      </div>

      {isMyTurn && !state.gameOver && (
        <div className="wordle-input">
          <input
            type="text"
            value={guess}
            onChange={(e) => setGuess(e.target.value.toUpperCase().slice(0, WORD_LENGTH))}
            placeholder="Enter 5-letter word"
            maxLength={WORD_LENGTH}
          />
          <button onClick={handleGuess} disabled={guess.length !== WORD_LENGTH}>
            Guess
          </button>
        </div>
      )}

      <div className="players-info">
        {state.players.map((p, i) => (
          <div key={i} className={`player-badge ${p === playerId ? 'you' : ''} ${state.turn === i && !state.gameOver ? 'active' : ''}`}>
            Player {i + 1}: {p === playerId ? 'You' : p}
          </div>
        ))}
      </div>
    </div>
  )
}
