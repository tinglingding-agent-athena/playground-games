import { useState, useEffect, useRef, useCallback } from 'react'
import './App.css'
import Lobby from './components/Lobby'
import WaitingRoom from './components/WaitingRoom'
import TicTacToe from './components/TicTacToe'
import Jeopardy from './components/Jeopardy'

// Message types
const MSG_TYPE_ROOM_STATE = 'room_state'
const MSG_TYPE_GAME_STATE = 'game_state'
const MSG_TYPE_ERROR = 'error'
const MSG_TYPE_CREATE_ROOM = 'create_room'
const MSG_TYPE_JOIN_ROOM = 'join_room'
const MSG_TYPE_LEAVE_ROOM = 'leave_room'
const MSG_TYPE_START_GAME = 'start_game'
const MSG_TYPE_ANSWER = 'answer'
const MSG_TYPE_PLAYER_JOINED = 'player_joined'

function App() {
  const [wsStatus, setWsStatus] = useState('disconnected')
  const [view, setView] = useState('lobby') // lobby, waiting, game
  const [playerId] = useState(() => 'player_' + Math.random().toString(36).substr(2, 8))
  const [room, setRoom] = useState(null)
  const [game, setGame] = useState(null)
  const [gameId, setGameId] = useState(null)
  const [gameType, setGameType] = useState(null)
  const wsRef = useRef(null)

  // Initialize WebSocket
  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws')
    wsRef.current = ws

    ws.onopen = () => {
      setWsStatus('connected')
      console.log('Connected to game server')
    }

    ws.onclose = () => {
      setWsStatus('disconnected')
      console.log('Disconnected from game server')
    }

    ws.onerror = () => {
      setWsStatus('error')
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        handleMessage(msg)
      } catch (err) {
        console.error('Failed to parse message:', err)
      }
    }

    return () => ws.close()
  }, [])

  const handleMessage = (msg) => {
    switch (msg.type) {
      case MSG_TYPE_ROOM_STATE:
        if (msg.payload.room) {
          setRoom(msg.payload.room)
          setView('waiting')
        } else {
          setRoom(null)
          setView('lobby')
        }
        break

      case MSG_TYPE_GAME_STATE:
        setGame(msg.payload.game)
        setGameId(msg.payload.game_id)
        if (msg.payload.room) {
          setRoom(msg.payload.room)
        }
        if (msg.payload.game) {
          setView('game')
        }
        break

      case MSG_TYPE_ERROR:
        alert('Error: ' + msg.payload)
        break

      case MSG_TYPE_PLAYER_JOINED:
        if (msg.payload.room) {
          setRoom(msg.payload.room)
        }
        break

      default:
        console.log('Unknown message type:', msg.type)
    }
  }

  const sendMessage = useCallback((type, payload) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }))
    }
  }, [])

  const [gameMode, setGameMode] = useState('classic')

  const handleCreateRoom = useCallback((selectedGame, selectedMode = 'classic') => {
    setGameType(selectedGame)
    setGameMode(selectedMode)
    sendMessage(MSG_TYPE_CREATE_ROOM, {
      game_type: selectedGame,
      game_mode: selectedMode,
      player_id: playerId
    })
  }, [playerId, sendMessage])

  const handleJoinRoom = useCallback((code) => {
    sendMessage(MSG_TYPE_JOIN_ROOM, {
      code: code,
      player_id: playerId
    })
  }, [playerId, sendMessage])

  const handleLeaveRoom = useCallback(() => {
    if (room) {
      sendMessage(MSG_TYPE_LEAVE_ROOM, {
        code: room.code,
        player_id: playerId
      })
    }
    setRoom(null)
    setGame(null)
    setGameId(null)
    setView('lobby')
  }, [room, playerId, sendMessage])

  const handleStartGame = useCallback(() => {
    if (room) {
      sendMessage(MSG_TYPE_START_GAME, {
        code: room.code,
        player_id: playerId
      })
    }
  }, [room, playerId, sendMessage])

  const handleTicTacToeMove = useCallback((index) => {
    sendMessage('make_move', {
      game_id: gameId,
      player_id: playerId,
      index: index
    })
  }, [gameId, playerId, sendMessage])

  const handleJeopardyAnswer = useCallback((questionIndex, answer) => {
    // We need to find the actual game question index
    // For now, pass the answer directly
    sendMessage(MSG_TYPE_ANSWER, {
      game_id: gameId,
      player_id: playerId,
      answer: answer
    })
  }, [gameId, playerId, sendMessage])

  return (
    <div className="app">
      <header className="header">
        <h1>🎮 Playground Games</h1>
        <div className="header-info">
          <p className="subtitle">Real-time multiplayer games</p>
          <div className={`status ${wsStatus}`}>
            <span className="status-dot"></span>
            Server: {wsStatus}
          </div>
        </div>
        {view !== 'lobby' && (
          <button className="back-btn" onClick={handleLeaveRoom}>
            ← Back
          </button>
        )}
      </header>

      <main className="main-content">
        {view === 'lobby' && (
          <Lobby
            onCreateRoom={handleCreateRoom}
            onJoinRoom={handleJoinRoom}
            ws={wsRef.current}
          />
        )}

        {view === 'waiting' && room && (
          <WaitingRoom
            room={room}
            playerId={playerId}
            onStartGame={handleStartGame}
            onLeaveRoom={handleLeaveRoom}
          />
        )}

        {view === 'game' && game && (
          <>
            {gameType === 'tictactoe' && (
              <TicTacToe
                game={game}
                gameId={gameId}
                playerId={playerId}
                gameMode={gameMode}
                onMove={handleTicTacToeMove}
                ws={wsRef.current}
              />
            )}
            {gameType === 'jeopardy' && (
              <Jeopardy
                game={game}
                gameId={gameId}
                playerId={playerId}
                gameMode={gameMode}
                onAnswer={handleJeopardyAnswer}
                ws={wsRef.current}
              />
            )}
          </>
        )}
      </main>

      <footer className="footer">
        <p>Built with React + Go + WebSockets</p>
      </footer>
    </div>
  )
}

export default App
