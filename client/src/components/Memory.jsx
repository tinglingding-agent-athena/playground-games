import { useState, useEffect } from 'react'
import './Memory.css'

export default function Memory({ game, gameId, playerId, onMove, ws }) {
  const [state, setState] = useState({ cards: [], flippedCards: [], scores: {}, players: [], currentPlayer: 0, canFlip: true, gameOver: false })
  const [flipped, setFlipped] = useState([])

  useEffect(() => {
    if (game) {
      setState({
        cards: game.cards || [],
        flippedCards: game.flipped_cards || [],
        scores: game.scores || {},
        players: game.players || [],
        currentPlayer: game.current_player || 0,
        canFlip: game.can_flip !== false,
        gameOver: game.game_over || false
      })
    }
  }, [game])

  const isMyTurn = state.players[state.currentPlayer] === playerId

  const handleCardClick = (index) => {
    if (!isMyTurn || !state.canFlip || state.cards[index].flipped || state.cards[index].matched || state.gameOver) return
    onMove(index)
  }

  const getMyScore = () => state.scores[playerId] || 0

  return (
    <div className="memory">
      <div className="memory-header">
        <h2>Memory</h2>
        <div className="memory-status">
          {state.gameOver ? (
            <span className="game-over">Game Over!</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? 'Your turn!' : `Waiting for ${state.players[state.currentPlayer] || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="memory-board">
        {state.cards.map((card, index) => (
          <button
            key={index}
            className={`memory-card ${card.flipped || card.matched ? 'flipped' : ''} ${card.matched ? 'matched' : ''}`}
            onClick={() => handleCardClick(index)}
            disabled={!isMyTurn || !state.canFlip || card.flipped || card.matched || state.gameOver}
          >
            <div className="card-inner">
              <div className="card-front">?</div>
              <div className="card-back">{card.value}</div>
            </div>
          </button>
        ))}
      </div>

      <div className="memory-scores">
        {state.players.map((p, i) => (
          <div key={i} className={`score-item ${p === playerId ? 'you' : ''} ${state.currentPlayer === i && !state.gameOver ? 'active' : ''}`}>
            <span>{p === playerId ? 'You' : p}</span>
            <span className="score">{state.scores[p] || 0}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
