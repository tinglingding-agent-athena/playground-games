import { useState, useEffect } from 'react'
import './Trivia.css'

export default function Trivia({ game, gameId, playerId, onAnswer, ws }) {
  const [state, setState] = useState({ players: [], scores: {}, currentQ: 0, questions: [], gameOver: false })

  useEffect(() => {
    if (game) {
      setState({
        players: game.players || [],
        scores: game.scores || {},
        currentQ: game.current_q || 0,
        questions: game.questions || [],
        gameOver: game.game_over || false
      })
    }
  }, [game])

  const currentQuestion = state.questions[state.currentQ]

  const getWinnerMessage = () => {
    const sortedPlayers = [...state.players].sort((a, b) => (state.scores[b] || 0) - (state.scores[a] || 0))
    if (sortedPlayers[0] === playerId) return 'You Win! 🎉'
    return `Winner: ${sortedPlayers[0]}`
  }

  return (
    <div className="trivia">
      <div className="trivia-header">
        <h2>Trivia Quiz</h2>
        <div className="trivia-progress">
          Question {Math.min(state.currentQ + 1, state.questions.length)} of {state.questions.length}
        </div>
      </div>

      <div className="trivia-content">
        {state.gameOver ? (
          <div className="game-over">
            <h3>Game Over!</h3>
            <p>{getWinnerMessage()}</p>
            <div className="final-scores">
              {[...state.players].sort((a, b) => (state.scores[b] || 0) - (state.scores[a] || 0)).map((p, i) => (
                <div key={p} className={`score-item ${p === playerId ? 'you' : ''}`}>
                  <span>#{i + 1}</span>
                  <span>{p === playerId ? 'You' : p}</span>
                  <span>{state.scores[p] || 0}</span>
                </div>
              ))}
            </div>
          </div>
        ) : currentQuestion ? (
          <div className="question-section">
            <div className="question-category">{currentQuestion.category}</div>
            <div className="question-text">{currentQuestion.question}</div>
            <div className="options">
              {currentQuestion.options?.map((option, idx) => (
                <button
                  key={idx}
                  className="option-btn"
                  onClick={() => onAnswer(idx)}
                >
                  {option}
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div className="waiting">Loading question...</div>
        )}

        <div className="scoreboard">
          <h3>Scores</h3>
          {state.players.map(p => (
            <div key={p} className={`score-item ${p === playerId ? 'you' : ''}`}>
              <span>{p === playerId ? 'You' : p}</span>
              <span>{state.scores[p] || 0}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
