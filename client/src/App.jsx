import { useState, useEffect } from 'react'
import './App.css'

const games = [
  {
    id: 'tictactoe',
    name: 'Tic Tac Toe',
    description: 'Classic two-player strategy game',
    icon: '⭕',
    color: '#3b82f6'
  },
  {
    id: 'jeopardy',
    name: 'Jeopardy',
    description: 'Quiz game with categories and points',
    icon: '🎯',
    color: '#8b5cf6'
  }
]

function App() {
  const [wsStatus, setWsStatus] = useState('disconnected')

  useEffect(() => {
    // Check WebSocket connection status
    const ws = new WebSocket('ws://localhost:8080/ws')
    
    ws.onopen = () => setWsStatus('connected')
    ws.onclose = () => setWsStatus('disconnected')
    ws.onerror = () => setWsStatus('error')
    
    return () => ws.close()
  }, [])

  const handleGameClick = (gameId) => {
    // For now, just show an alert - routing will come later
    alert(`Starting ${gameId} game... (Game UI coming soon)`)
  }

  return (
    <div className="app">
      <header className="header">
        <h1>🎮 Playground Games</h1>
        <p className="subtitle">Real-time multiplayer games</p>
        <div className={`status ${wsStatus}`}>
          <span className="status-dot"></span>
          Server: {wsStatus}
        </div>
      </header>

      <main className="games-grid">
        {games.map(game => (
          <div 
            key={game.id} 
            className="game-card"
            onClick={() => handleGameClick(game.id)}
            style={{ '--accent-color': game.color }}
          >
            <div className="game-icon">{game.icon}</div>
            <h2 className="game-name">{game.name}</h2>
            <p className="game-description">{game.description}</p>
            <button className="play-button">Play Now</button>
          </div>
        ))}
      </main>

      <footer className="footer">
        <p>Built with React + Go + WebSockets</p>
      </footer>
    </div>
  )
}

export default App
