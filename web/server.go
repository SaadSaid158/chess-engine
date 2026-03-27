package web

import (
	"chessengine/engine"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// GameState holds the current game state
type GameState struct {
	Board    *engine.Board
	TT       *engine.TransTable
	Search   *engine.SearchInfo
	GameOver bool
	Result   string
	MoveList []MoveRecord
	mu       sync.Mutex
}

type MoveRecord struct {
	UCI  string `json:"uci"`
	From string `json:"from"`
	To   string `json:"to"`
	Side string `json:"side"`
}

type MoveRequest struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Promo string `json:"promo,omitempty"`
}

type MoveResponse struct {
	OK             bool         `json:"ok"`
	Error          string       `json:"error,omitempty"`
	PlayerFEN      string       `json:"playerFen,omitempty"`
	PlayerFrom     string       `json:"playerFrom,omitempty"`
	PlayerTo       string       `json:"playerTo,omitempty"`
	EngineMove     string       `json:"engineMove,omitempty"`
	EngineFrom     string       `json:"engineFrom,omitempty"`
	EngineTo       string       `json:"engineTo,omitempty"`
	EnginePromo    string       `json:"enginePromo,omitempty"`
	FEN            string       `json:"fen"`
	GameOver       bool         `json:"gameOver"`
	Result         string       `json:"result,omitempty"`
	Eval           int          `json:"eval"`
	Depth          int          `json:"depth"`
	Nodes          uint64       `json:"nodes"`
	TimeMs         int64        `json:"timeMs"`
	NPS            uint64       `json:"nps"`
	LegalMoves     []string     `json:"legalMoves,omitempty"`
	MoveList       []MoveRecord `json:"moveList,omitempty"`
	Captured       CapturedInfo `json:"captured"`
	PlayerGameOver bool         `json:"playerGameOver"`
	PlayerResult   string       `json:"playerResult,omitempty"`
	PlayerCaptured CapturedInfo `json:"playerCaptured"`
}

type CapturedInfo struct {
	White []string `json:"white"` // pieces captured by white (black pieces)
	Black []string `json:"black"` // pieces captured by black (white pieces)
}

type BoardResponse struct {
	FEN        string       `json:"fen"`
	Side       string       `json:"side"`
	GameOver   bool         `json:"gameOver"`
	Result     string       `json:"result,omitempty"`
	LegalMoves []string     `json:"legalMoves"`
	InCheck    bool         `json:"inCheck"`
	MoveList   []MoveRecord `json:"moveList"`
	Captured   CapturedInfo `json:"captured"`
}

type NewGameRequest struct {
	Color string `json:"color"`
	Depth int    `json:"depth"`
}

type Server struct {
	game        *GameState
	depth       int
	playerColor int
}

func NewServer() *Server {
	engine.InitZobrist()
	engine.InitAttacks()
	engine.InitEval()
	return &Server{depth: 8}
}

func (s *Server) newGame(playerColor int) {
	board := engine.NewBoard()
	tt := engine.NewTransTable(128) // 128 MB TT
	search := engine.NewSearchInfo(board, tt)
	s.game = &GameState{
		Board:    board,
		TT:       tt,
		Search:   search,
		MoveList: []MoveRecord{},
	}
	s.playerColor = playerColor
}

func (s *Server) getLegalMoves() []string {
	var ml engine.MoveList
	s.game.Board.GenerateMoves(&ml)
	moves := make([]string, ml.Count)
	for i := 0; i < ml.Count; i++ {
		moves[i] = ml.Moves[i].String()
	}
	return moves
}

func (s *Server) getCaptured() CapturedInfo {
	ci := CapturedInfo{White: []string{}, Black: []string{}}
	// Count current pieces and compare to starting material
	startCount := [2][6]int{
		{8, 2, 2, 2, 1, 1}, // white
		{8, 2, 2, 2, 1, 1}, // black
	}
	pieceNames := [6]string{"pawn", "knight", "bishop", "rook", "queen", "king"}
	for color := 0; color < 2; color++ {
		for piece := 0; piece < 6; piece++ {
			cur := engine.PopCount(s.game.Board.Pieces[color][piece])
			missing := startCount[color][piece] - cur
			for j := 0; j < missing; j++ {
				if color == engine.White {
					ci.Black = append(ci.Black, pieceNames[piece]) // black captured white's piece
				} else {
					ci.White = append(ci.White, pieceNames[piece])
				}
			}
		}
	}
	return ci
}

