# 🎮 Playground Games

Real-time multiplayer web games with a React frontend and Go backend using WebSockets.

## Features

- **Tic Tac Toe** - Classic two-player strategy game
- **Jeopardy** - Quiz game with categories and points

## Tech Stack

- **Frontend**: React + Vite
- **Backend**: Go with gorilla/websocket
- **Communication**: WebSockets for real-time gameplay

## Quick Start

### Prerequisites

- Node.js 18+
- Go 1.21+
- Git

### Running the Backend

```bash
cd server
go mod download
go run main.go
```

The server will start on `http://localhost:8080`

### Running the Frontend

```bash
cd client
npm install
npm run dev
```

The client will start on `http://localhost:5173`

## WebSocket API

### Connect
```
ws://localhost:8080/ws
```

### Message Types

#### Create Game
```json
{
  "type": "create_game",
  "payload": {
    "game_type": "tictactoe" | "jeopardy",
    "player_id": "player_123"
  }
}
```

#### Join Game
```json
{
  "type": "join_game",
  "payload": {
    "game_type": "tictactoe" | "jeopardy",
    "game_id": "game_abc123",
    "player_id": "player_456"
  }
}
```

#### TicTacToe Move
```json
{
  "type": "make_move",
  "payload": {
    "game_id": "game_abc123",
    "player_id": "player_123",
    "index": 4
  }
}
```

#### Jeopardy Answer
```json
{
  "type": "answer",
  "payload": {
    "game_id": "game_abc123",
    "player_id": "player_123",
    "answer": "Au"
  }
}
```

## Project Structure

```
playground-games/
├── client/          # React frontend
│   ├── src/
│   │   ├── App.jsx
│   │   └── App.css
│   └── package.json
├── server/          # Go backend
│   ├── main.go
│   └── go.mod
└── README.md
```

## License

MIT
