/**
 * Playground Games - Cloudflare Worker with Durable Objects
 */

export class GameRoom {
  constructor(state, env) {
    this.state = state;
    this.env = env;
    this.room = null;
    this.game = null;
    this.clients = new Map();
  }

  async webSocketConnect(ws) {
    this.clients.set(ws, { playerId: null, playerName: null });
    
    ws.accept();
    
    ws.addEventListener('message', async (event) => {
      try {
        const msg = JSON.parse(event.data);
        await this.handleMessage(msg, ws);
      } catch (e) {
        console.error('Error handling message:', e);
      }
    });

    ws.addEventListener('close', () => {
      this.handleClose(ws);
    });
  }

  async handleMessage(msg, ws) {
    const { type, payload } = msg;

    switch (type) {
      case 'create_room':
        this.handleCreateRoom(payload, ws);
        break;
      case 'join_room':
        this.handleJoinRoom(payload, ws);
        break;
      case 'leave_room':
        this.handleLeaveRoom(payload, ws);
        break;
      case 'start_game':
        this.handleStartGame(payload, ws);
        break;
      case 'make_move':
        this.handleMakeMove(payload, ws);
        break;
      case 'answer':
        this.handleAnswer(payload, ws);
        break;
    }
  }

  generateRoomCode() {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
    let code = '';
    for (let i = 0; i < 6; i++) {
      code += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return code;
  }

  generateGameId() {
    return 'game_' + Math.random().toString(36).substr(2, 9);
  }

  send(ws, msg) {
    try { ws.send(JSON.stringify(msg)); } catch (e) { console.error('Send error:', e); }
  }

  broadcast(msg, excludeWs = null) {
    for (const [clientWs] of this.clients) {
      if (clientWs !== excludeWs) this.send(clientWs, msg);
    }
  }

  handleCreateRoom(payload, ws) {
    const { game_type, game_mode, player_id, player_name } = payload;
    if (!player_name || !player_name.trim()) {
      this.send(ws, { type: 'error', payload: 'Player name is required' });
      return;
    }
    this.room = {
      code: this.generateRoomCode(),
      host: player_id,
      players: [player_id],
      playerNames: { [player_id]: player_name.trim() },
      spectators: [],
      gameType: game_type,
      gameMode: game_mode || 'classic',
      gameId: null,
      status: 'waiting'
    };
    this.game = null;
    this.clients.set(ws, { playerId: player_id, playerName: player_name });
    this.send(ws, { type: 'room_state', payload: { room: this.getRoomState() } });
  }

  handleJoinRoom(payload, ws) {
    const { code, player_id, player_name } = payload;
    if (!player_name || !player_name.trim()) {
      this.send(ws, { type: 'error', payload: 'Player name is required' });
      return;
    }
    if (!this.room || this.room.code !== code) {
      this.send(ws, { type: 'error', payload: 'Room not found' });
      return;
    }
    
    // Games that support more than 2 players
    const multiPlayerGames = ['uno', 'mafia', 'trivia', 'memory'];
    const isMultiPlayerGame = this.room.gameType && multiPlayerGames.includes(this.room.gameType);
    const maxPlayers = isMultiPlayerGame ? 8 : 2;
    
    // Always try to add as player first (even if game in progress for multi-player games)
    const canJoinAsPlayer = this.room.players.length < maxPlayers && 
      (this.room.status !== 'playing' || isMultiPlayerGame);
    
    if (canJoinAsPlayer) {
      this.room.players.push(player_id);
      this.room.playerNames[player_id] = player_name.trim();
    } else if (this.room.status === 'playing') {
      // Game in progress and room full - join as spectator
      this.room.spectators.push(player_id);
      this.room.playerNames[player_id] = player_name.trim();
    } else {
      this.send(ws, { type: 'error', payload: 'Room is full' });
      return;
    }
    this.clients.set(ws, { playerId: player_id, playerName: player_name });
    
    // Send room state first
    this.send(ws, { type: 'room_state', payload: { room: this.getRoomState() } });
    
    // If game is in progress, send game state to the joining player
    if (this.room.status === 'playing' && this.game) {
      this.send(ws, { type: 'game_state', payload: { game_id: this.room.gameId, game: this.getGameState(), room: this.getRoomState() } });
    }
    
    this.broadcast({ type: 'room_state', payload: { room: this.getRoomState() } });
    this.broadcast({ type: 'player_joined', payload: { player_id, player_name, room: this.getRoomState() } });
  }

  handleLeaveRoom(payload, ws) {
    const { player_id } = payload;
    if (!this.room) return;
    const clientInfo = this.clients.get(ws);
    if (!clientInfo) return;
    
    this.room.players = this.room.players.filter(p => p !== player_id);
    delete this.room.playerNames[player_id];
    this.clients.delete(ws);

    if (this.room.players.length === 0) {
      this.room = null;
      this.game = null;
    } else if (player_id === this.room.host) {
      this.room.host = this.room.players[0];
    }
    this.broadcast({ type: 'room_state', payload: { room: this.room ? this.getRoomState() : null } });
  }

  handleStartGame(payload, ws) {
    const { player_id } = payload;
    if (!this.room || player_id !== this.room.host || this.room.players.length < 1) return;
    
    this.game = this.initializeGame(this.room.gameType, this.room.gameMode);
    this.room.gameId = this.generateGameId();
    this.room.status = 'playing';
    
    this.broadcast({ type: 'game_state', payload: { game_id: this.room.gameId, game: this.getGameState(), room: this.getRoomState() } });
  }

  handleMakeMove(payload, ws) {
    const { game_id, player_id, index, letter } = payload;
    if (!this.game || this.room.gameId !== game_id) return;
    const clientInfo = this.clients.get(ws);
    if (!clientInfo || clientInfo.playerId !== player_id) return;

    let updateNeeded = false;
    switch (this.room.gameType) {
      case 'tictactoe': updateNeeded = this.handleTicTacToeMove(player_id, index); break;
      case 'hangman': updateNeeded = this.handleHangmanMove(player_id, letter); break;
      case 'memory': updateNeeded = this.handleMemoryMove(player_id, index); break;
    }
    if (updateNeeded) {
      this.broadcast({ type: 'game_state', payload: { game_id: this.room.gameId, game: this.getGameState(), room: this.getRoomState() } });
    }
  }

  handleAnswer(payload, ws) {
    const { game_id, player_id, answer, idx } = payload;
    if (!this.game || this.room.gameId !== game_id) return;
    
    let updateNeeded = false;
    switch (this.room.gameType) {
      case 'jeopardy': updateNeeded = this.handleJeopardyAnswer(player_id, answer); break;
      case 'trivia': updateNeeded = this.handleTriviaAnswer(player_id, idx); break;
    }
    if (updateNeeded) {
      this.broadcast({ type: 'game_state', payload: { game_id: this.room.gameId, game: this.getGameState(), room: this.getRoomState() } });
    }
  }

  handleClose(ws) {
    const clientInfo = this.clients.get(ws);
    if (!clientInfo?.playerId || !this.room) return;
    
    this.room.players = this.room.players.filter(p => p !== clientInfo.playerId);
    delete this.room.playerNames[clientInfo.playerId];
    this.clients.delete(ws);

    if (this.room.players.length === 0) {
      this.room = null;
      this.game = null;
    } else if (clientInfo.playerId === this.room.host) {
      this.room.host = this.room.players[0];
    }
    this.broadcast({ type: 'room_state', payload: { room: this.room ? this.getRoomState() : null } });
  }

  initializeGame(gameType, gameMode) {
    switch (gameType) {
      case 'tictactoe':
        return { board: Array(9).fill(''), players: this.room.players.slice(0, 2), turn: 0, winner: '', move_history: [] };
      case 'hangman':
        const words = ['APPLE', 'BANANA', 'CHERRY', 'DRAGON', 'ELEPHANT', 'GARDEN', 'ISLAND', 'JUNGLE'];
        return { players: this.room.players.slice(0, 2), turn: 0, word: words[Math.floor(Math.random() * words.length)], guessed_letters: [], wrong_guesses: 0, winner: '' };
      case 'memory':
        const symbols = ['🍎', '🍌', '🍇', '🍊', '🥝', '🍓', '🥭', '🍍'];
        const cards = [...symbols, ...symbols].sort(() => Math.random() - 0.5);
        return { players: this.room.players.slice(0, 2), scores: {}, cards: cards.map(s => ({ value: s, flipped: false, matched: false })), flipped_cards: [], matched_pairs: 0, current_player: 0, can_flip: true, first_flip: -1 };
      case 'jeopardy':
        return this.initializeJeopardy();
      case 'trivia':
        return this.initializeTrivia();
      default:
        return null;
    }
  }

  initializeJeopardy() {
    const categories = ['Science', 'History', 'Geography', 'Entertainment', 'Sports'];
    const questions = [
      [{ q: 'What is H2O?', a: 'water', v: 100 }, { q: 'Mars?', a: 'mars', v: 200 }],
      [{ q: 'First US President?', a: 'george washington', v: 100 }, { q: 'WWII end?', a: '1945', v: 200 }],
      [{ q: 'Capital of France?', a: 'paris', v: 100 }, { q: 'Largest ocean?', a: 'pacific', v: 200 }],
      [{ q: 'Jurassic Park director?', a: 'steven spielberg', v: 100 }, { q: 'First iPhone year?', a: '2007', v: 200 }],
      [{ q: 'Soccer team size?', a: '11', v: 100 }, { q: 'Shuttlecock sport?', a: 'badminton', v: 200 }]
    ];
    return { players: this.room.players, scores: this.room.players.reduce((a, p) => { a[p] = 0; return a; }, {}), current_q: 0, questions: questions.map((c, i) => c.map(q => ({ category: categories[i], ...q }))).flat() };
  }

  initializeTrivia() {
    const questions = [
      { category: 'Science', question: 'Gold symbol?', options: ['Au', 'Ag', 'Fe', 'Cu'], correct: 0 },
      { category: 'History', question: 'Titanic sank?', options: ['1912', '1905', '1920', '1898'], correct: 0 },
      { category: 'Geography', question: 'Smallest country?', options: ['Monaco', 'Vatican City', 'San Marino', 'Liechtenstein'], correct: 1 },
      { category: 'Entertainment', question: 'Iron Man actor?', options: ['Chris Evans', 'Robert Downey Jr', 'Chris Hemsworth', 'Mark Ruffalo'], correct: 1 },
      { category: 'Sports', question: 'Olympic rings?', options: ['4', '5', '6', '7'], correct: 1 }
    ];
    return { players: this.room.players, scores: this.room.players.reduce((a, p) => { a[p] = 0; return a; }, {}), current_q: 0, questions, game_over: false };
  }

  handleTicTacToeMove(playerId, index) {
    if (!this.game || this.game.winner || this.game.board[index] || this.game.players[this.game.turn] !== playerId) return false;
    this.game.board[index] = this.game.turn === 0 ? 'X' : 'O';
    this.game.move_history.push(index);
    const wins = [[0,1,2],[3,4,5],[6,7,8],[0,3,6],[1,4,7],[2,5,8],[0,4,8],[2,4,6]];
    for (const [a,b,c] of wins) {
      if (this.game.board[a] && this.game.board[a] === this.game.board[b] && this.game.board[a] === this.game.board[c]) {
        this.game.winner = playerId;
        return true;
      }
    }
    if (!this.game.board.includes('')) { this.game.winner = 'draw'; return true; }
    this.game.turn = 1 - this.game.turn;
    return true;
  }

  handleHangmanMove(playerId, letter) {
    if (!this.game || this.game.winner || this.game.players[this.game.turn] !== playerId) return false;
    letter = letter.toUpperCase();
    if (this.game.guessed_letters.includes(letter)) return false;
    this.game.guessed_letters.push(letter);
    if (!this.game.word.includes(letter)) {
      this.game.wrong_guesses++;
      if (this.game.wrong_guesses >= 6) { this.game.winner = 'hangman'; return true; }
    }
    if (this.game.word.split('').every(l => this.game.guessed_letters.includes(l))) { this.game.winner = playerId; return true; }
    if (this.game.players.length > 1) this.game.turn = 1 - this.game.turn;
    return true;
  }

  handleMemoryMove(playerId, index) {
    if (!this.game || !this.game.can_flip || this.game.cards[index].flipped || this.game.cards[index].matched || this.game.players[this.game.current_player] !== playerId) return false;
    this.game.cards[index].flipped = true;
    this.game.flipped_cards.push(index);
    if (this.game.first_flip === -1) { this.game.first_flip = index; return true; }
    const firstIdx = this.game.first_flip;
    if (this.game.cards[firstIdx].value === this.game.cards[index].value) {
      this.game.cards[firstIdx].matched = true;
      this.game.cards[index].matched = true;
      this.game.matched_pairs++;
      if (!this.game.scores[playerId]) this.game.scores[playerId] = 0;
      this.game.scores[playerId] += 1;
      this.game.flipped_cards = [];
      this.game.first_flip = -1;
      if (this.game.matched_pairs >= this.game.cards.length / 2) this.game.game_over = true;
    } else {
      this.game.can_flip = false;
      setTimeout(() => {
        if (this.game) {
          this.game.cards[firstIdx].flipped = false;
          this.game.cards[index].flipped = false;
          this.game.flipped_cards = [];
          this.game.first_flip = -1;
          this.game.current_player = 1 - this.game.current_player;
          this.game.can_flip = true;
          this.broadcast({ type: 'game_state', payload: { game_id: this.room?.gameId, game: this.getGameState(), room: this.getRoomState() } });
        }
      }, 1000);
    }
    return true;
  }

  handleJeopardyAnswer(playerId, answer) {
    if (!this.game) return false;
    const q = this.game.questions[this.game.current_q];
    if (!q) return false;
    if (answer.toLowerCase().trim() === q.answer.toLowerCase().trim()) {
      if (!this.game.scores[playerId]) this.game.scores[playerId] = 0;
      this.game.scores[playerId] += q.value;
      this.game.current_q++;
      return true;
    }
    return false;
  }

  handleTriviaAnswer(playerId, idx) {
    if (!this.game || this.game.game_over) return false;
    const q = this.game.questions[this.game.current_q];
    if (!q) return false;
    if (idx === q.correct_idx) {
      if (!this.game.scores[playerId]) this.game.scores[playerId] = 0;
      this.game.scores[playerId] += 100;
      this.game.current_q++;
      if (this.game.current_q >= this.game.questions.length) this.game.game_over = true;
      return true;
    }
    return false;
  }

  getRoomState() {
    if (!this.room) return null;
    // Create player indices mapping (player_id -> 1-based index)
    const playerIndices = {};
    this.room.players.forEach((playerId, index) => {
      playerIndices[playerId] = index + 1;
    });
    return { 
      code: this.room.code, 
      host: this.room.host, 
      players: this.room.players, 
      player_names: this.room.playerNames, 
      player_indices: playerIndices,
      spectators: this.room.spectators, 
      game_type: this.room.gameType, 
      game_mode: this.room.gameMode, 
      game_id: this.room.gameId, 
      status: this.room.status 
    };
  }

  getGameState() {
    return this.game;
  }
}

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    if (url.pathname === '/health') {
      return new Response('OK', { status: 200 });
    }

    if (url.pathname === '/ws' || url.pathname.startsWith('/ws')) {
      try {
        const pair = new WebSocketPair();
        const [client, server] = Object.values(pair);
        const roomCode = url.searchParams.get('room') || 'default';
        const id = env.GAME_ROOM.idFromName(roomCode);
        const stub = env.GAME_ROOM.get(id);
        stub.webSocketConnect(server);
        return new Response(null, { status: 101, webSocket: client });
      } catch (e) {
        return new Response('Error: ' + e.message, { status: 500 });
      }
    }

    return new Response('Playground Games Worker', { status: 200 });
  }
};