func (s *Server) checkGameOver() {
	var ml engine.MoveList
	s.game.Board.GenerateMoves(&ml)
	if ml.Count == 0 {
		s.game.GameOver = true
		if s.game.Board.InCheck() {
			if s.game.Board.Side == engine.White {
				s.game.Result = "black"
			} else {
				s.game.Result = "white"
			}
		} else {
			s.game.Result = "draw"
		}
		return
	}
	if s.game.Board.HalfMove >= 100 {
		s.game.GameOver = true
		s.game.Result = "draw"
		return
	}
	if s.game.Board.IsRepetition() {
		s.game.GameOver = true
		s.game.Result = "draw"
		return
	}
	if s.game.Board.IsInsufficientMaterial() {
		s.game.GameOver = true
		s.game.Result = "draw"
	}
}

func (s *Server) engineMove() (engine.Move, int, uint64, int64) {
	s.game.Search.Board = s.game.Board
	s.game.Search.Timer.SetFixedDepth(s.depth)
	start := time.Now()
	bestMove := s.game.Search.IterativeDeepening()
	elapsed := time.Since(start).Milliseconds()
	return bestMove, s.depth, s.game.Search.Nodes, elapsed
}

func (s *Server) Start(port int) {
	s.newGame(engine.White)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/move", s.handleMove)
	mux.HandleFunc("/api/newgame", s.handleNewGame)
	mux.HandleFunc("/api/undo", s.handleUndo)
	mux.HandleFunc("/api/legalmoves", s.handleLegalMoves)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Chess Web UI running at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	s.game.mu.Lock()
	defer s.game.mu.Unlock()
	resp := BoardResponse{
		FEN:        s.game.Board.FEN(),
		GameOver:   s.game.GameOver,
		Result:     s.game.Result,
		LegalMoves: s.getLegalMoves(),
		InCheck:    s.game.Board.InCheck(),
		MoveList:   s.game.MoveList,
		Captured:   s.getCaptured(),
	}
	if s.game.Board.Side == engine.White {
		resp.Side = "white"
	} else {
		resp.Side = "black"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.game.mu.Lock()
	defer s.game.mu.Unlock()

	var req MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, MoveResponse{OK: false, Error: "invalid request"})
		return
	}
	if s.game.GameOver {
		writeJSON(w, MoveResponse{OK: false, Error: "game is over", FEN: s.game.Board.FEN(), GameOver: true, Result: s.game.Result})
		return
	}

	moveStr := req.From + req.To
	if req.Promo != "" {
		moveStr += req.Promo
	}
	m := s.game.Board.ParseMove(moveStr)
	if m == engine.NullMove {
		writeJSON(w, MoveResponse{OK: false, Error: "illegal move", FEN: s.game.Board.FEN()})
		return
	}

	// Record player move
	playerSide := "white"
	if s.game.Board.Side == engine.Black {
		playerSide = "black"
	}
	s.game.MoveList = append(s.game.MoveList, MoveRecord{
		UCI: m.String(), From: req.From, To: req.To, Side: playerSide,
	})

	s.game.Board.MakeMove(m)
	s.game.Search.Board = s.game.Board

	// Capture the intermediate state after player's move
	playerFEN := s.game.Board.FEN()
	playerCaptured := s.getCaptured()

	s.checkGameOver()
	if s.game.GameOver {
		writeJSON(w, MoveResponse{
			OK: true, FEN: s.game.Board.FEN(), GameOver: true, Result: s.game.Result,
			PlayerFEN: playerFEN, PlayerFrom: req.From, PlayerTo: req.To,
			PlayerCaptured: playerCaptured, PlayerGameOver: true, PlayerResult: s.game.Result,
			MoveList: s.game.MoveList, Captured: s.getCaptured(),
		})
		return
	}

	engineM, depth, nodes, elapsed := s.engineMove()
	if engineM == engine.NullMove {
		s.checkGameOver()
		writeJSON(w, MoveResponse{
			OK: true, FEN: s.game.Board.FEN(), GameOver: s.game.GameOver, Result: s.game.Result,
			PlayerFEN: playerFEN, PlayerFrom: req.From, PlayerTo: req.To,
			PlayerCaptured: playerCaptured,
			MoveList:       s.game.MoveList, Captured: s.getCaptured(),
		})
		return
	}

	eval := engine.Evaluate(s.game.Board)
	fromSq := engineM.From()
	toSq := engineM.To()
	promoStr := ""
	if engineM.IsPromotion() {
		promoChars := [4]string{"n", "b", "r", "q"}
		promoStr = promoChars[engineM.Flags()-engine.FlagPromoN]
	}

	engSide := "white"
	if s.game.Board.Side == engine.Black {
		engSide = "black"
	}
	s.game.MoveList = append(s.game.MoveList, MoveRecord{
		UCI: engineM.String(), From: squareStr(fromSq), To: squareStr(toSq), Side: engSide,
	})

	s.game.Board.MakeMove(engineM)
	s.game.Search.Board = s.game.Board
	s.checkGameOver()

	nps := uint64(0)
	if elapsed > 0 {
		nps = nodes * 1000 / uint64(elapsed)
	}

	writeJSON(w, MoveResponse{
		OK: true, EngineMove: engineM.String(), EngineFrom: squareStr(fromSq),
		EngineTo: squareStr(toSq), EnginePromo: promoStr, FEN: s.game.Board.FEN(),
		GameOver: s.game.GameOver, Result: s.game.Result, Eval: eval, Depth: depth,
		Nodes: nodes, TimeMs: elapsed, NPS: nps, LegalMoves: s.getLegalMoves(),
		MoveList: s.game.MoveList, Captured: s.getCaptured(),
		PlayerFEN: playerFEN, PlayerFrom: req.From, PlayerTo: req.To,
		PlayerCaptured: playerCaptured,
	})
}

