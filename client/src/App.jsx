import { useState, useEffect, useRef, useCallback } from 'react'
import './App.css'
import Lobby from './components/Lobby'
import WaitingRoom from './components/WaitingRoom'
import TicTacToe from './components/TicTacToe'
import Jeopardy from './components/Jeopardy'
import Hangman from './components/Hangman'
import Memory from './components/Memory'
import Battleship from './components/Battleship'
import Trivia from './components/Trivia'
import RPS from './components/RPS'
import ConnectFour from './components/ConnectFour'
import Checkers from './components/Checkers'
import DotsBoxes from './components/DotsBoxes'
import Uno from './components/Uno'
import Mafia from './components/Mafia'

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
  const [playerName, setPlayerName] = useState(() => localStorage.getItem('playerName') || '')
  const [playerId] = useState(() => 'player_' + Math.random().toString(36).substr(2, 8))
  const [room, setRoom] = useState(null)
  const [game, setGame] = useState(null)
  const [gameId, setGameId] = useState(null)
  const [gameType, setGameType] = useState(null)
  const wsRef = useRef(null)

  // Get WebSocket URL from environment or use default
  const wsUrl = import.meta.env.VITE_WS_URL || 'wss://api.tinglingding.win/ws';

  // Initialize WebSocket
  useEffect(() => {
    const ws = new WebSocket(wsUrl)
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
          // If room is already playing, set gameType so correct component renders
          if (msg.payload.room.status === 'playing' && msg.payload.room.game_type) {
            setGameType(msg.payload.room.game_type)
          }
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
          // Set gameType from room if not already set
          if (msg.payload.room.game_type && !gameType) {
            setGameType(msg.payload.room.game_type)
          }
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
      player_id: playerId,
      player_name: playerName
    })
  }, [playerId, playerName, sendMessage])

  const handleJoinRoom = useCallback((code) => {
    sendMessage(MSG_TYPE_JOIN_ROOM, {
      code: code,
      player_id: playerId,
      player_name: playerName
    })
  }, [playerId, playerName, sendMessage])

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
            playerName={playerName}
            setPlayerName={setPlayerName}
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
                room={room}
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
            {gameType === 'hangman' && (
              <Hangman
                game={game}
                gameId={gameId}
                playerId={playerId}
                onMove={(letter) => sendMessage('make_move', { game_id: gameId, player_id: playerId, letter })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'memory' && (
              <Memory
                game={game}
                gameId={gameId}
                playerId={playerId}
                onMove={(cardIdx) => sendMessage('make_move', { game_id: gameId, player_id: playerId, card_idx: cardIdx })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'battleship' && (
              <Battleship
                game={game}
                gameId={gameId}
                playerId={playerId}
                onMove={(x, y) => sendMessage('make_move', { game_id: gameId, player_id: playerId, x, y })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'trivia' && (
              <Trivia
                game={game}
                gameId={gameId}
                playerId={playerId}
                onAnswer={(idx) => sendMessage('make_move', { game_id: gameId, player_id: playerId, idx })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'rps' && (
              <RPS
                game={game}
                gameId={gameId}
                playerId={playerId}
                onMove={(move) => sendMessage('make_move', { game_id: gameId, player_id: playerId, move })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'connectfour' && (
              <ConnectFour
                game={game}
                gameId={gameId}
                playerId={playerId}
                onMove={(column) => sendMessage('make_move', { game_id: gameId, player_id: playerId, column })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'checkers' && (
              <Checkers
                game={game}
                gameId={gameId}
                playerId={playerId}
                onMove={(fromRow, fromCol, toRow, toCol) => sendMessage('make_move', { game_id: gameId, player_id: playerId, from_row: fromRow, from_col: fromCol, to_row: toRow, to_col: toCol })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'dotsboxes' && (
              <DotsBoxes
                game={game}
                gameId={gameId}
                playerId={playerId}
                onMove={(type, row, col) => sendMessage('make_move', { game_id: gameId, player_id: playerId, type, row, col })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'uno' && (
              <Uno
                game={game}
                room={room}
                playerId={playerId}
                onMakeMove={(payload) => sendMessage('make_move', { game_id: gameId, player_id: playerId, ...payload })}
                ws={wsRef.current}
              />
            )}
            {gameType === 'mafia' && (
              <Mafia
                game={game}
                room={room}
                playerId={playerId}
                onMakeMove={(payload) => sendMessage('make_move', { game_id: gameId, player_id: playerId, ...payload })}
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
