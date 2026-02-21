import { useState, useEffect } from 'react'
import './Mafia.css'

const ROLE_INFO = {
  mafia: { icon: '🎭', description: 'Work with other Mafia to eliminate villagers' },
  doctor: { icon: '👨‍⚕️', description: 'Save one player each night from being killed' },
  detective: { icon: '🕵️', description: 'Investigate one player each night to learn their role' },
  villager: { icon: '👤', description: 'Help identify and lynch the Mafia during the day' }
}

export default function Mafia({ game, room, playerID, onMakeMove, ws }) {
  const [selectedTarget, setSelectedTarget] = useState('')
  const [showInvestigation, setShowInvestigation] = useState(false)

  const myRole = game.roles?.[playerID]
  const isAlive = game.alive_players?.includes(playerID)
  const isNight = game.phase === 'night'
  const isDay = game.phase === 'day' || game.phase === 'lynch'

  useEffect(() => {
    if (ws) {
      const handleMessage = (msg) => {
        if (msg.type === 'game_state') {
          const payload = msg.payload
          if (payload.game?.investigation && payload.game.investigation !== '') {
            setShowInvestigation(true)
            setTimeout(() => setShowInvestigation(false), 5000)
          }
        }
      }
      ws.addEventListener('message', handleMessage)
      return () => ws.removeEventListener('message', handleMessage)
    }
  }, [ws])

  const handleAction = (action, target) => {
    if (!isAlive || (isNight && myRole === 'villager')) return

    onMakeMove({
      game_id: room.game_id,
      player_id: playerID,
      action: action,
      target: target
    })
    setSelectedTarget('')
  }

  const canTarget = (target) => {
    if (target === playerID) return false
    return game.alive_players?.includes(target)
  }

  const hasActed = () => {
    if (!game.night_actions) return false
    return Object.keys(game.night_actions).includes(playerID)
  }

  const hasVoted = () => {
    if (!game.votes) return false
    return Object.keys(game.votes).includes(playerID)
  }

  const getPlayerClass = (p) => {
    let cls = 'mafia-player'
    if (!game.alive_players?.includes(p)) {
      cls += ' eliminated'
    }
    if (p === game.kill_target && game.phase !== 'night') {
      cls += ' killed'
    }
    if (p === game.lynched_player) {
      cls += ' lynched'
    }
    return cls
  }

  return (
    <div className="mafia-game">
      <div className="mafia-header">
        <h2>🕵️ Mafia</h2>
        <div className="phase-indicator">
          <span className={`phase-badge phase-${game.phase}`}>
            {game.phase === 'night' && '🌙 Night'}
            {game.phase === 'day' && '☀️ Day'}
            {game.phase === 'lynch' && '🗳️ Voting'}
            {game.phase === 'gameover' && '🏁 Game Over'}
          </span>
          {game.day_number > 0 && <span>Day {game.day_number}</span>}
        </div>
      </div>

      <div className="role-reveal">
        {isAlive ? (
          <div className="my-role">
            <span className="role-icon">{ROLE_INFO[myRole]?.icon}</span>
            <span className="role-name">You are: {myRole?.toUpperCase()}</span>
            <p className="role-desc">{ROLE_INFO[myRole]?.description}</p>
          </div>
        ) : (
          <div className="my-role eliminated-role">
            <span>💀 You have been eliminated</span>
            <p>You can still watch the game</p>
          </div>
        )}
      </div>

      {showInvestigation && game.investigation && (
        <div className="investigation-result">
          <h3>🔍 Investigation Result</h3>
          <p>{game.investigation === 'mafia' ? '🚨 This player is MAFIA!' : '✅ This player is INNOCENT'}</p>
        </div>
      )}

      <div className="players-section">
        <h3>Players</h3>
        <div className="players-grid">
          {game.players?.map(p => (
            <div key={p} className={getPlayerClass(p)}>
              <span className="player-icon">{game.alive_players?.includes(p) ? '👤' : '💀'}</span>
              <span className="player-name">{p}</span>
              {game.phase !== 'night' && game.roles?.[p] && (
                <span className="player-role">{game.roles[p]}</span>
              )}
              {game.vote_counts?.[p] > 0 && (
                <span className="vote-count">🗳️ {game.vote_counts[p]}</span>
              )}
            </div>
          ))}
        </div>
      </div>

      {game.phase === 'night' && isAlive && myRole !== 'villager' && !hasActed() && (
        <div className="night-actions">
          <h3>🌙 Night Action</h3>
          <p>Select a target for your action:</p>
          
          {myRole === 'mafia' && (
            <div className="action-section">
              <p className="action-desc">Choose who the Mafia should eliminate:</p>
              <div className="targets">
                {game.alive_players?.filter(p => p !== playerID).map(p => (
                  <button
                    key={p}
                    className={`target-btn ${selectedTarget === p ? 'selected' : ''}`}
                    onClick={() => setSelectedTarget(p)}
                    disabled={!canTarget(p)}
                  >
                    {p}
                  </button>
                ))}
              </div>
              <button
                className="action-btn"
                onClick={() => handleAction('kill', selectedTarget)}
                disabled={!selectedTarget}
              >
                Confirm Kill
              </button>
            </div>
          )}

          {myRole === 'doctor' && (
            <div className="action-section">
              <p className="action-desc">Choose who to save:</p>
              <div className="targets">
                {game.alive_players?.map(p => (
                  <button
                    key={p}
                    className={`target-btn ${selectedTarget === p ? 'selected' : ''}`}
                    onClick={() => setSelectedTarget(p)}
                  >
                    {p}
                  </button>
                ))}
              </div>
              <button
                className="action-btn"
                onClick={() => handleAction('save', selectedTarget)}
                disabled={!selectedTarget}
              >
                Confirm Save
              </button>
            </div>
          )}

          {myRole === 'detective' && (
            <div className="action-section">
              <p className="action-desc">Choose who to investigate:</p>
              <div className="targets">
                {game.alive_players?.filter(p => p !== playerID).map(p => (
                  <button
                    key={p}
                    className={`target-btn ${selectedTarget === p ? 'selected' : ''}`}
                    onClick={() => setSelectedTarget(p)}
                    disabled={!canTarget(p)}
                  >
                    {p}
                  </button>
                ))}
              </div>
              <button
                className="action-btn"
                onClick={() => handleAction('investigate', selectedTarget)}
                disabled={!selectedTarget}
              >
                Confirm Investigation
              </button>
            </div>
          )}
        </div>
      )}

      {game.phase === 'night' && isAlive && myRole === 'villager' && (
        <div className="waiting-message">
          <p>😴 Villagers sleep during the night...</p>
          <p>Wait for the morning results</p>
        </div>
      )}

      {game.phase === 'night' && hasActed() && (
        <div className="waiting-message">
          <p>✅ Action submitted</p>
          <p>Wait for other players...</p>
        </div>
      )}

      {(game.phase === 'day' || game.phase === 'lynch') && isAlive && !hasVoted() && (
        <div className="day-voting">
          <h3>🗳️ Vote to Lynch</h3>
          <p>Discuss with others, then vote to eliminate a suspect:</p>
          <div className="targets">
            {game.alive_players?.filter(p => p !== playerID).map(p => (
              <button
                key={p}
                className={`target-btn ${selectedTarget === p ? 'selected' : ''}`}
                onClick={() => setSelectedTarget(p)}
                disabled={!canTarget(p)}
              >
                {p}
              </button>
            ))}
          </div>
          <button
            className="action-btn vote-btn"
            onClick={() => handleAction('vote', selectedTarget)}
            disabled={!selectedTarget}
          >
            Cast Vote
          </button>
        </div>
      )}

      {(game.phase === 'day' || game.phase === 'lynch') && hasVoted() && (
        <div className="waiting-message">
          <p>✅ Vote cast</p>
          <p>Wait for other players...</p>
        </div>
      )}

      {game.phase !== 'night' && game.phase !== 'gameover' && !isAlive && (
        <div className="waiting-message">
          <p>👻 You are spectating</p>
          <p>Watch the game unfold</p>
        </div>
      )}

      {game.kill_target && game.phase !== 'night' && (
        <div className="night-result">
          <h4>🌙 Last Night</h4>
          <p>{game.kill_target} was targeted!</p>
          {game.alive_players?.includes(game.kill_target) ? (
            <p>💚 But they were saved by the Doctor!</p>
          ) : (
            <p>💀 They were eliminated!</p>
          )}
        </div>
      )}

      {game.lynched_player && (
        <div className="lynch-result">
          <h4>🗳️ Lynching Result</h4>
          <p>{game.lynched_player} was lynched!</p>
          <p>Role: {game.roles?.[game.lynched_player]?.toUpperCase()}</p>
        </div>
      )}

      {game.winner && (
        <div className="game-over">
          <h3>🏆 Game Over!</h3>
          <p className="winner-text">
            {game.winner === 'mafia' && '🎭 Mafia Wins!'}
            {game.winner === 'villagers' && '👥 Villagers Win!'}
          </p>
        </div>
      )}
    </div>
  )
}
