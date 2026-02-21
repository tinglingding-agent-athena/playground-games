import { useState, useEffect } from 'react'
import './Jeopardy.css'

const CATEGORIES = ['Science', 'History', 'Geography']
const VALUES = [100, 200, 300]

export default function Jeopardy({ game, gameId, playerId, onAnswer, ws }) {
  const [players, setPlayers] = useState([])
  const [scores, setScores] = useState({})
  const [currentQ, setCurrentQ] = useState(0)
  const [questions, setQuestions] = useState([])
  const [selectedQ, setSelectedQ] = useState(null)
  const [answer, setAnswer] = useState('')
  const [result, setResult] = useState(null)
  const [answered, setAnswered] = useState(new Set())

  useEffect(() => {
    if (game) {
      setPlayers(game.players || [])
      setScores(game.scores || {})
      setCurrentQ(game.current_q || 0)
      setQuestions(game.questions || [])
    }
  }, [game])

  const currentQuestion = questions[currentQ]
  const isGameOver = currentQ >= questions.length
  const myScore = scores[playerId] || 0

  const handleQuestionClick = (index) => {
    if (answered.has(index)) return
    setSelectedQ(index)
    setResult(null)
    setAnswer('')
  }

  const handleSubmitAnswer = (e) => {
    e.preventDefault()
    if (!answer.trim() || selectedQ === null) return
    
    onAnswer(selectedQ, answer)
    setAnswer('')
  }

  // Group questions by category
  const getQuestionForCell = (catIndex, valIndex) => {
    const category = CATEGORIES[catIndex]
    const value = VALUES[valIndex]
    const q = questions.find(q => q.category === category && q.value === value)
    if (!q) return null
    const qIndex = questions.indexOf(q)
    return { ...q, index: qIndex, answered: answered.has(qIndex) }
  }

  const sortedPlayers = [...players].sort((a, b) => (scores[b] || 0) - (scores[a] || 0))

  return (
    <div className="jeopardy">
      <div className="jeopardy-header">
        <h2>Jeopardy</h2>
        <div className="game-progress">
          Question {Math.min(currentQ + 1, questions.length)} of {questions.length}
        </div>
      </div>

      <div className="jeopardy-layout">
        <div className="jeopardy-main">
          {isGameOver ? (
            <div className="game-over-container">
              <h3>Game Over!</h3>
              <div className="final-scores">
                {sortedPlayers.map((p, i) => (
                  <div key={p} className={`final-score-item ${p === playerId ? 'you' : ''}`}>
                    <span className="rank">#{i + 1}</span>
                    <span className="name">{p}</span>
                    <span className="score">{scores[p] || 0}</span>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <>
              <div className="jeopardy-board">
                <div className="category-row">
                  {CATEGORIES.map(cat => (
                    <div key={cat} className="category-header">{cat}</div>
                  ))}
                </div>
                {VALUES.map((value, valIdx) => (
                  <div key={value} className="question-row">
                    {CATEGORIES.map((_, catIdx) => {
                      const q = getQuestionForCell(catIdx, valIdx)
                      return (
                        <button
                          key={`${catIdx}-${valIdx}`}
                          className={`question-cell ${q?.answered ? 'answered' : ''}`}
                          onClick={() => q && handleQuestionClick(q.index)}
                          disabled={!q || q.answered}
                        >
                          {q ? (q.answered ? '✓' : `$${value}`) : ''}
                        </button>
                      )
                    })}
                  </div>
                ))}
              </div>

              {selectedQ !== null && questions[selectedQ] && (
                <div className="question-reveal">
                  <div className="question-card">
                    <div className="question-category">{questions[selectedQ].category}</div>
                    <div className="question-value">${questions[selectedQ].value}</div>
                    <div className="question-text">{questions[selectedQ].question}</div>
                    
                    {result ? (
                      <div className={`answer-result ${result.correct ? 'correct' : 'incorrect'}`}>
                        {result.correct ? '✓ Correct!' : `✗ Wrong! Answer: ${result.answer}`}
                      </div>
                    ) : (
                      <form className="answer-form" onSubmit={handleSubmitAnswer}>
                        <input
                          type="text"
                          value={answer}
                          onChange={(e) => setAnswer(e.target.value)}
                          placeholder="Type your answer..."
                          autoFocus
                        />
                        <button type="submit">Submit</button>
                      </form>
                    )}
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        <div className="scoreboard">
          <h3>Scoreboard</h3>
          <div className="scores-list">
            {sortedPlayers.map((p, i) => (
              <div key={p} className={`score-item ${p === playerId ? 'you' : ''} ${i === 0 ? 'leading' : ''}`}>
                <span className="player-rank">#{i + 1}</span>
                <span className="player-name">{p} {p === playerId && '(You)'}</span>
                <span className="player-score">${scores[p] || 0}</span>
              </div>
            ))}
          </div>
          <div className="your-score">
            Your Score: <span>${myScore}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
