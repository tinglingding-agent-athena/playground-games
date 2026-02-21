import { useState, useEffect } from 'react'
import './Uno.css'

const COLOR_NAMES = {
  red: '🔴 Red',
  yellow: '🟡 Yellow',
  green: '🟢 Green',
  blue: '🔵 Blue',
  wild: '🌈 Wild'
}

const CARD_DISPLAY = {
  '0': '0', '1': '1', '2': '2', '3': '3', '4': '4', '5': '5', '6': '6', '7': '7', '8': '8', '9': '9',
  'skip': '⛔',
  'reverse': '🔄',
  'draw2': '+2',
  'wild': '🌈',
  'wild4': '+4'
}

export default function Uno({ game, room, playerID, onMakeMove, ws }) {
  const [selectedCard, setSelectedCard] = useState(null)
  const [chosenColor, setChosenColor] = useState('')
  const [showColorPicker, setShowColorPicker] = useState(false)

  const myHand = game.hands?.[playerID] || []
  const currentCard = game.current_card
  const isMyTurn = game.players?.[game.current_player] === playerID

  useEffect(() => {
    if (ws) {
      const handleMessage = (msg) => {
        if (msg.type === 'game_state') {
          // Reset selection on game state update
          setSelectedCard(null)
          setShowColorPicker(false)
        }
      }
      ws.addEventListener('message', handleMessage)
      return () => ws.removeEventListener('message', handleMessage)
    }
  }, [ws])

  const handleCardClick = (card, index) => {
    if (!isMyTurn) return
    
    const canPlay = canPlayCard(card, currentCard)
    if (!canPlay) return

    setSelectedCard(index)
    if (card.color === 'wild') {
      setShowColorPicker(true)
    } else {
      playCard(index, '')
    }
  }

  const handleColorSelect = (color) => {
    setChosenColor(color)
    if (selectedCard !== null) {
      playCard(selectedCard, color)
    }
    setShowColorPicker(false)
    setSelectedCard(null)
  }

  const playCard = (cardIdx, color) => {
    onMakeMove({
      game_id: room.game_id,
      player_id: playerID,
      card_idx: cardIdx,
      chosen_color: color
    })
  }

  const canPlayCard = (card, current) => {
    if (!current) return true
    if (card.color === 'wild') return true
    if (current.color === 'wild') return true
    if (card.color === current.color) return true
    if (card.value === current.value) return true
    return false
  }

  const getCardClass = (card, index) => {
    let cls = 'uno-card'
    cls += ` card-${card.color}`
    if (card.value === 'skip' || card.value === 'reverse' || card.value === 'draw2' || card.value === 'wild' || card.value === 'wild4') {
      cls += ' card-special'
    }
    if (selectedCard === index) {
      cls += ' selected'
    }
    if (!isMyTurn || !canPlayCard(card, currentCard)) {
      cls += ' disabled'
    }
    return cls
  }

  const getCardDisplay = (card) => {
    return CARD_DISPLAY[card.value] || card.value
  }

  return (
    <div className="uno-game">
      <div className="uno-header">
        <h2>🎴 Uno</h2>
        <div className="uno-info">
          <span>Direction: {game.direction === 1 ? '🔄 Clockwise' : '🔄 Counter-Clockwise'}</span>
        </div>
      </div>

      <div className="uno-play-area">
        <div className="deck">
          <div className="uno-card card-back">
            <span>UNO</span>
          </div>
          <div className="deck-count">{game.deck?.length || 0}</div>
        </div>

        <div className="current-card">
          <div className={`uno-card card-${currentCard?.color || 'wild'}`}>
            <span className="card-value">{getCardDisplay(currentCard)}</span>
          </div>
          <div className="current-card-label">Current Card</div>
        </div>
      </div>

      <div className="players-info">
        {game.players?.map((p, idx) => (
          <div key={p} className={`player-info ${idx === game.current_player ? 'current-turn' : ''}`}>
            <span className="player-name">{p}</span>
            <span className="card-count">🃏 {game.hands?.[p]?.length || 0}</span>
          </div>
        ))}
      </div>

      {game.winner && (
        <div className="game-over">
          <h3>🏆 {game.winner === playerID ? 'You Won!' : `${game.winner} Won!`}</h3>
        </div>
      )}

      {isMyTurn && !game.winner && (
        <div className="my-hand">
          <h3>Your Hand</h3>
          <div className="cards-row">
            {myHand.map((card, index) => (
              <div
                key={index}
                className={getCardClass(card, index)}
                onClick={() => handleCardClick(card, index)}
              >
                <span className="card-value">{getCardDisplay(card)}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {showColorPicker && (
        <div className="color-picker-overlay">
          <div className="color-picker">
            <h3>Choose a Color</h3>
            <div className="color-options">
              {['red', 'yellow', 'green', 'blue'].map(color => (
                <button
                  key={color}
                  className={`color-btn color-${color}`}
                  onClick={() => handleColorSelect(color)}
                >
                  {COLOR_NAMES[color]}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {!isMyTurn && !game.winner && (
        <div className="waiting-message">
          Waiting for {game.players?.[game.current_player]}...
        </div>
      )}
    </div>
  )
}
