import { useState, useEffect } from 'react'
import './RPS.css'

const MOVES = ['rock', 'paper', 'scissors']
const MOVE_ICONS = { rock: '✊', paper: '✋', scissors: '✌️' }

export default function RPS({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ players: [], turn: 0, moves: [], scores: [0, 0], winner: '', roundOver: false, gameOver: false, bestOf: 3 })

  useEffect(() => {
    if (game) {
      setState({
        players: game.players || [],
        turn: game.turn || 0,
        moves: game.moves || [],
        scores: game.scores || [0, 0],
        winner: game.winner || '',
        roundOver: game.round_over || false,
        gameOver: game.game_over || false,
        bestOf: game.best_of || 3
      })
    }
  }, [game])

  const myIndex = state.players.indexOf(playerId)
  const myMove = myIndex !== -1 ? state.moves[myIndex] : ''
  const canPlay = !state.gameOver && state.moves[myIndex] === ''

  const getWinnerMessage = () => {
    if (state.winner === playerId) return 'You Win! 🎉'
    if (state.winner === 'draw') return "It's a Draw!"
    if (state.winner) return 'You Lose 😔'
    return null
  }

  const getRoundResult = () => {
    if (!state.roundOver || state.moves[0] === '' || state.moves[1] === '') return null
    
    const m0 = state.moves[0]
    const m1 = state.moves[1]
    
    if (m0 === m1) return "It's a Draw!"
    if ((m0 === 'rock' && m1 === 'scissors') ||
        (m0 === 'paper' && m1 === 'rock') ||
        (m0 === 'scissors' && m1 === 'paper')) {
      return state.players[0] === playerId ? 'You won this round!' : 'You lost this round'
    }
    return state.players[1] === playerId ? 'You won this round!' : 'You lost this round'
  }

  return (
    <div className="rps">
      <div className="rps-header">
        <h2>Rock Paper Scissors</h2>
        <div className="rps-status">
          {state.gameOver ? (
            <span className="game-over">{getWinnerMessage()}</span>
          ) : state.roundOver ? (
            <span className="round-result">{getRoundResult()}</span>
          ) : canPlay ? (
            <span className="your-turn">Play your move!</span>
          ) : (
            <span className="waiting">Waiting for opponent...</span>
          )}
        </div>
      </div>

      <div className="rps-game">
        <div className="rps-scores">
          <div className={`score-section ${state.players[0] === playerId ? 'you' : ''}`}>
            <div className="player-name">{state.players[0] === playerId ? 'You' : state.players[0] || 'Player 1'}</div>
            <div className="score">{state.scores[0]}</div>
            <div className="player-move">
              {state.moves[0] ? MOVE_ICONS[state.moves[0]] : '?'}
            </div>
          </div>
          <div className="vs">VS</div>
          <div className={`score-section ${state.players[1] === playerId ? 'you' : ''}`}>
            <div className="player-name">{state.players[1] === playerId ? 'You' : state.players[1] || 'Player 2'}</div>
            <div className="score">{state.scores[1]}</div>
            <div className="player-move">
              {state.moves[1] ? MOVE_ICONS[state.moves[1]] : '?'}
            </div>
          </div>
        </div>

        {canPlay && (
          <div className="rps-moves">
            {MOVES.map(move => (
              <button
                key={move}
                className="move-btn"
                onClick={() => onMove(move)}
              >
                <span className="move-icon">{MOVE_ICONS[move]}</span>
                <span className="move-name">{move}</span>
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="rps-info">
        Best of {state.bestOf} - First to {Math.ceil(state.bestOf / 2)} wins
      </div>
    </div>
  )
}
