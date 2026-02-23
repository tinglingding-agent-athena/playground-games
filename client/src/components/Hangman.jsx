import { useState, useEffect } from 'react'
import './Hangman.css'

const ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('')
const MAX_WRONG = 6

export default function Hangman({ game, gameId, playerId, onMove, ws, room }) {
  const [board, setBoard] = useState({ word: '', guessedLetters: [], wrongGuesses: 0, winner: '', players: [], turn: 0 })
  const [selectedLetter, setSelectedLetter] = useState('')

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

  useEffect(() => {
    if (game) {
      setBoard({
        word: game.word || '',
        guessedLetters: game.guessed_letters || [],
        wrongGuesses: game.wrong_guesses || 0,
        winner: game.winner || '',
        players: game.players || [],
        turn: game.turn || 0
      })
    }
  }, [game])

  const isMyTurn = board.players[board.turn] === playerId
  const isGameOver = board.winner !== ''

  const getDisplayWord = () => {
    return board.word.split('').map(letter => 
      board.guessedLetters.includes(letter) ? letter : '_'
    ).join(' ')
  }

  const handleLetterClick = (letter) => {
    if (!isMyTurn || isGameOver || board.guessedLetters.includes(letter)) return
    setSelectedLetter(letter)
    onMove(letter)
  }

  const getWinnerMessage = () => {
    if (board.winner === playerId) return 'You Win! 🎉'
    if (board.winner === 'lose') return 'Game Over - You Lose 😔'
    if (board.winner && board.winner !== playerId) return `Winner: ${board.winner}`
    return null
  }

  return (
    <div className="hangman">
      <div className="hangman-header">
        <h2>Hangman</h2>
        <div className="hangman-status">
          {isGameOver ? (
            <span className="game-over">{getWinnerMessage()}</span>
          ) : (
            <span className={isMyTurn ? 'your-turn' : 'waiting'}>
              {isMyTurn ? 'Your turn!' : `Waiting for ${getPlayerDisplayName(board.players[board.turn]) || 'opponent'}...`}
            </span>
          )}
        </div>
      </div>

      <div className="hangman-game">
        <div className="hangman-display">
          <div className="hangman-figure">
            <svg viewBox="0 0 100 120" className="hangman-svg">
              {/* Gallows */}
              <line x1="10" y1="115" x2="90" y2="115" stroke="#333" strokeWidth="3" />
              <line x1="30" y1="115" x2="30" y2="10" stroke="#333" strokeWidth="3" />
              <line x1="30" y1="10" x2="70" y2="10" stroke="#333" strokeWidth="3" />
              <line x1="70" y1="10" x2="70" y2="25" stroke="#333" strokeWidth="3" />
              
              {/* Hangman parts - show based on wrong guesses */}
              {board.wrongGuesses >= 1 && <circle cx="70" cy="35" r="10" fill="none" stroke="#333" strokeWidth="2" />}
              {board.wrongGuesses >= 2 && <line x1="70" y1="45" x2="70" y2="70" stroke="#333" strokeWidth="2" />}
              {board.wrongGuesses >= 3 && <line x1="70" y1="50" x2="55" y2="60" stroke="#333" strokeWidth="2" />}
              {board.wrongGuesses >= 4 && <line x1="70" y1="50" x2="85" y2="60" stroke="#333" strokeWidth="2" />}
              {board.wrongGuesses >= 5 && <line x1="70" y1="70" x2="60" y2="90" stroke="#333" strokeWidth="2" />}
              {board.wrongGuesses >= 6 && <line x1="70" y1="70" x2="80" y2="90" stroke="#333" strokeWidth="2" />}
            </svg>
          </div>
          
          <div className="hangman-word">
            {getDisplayWord()}
          </div>
          
          <div className="hangman-stats">
            Wrong guesses: {board.wrongGuesses} / {MAX_WRONG}
          </div>
        </div>

        <div className="hangman-keyboard">
          {ALPHABET.map(letter => {
            const isGuessed = board.guessedLetters.includes(letter)
            const isDisabled = isGuessed || isGameOver || !isMyTurn
            return (
              <button
                key={letter}
                className={`letter-btn ${isGuessed ? 'guessed' : ''} ${isDisabled ? 'disabled' : ''}`}
                onClick={() => handleLetterClick(letter)}
                disabled={isDisabled}
              >
                {letter}
              </button>
            )
          })}
        </div>
      </div>

      <div className="players-info">
        <div className={`player-badge ${board.players[0] === playerId ? 'you' : ''} ${board.turn === 0 && !isGameOver ? 'active' : ''}`}>
          Player 1: {getPlayerDisplayName(board.players[0]) || 'Waiting...'}
        </div>
        <div className={`player-badge ${board.players[1] === playerId ? 'you' : ''} ${board.turn === 1 && !isGameOver ? 'active' : ''}`}>
          Player 2: {getPlayerDisplayName(board.players[1]) || 'Waiting...'}
        </div>
      </div>
    </div>
  )
}