func (s *Server) handleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.game.mu.Lock()
	defer s.game.mu.Unlock()

	var req NewGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Color = "white"
		req.Depth = 8
	}
	if req.Depth < 1 {
		req.Depth = 8
	}
	if req.Depth > 20 {
		req.Depth = 20
	}
	s.depth = req.Depth
	playerColor := engine.White
	if req.Color == "black" {
		playerColor = engine.Black
	}
	s.newGame(playerColor)

	resp := MoveResponse{
		OK: true, FEN: s.game.Board.FEN(), LegalMoves: s.getLegalMoves(),
		MoveList: s.game.MoveList, Captured: s.getCaptured(),
	}

	if playerColor == engine.Black {
		engineM, depth, nodes, elapsed := s.engineMove()
		if engineM != engine.NullMove {
			eval := engine.Evaluate(s.game.Board)
			fromSq := engineM.From()
			toSq := engineM.To()
			promoStr := ""
			if engineM.IsPromotion() {
				promoChars := [4]string{"n", "b", "r", "q"}
				promoStr = promoChars[engineM.Flags()-engine.FlagPromoN]
			}
			s.game.MoveList = append(s.game.MoveList, MoveRecord{
				UCI: engineM.String(), From: squareStr(fromSq), To: squareStr(toSq), Side: "white",
			})
			s.game.Board.MakeMove(engineM)
			s.game.Search.Board = s.game.Board

			nps := uint64(0)
			if elapsed > 0 {
				nps = nodes * 1000 / uint64(elapsed)
			}
			resp.EngineMove = engineM.String()
			resp.EngineFrom = squareStr(fromSq)
			resp.EngineTo = squareStr(toSq)
			resp.EnginePromo = promoStr
			resp.FEN = s.game.Board.FEN()
			resp.Eval = eval
			resp.Depth = depth
			resp.Nodes = nodes
			resp.TimeMs = elapsed
			resp.NPS = nps
			resp.LegalMoves = s.getLegalMoves()
			resp.MoveList = s.game.MoveList
			resp.Captured = s.getCaptured()
		}
	}
	writeJSON(w, resp)
}

func (s *Server) handleUndo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.game.mu.Lock()
	defer s.game.mu.Unlock()

	if s.game.Board.HistPly >= 2 {
		s.game.Board.UnmakeMove()
		s.game.Board.UnmakeMove()
		s.game.GameOver = false
		s.game.Result = ""
		s.game.Search.Board = s.game.Board
		if len(s.game.MoveList) >= 2 {
			s.game.MoveList = s.game.MoveList[:len(s.game.MoveList)-2]
		}
	}
	writeJSON(w, MoveResponse{
		OK: true, FEN: s.game.Board.FEN(), LegalMoves: s.getLegalMoves(),
		MoveList: s.game.MoveList, Captured: s.getCaptured(),
	})
}

func (s *Server) handleLegalMoves(w http.ResponseWriter, r *http.Request) {
	s.game.mu.Lock()
	defer s.game.mu.Unlock()
	square := r.URL.Query().Get("square")
	if square == "" {
		writeJSON(w, map[string]interface{}{"moves": s.getLegalMoves()})
		return
	}
	sq := engine.StringToSquare(square)
	if sq == engine.NoSquare {
		writeJSON(w, map[string]interface{}{"moves": []string{}})
		return
	}
	var ml engine.MoveList
	s.game.Board.GenerateMoves(&ml)
	var moves []string
	for i := 0; i < ml.Count; i++ {
		m := ml.Moves[i]
		if m.From() == sq {
			moves = append(moves, squareStr(m.To()))
		}
	}
	writeJSON(w, map[string]interface{}{"moves": moves})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func squareStr(sq int) string {
	file := sq % 8
	rank := sq / 8
	return string([]byte{byte('a' + file), byte('1' + rank)})
}
