import { useState, useEffect, useRef } from 'react'
import './TicTacToe.css'

const MODE_INFO = {
  classic: { name: 'Classic', description: 'Normal rules' },
  fading: { name: 'Fading', description: 'Oldest move disappears after 4 moves' },
  speed: { name: 'Speed', description: '5 second timer per move' },
  infinite: { name: 'Infinite', description: "Board doesn't reset" }
}

export default function TicTacToe({ game, gameId, playerId, gameMode, onMove, ws, room }) {
  const [board, setBoard] = useState(game?.board || Array(9).fill(''))
  const [turn, setTurn] = useState(game?.turn || 0)
  const [winner, setWinner] = useState(game?.winner || '')
  const [players, setPlayers] = useState(game?.players || ['', ''])
  const [timer, setTimer] = useState(game?.move_timer || 5)
  const [timerActive, setTimerActive] = useState(false)
  const timerRef = useRef(null)

  const playerNames = room?.player_names || {}
  const playerIndices = room?.player_indices || {}
  const getPlayerName = (playerId) => playerNames[playerId] || playerId
  
  // Get display name with unique identifier
  const getPlayerDisplayName = (pId) => {
    const name = getPlayerName(pId)
    if (pId === playerId) {
      return `${name} (You)`
    }
    const index = playerIndices[pId]
    return index ? `${name} (Player ${index})` : name
  }

  const currentMode = gameMode || game?.game_mode || 'classic'

  useEffect(() => {
    if (game) {
      setBoard(game.board)
      setTurn(game.turn)
      setWinner(game.winner)
      setPlayers(game.players)
      setTimer(game.move_timer || 5)
      setTimerActive(game.timer_active || false)
    }
  }, [game])

  // Timer for speed mode
  useEffect(() => {
    if (currentMode === 'speed' && timerActive && !winner) {
      timerRef.current = setInterval(() => {
        setTimer((prev) => {
          if (prev <= 1) {
            // Time's up - skip turn
            return 5
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
  }, [currentMode, timerActive, winner, turn])

  const symbols = ['X', 'O']
  const currentSymbol = symbols[turn]
  const isMyTurn = players[turn] === playerId
  const playerIndex = players.indexOf(playerId)

  const handleCellClick = (index) => {
    if (winner || board[index] !== '' || !isMyTurn) return
    // Reset timer on move in speed mode
    if (currentMode === 'speed') {
      setTimer(5)
    }
    onMove(index)
  }

  const getWinnerMessage = () => {
    if (winner === 'draw') return "It's a Draw!"
    if (winner === playerId) return 'You Win! 🎉'
    if (winner && winner !== playerId) return 'You Lose 😔'
    return null
  }

  const winnerMessage = getWinnerMessage()

  // Get mode-specific UI hints
  const getModeHint = () => {
    switch (currentMode) {
      case 'fading':
        return '💨 Oldest move fades after each player makes 2 moves'
      case 'speed':
        return `⏱️ ${timer}s`
      case 'infinite':
        return '♾️ Board keeps filling - win when you can!'
      default:
        return null
    }
  }

  return (
    <div className="tictactoe">
      <div className="ttt-header">
        <h2>Tic Tac Toe</h2>
        <div className="mode-badge">{MODE_INFO[currentMode]?.name || currentMode}</div>
        <div className="ttt-status">
          {winner ? (
            <span className="game-over">{winnerMessage}</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? `Your turn (${currentSymbol})` : `Waiting for ${getPlayerName(players[turn]) || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      {getModeHint() && (
        <div className="mode-hint">{getModeHint()}</div>
      )}

      <div className="ttt-board">
        {board.map((cell, index) => (
          <button
            key={index}
            className={`ttt-cell ${cell} ${cell === '' && isMyTurn && !winner ? 'clickable' : ''} ${currentMode === 'fading' && cell !== '' ? 'fading-cell' : ''}`}
            onClick={() => handleCellClick(index)}
            disabled={!!winner || cell !== '' || !isMyTurn}
          >
            {cell}
          </button>
        ))}
      </div>

      <div className="ttt-info">
        <div className="players-info">
          <div className={`player-badge ${players[0] === playerId ? 'you' : ''} ${turn === 0 && !winner ? 'active' : ''}`}>
            X: {getPlayerName(players[0]) || 'Waiting...'}
          </div>
          <div className={`player-badge ${players[1] === playerId ? 'you' : ''} ${turn === 1 && !winner ? 'active' : ''}`}>
            O: {getPlayerName(players[1]) || 'Waiting...'}
          </div>
        </div>
      </div>
    </div>
  )
}
