import { useState, useEffect, useRef } from 'react'
import './Jeopardy.css'

const CATEGORIES = ['Science', 'History', 'Geography']
const VALUES = [100, 200, 300]

const MODE_INFO = {
  classic: { name: 'Classic', description: 'Normal Jeopardy rules' },
  speed: { name: 'Speed Round', description: '10 second timer per question' },
  teams: { name: 'Teams', description: '2v2 team mode' }
}

export default function Jeopardy({ game, gameId, playerId, gameMode, onAnswer, ws, room }) {
  const [players, setPlayers] = useState([])
  const [scores, setScores] = useState({})
  const [teams, setTeams] = useState({})
  const [currentQ, setCurrentQ] = useState(0)
  const [questions, setQuestions] = useState([])
  const [selectedQ, setSelectedQ] = useState(null)
  const [answer, setAnswer] = useState('')

  const playerNames = room?.player_names || {}
  const playerIndices = room?.player_indices || {}
  const getPlayerName = (pId) => playerNames[pId] || pId
  
  const getPlayerDisplayName = (pId) => {
    const name = getPlayerName(pId)
    if (pId === playerId) {
      return `${name} (You)`
    }
    const index = playerIndices[pId]
    return index ? `${name} (Player ${index})` : name
  }
  const [result, setResult] = useState(null)
  const [answered, setAnswered] = useState(new Set())
  const [timer, setTimer] = useState(10)
  const [timerActive, setTimerActive] = useState(false)
  const timerRef = useRef(null)

  const currentMode = gameMode || game?.game_mode || 'classic'

  useEffect(() => {
    if (game) {
      setPlayers(game.players || [])
      setScores(game.scores || {})
      setTeams(game.teams || {})
      setCurrentQ(game.current_q || 0)
      setQuestions(game.questions || [])
      setTimer(game.timer || 10)
      setTimerActive(game.timer_active || false)
    }
  }, [game])

  // Timer for speed mode
  useEffect(() => {
    if (currentMode === 'speed' && timerActive && selectedQ !== null && !result) {
      timerRef.current = setInterval(() => {
        setTimer((prev) => {
          if (prev <= 1) {
            // Time's up - auto-submit empty answer
            onAnswer(selectedQ, '')
            return 10
          }
          return prev - 1
        })
      }, 1000)
    }
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current)
      }
    }
  }, [currentMode, timerActive, selectedQ, result, currentQ])

  // Reset timer when question changes
  useEffect(() => {
    if (selectedQ === null) {
      setTimer(10)
    }
  }, [selectedQ])

  const currentQuestion = questions[currentQ]
  const isGameOver = currentQ >= questions.length
  const myScore = scores[playerId] || 0

  // Calculate team scores for teams mode
  const getTeamScores = () => {
    const teamScores = { team1: 0, team2: 0 }
    if (teams.team1) {
      teams.team1.forEach(p => {
        teamScores.team1 += scores[p] || 0
      })
    }
    if (teams.team2) {
      teams.team2.forEach(p => {
        teamScores.team2 += scores[p] || 0
      })
    }
    return teamScores
  }

  const getMyTeam = () => {
    if (teams.team1?.includes(playerId)) return 'team1'
    if (teams.team2?.includes(playerId)) return 'team2'
    return null
  }

  const handleQuestionClick = (index) => {
    if (answered.has(index)) return
    setSelectedQ(index)
    setResult(null)
    setAnswer('')
    setTimer(currentMode === 'speed' ? 10 : 0)
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
  const teamScores = currentMode === 'teams' ? getTeamScores() : null

  return (
    <div className="jeopardy">
      <div className="jeopardy-header">
        <h2>Jeopardy</h2>
        <div className="mode-badge">{MODE_INFO[currentMode]?.name || currentMode}</div>
        <div className="game-progress">
          Question {Math.min(currentQ + 1, questions.length)} of {questions.length}
        </div>
        {currentMode === 'speed' && selectedQ !== null && !result && (
          <div className="timer-display">⏱️ {timer}s</div>
        )}
      </div>

      <div className="jeopardy-layout">
        <div className="jeopardy-main">
          {isGameOver ? (
            <div className="game-over-container">
              <h3>Game Over!</h3>
              {currentMode === 'teams' ? (
                <div className="team-results">
                  <div className={`team-score ${teamScores.team1 > teamScores.team2 ? 'winner' : ''}`}>
                    <span className="team-name">Team 1</span>
                    <span className="team-members">{teams.team1?.join(', ') || 'No players'}</span>
                    <span className="team-points">${teamScores.team1}</span>
                  </div>
                  <div className="vs">VS</div>
                  <div className={`team-score ${teamScores.team2 > teamScores.team1 ? 'winner' : ''}`}>
                    <span className="team-name">Team 2</span>
                    <span className="team-members">{teams.team2?.join(', ') || 'No players'}</span>
                    <span className="team-points">${teamScores.team2}</span>
                  </div>
                </div>
              ) : (
                <div className="final-scores">
                  {sortedPlayers.map((p, i) => (
                    <div key={p} className={`final-score-item ${p === playerId ? 'you' : ''}`}>
                      <span className="rank">#{i + 1}</span>
                      <span className="name">{getPlayerDisplayName(p)}</span>
                      <span className="score">{scores[p] || 0}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : (
            <>
              {currentMode === 'teams' && (
                <div className="team-indicator">
                  You're on: <strong>{getMyTeam() === 'team1' ? 'Team 1 🔵' : getMyTeam() === 'team2' ? 'Team 2 🔴' : 'No Team'}</strong>
                </div>
              )}
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
          <h3>{currentMode === 'teams' ? 'Team Scores' : 'Scoreboard'}</h3>
          {currentMode === 'teams' ? (
            <div className="team-scores">
              <div className={`team-item ${getMyTeam() === 'team1' ? 'your-team' : ''}`}>
                <span className="team-label">🔵 Team 1</span>
                <span className="team-score-value">${teamScores.team1}</span>
              </div>
              <div className={`team-item ${getMyTeam() === 'team2' ? 'your-team' : ''}`}>
                <span className="team-label">🔴 Team 2</span>
                <span className="team-score-value">${teamScores.team2}</span>
              </div>
            </div>
          ) : (
            <div className="scores-list">
              {sortedPlayers.map((p, i) => (
                <div key={p} className={`score-item ${p === playerId ? 'you' : ''} ${i === 0 ? 'leading' : ''}`}>
                  <span className="player-rank">#{i + 1}</span>
                  <span className="player-name">{getPlayerDisplayName(p)}</span>
                  <span className="player-score">${scores[p] || 0}</span>
                </div>
              ))}
            </div>
          )}
          {currentMode !== 'teams' && (
            <div className="your-score">
              Your Score: <span>${myScore}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
