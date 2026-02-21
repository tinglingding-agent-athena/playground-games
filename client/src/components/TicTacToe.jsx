import { useState, useEffect } from 'react'
import './TicTacToe.css'

export default function TicTacToe({ game, gameId, playerId, onMove, ws }) {
  const [board, setBoard] = useState(game?.board || Array(9).fill(''))
  const [turn, setTurn] = useState(game?.turn || 0)
  const [winner, setWinner] = useState(game?.winner || '')
  const [players, setPlayers] = useState(game?.players || ['', ''])

  useEffect(() => {
    if (game) {
      setBoard(game.board)
      setTurn(game.turn)
      setWinner(game.winner)
      setPlayers(game.players)
    }
  }, [game])

  const symbols = ['X', 'O']
  const currentSymbol = symbols[turn]
  const isMyTurn = players[turn] === playerId
  const playerIndex = players.indexOf(playerId)

  const handleCellClick = (index) => {
    if (winner || board[index] !== '' || !isMyTurn) return
    onMove(index)
  }

  const getWinnerMessage = () => {
    if (winner === 'draw') return "It's a Draw!"
    if (winner === playerId) return 'You Win! 🎉'
    if (winner && winner !== playerId) return 'You Lose 😔'
    return null
  }

  const winnerMessage = getWinnerMessage()

  return (
    <div className="tictactoe">
      <div className="ttt-header">
        <h2>Tic Tac Toe</h2>
        <div className="ttt-status">
          {winner ? (
            <span className="game-over">{winnerMessage}</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? `Your turn (${currentSymbol})` : `Waiting for ${players[turn] || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="ttt-board">
        {board.map((cell, index) => (
          <button
            key={index}
            className={`ttt-cell ${cell} ${cell === '' && isMyTurn && !winner ? 'clickable' : ''}`}
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
            X: {players[0] || 'Waiting...'}
          </div>
          <div className={`player-badge ${players[1] === playerId ? 'you' : ''} ${turn === 1 && !winner ? 'active' : ''}`}>
            O: {players[1] || 'Waiting...'}
          </div>
        </div>
      </div>
    </div>
  )
}
