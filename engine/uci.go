package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// UCI implements the Universal Chess Interface protocol
type UCI struct {
	Board    *Board
	TT       *TransTable
	Search   *SearchInfo
	HashMB   int
	Contempt int
}

// NewUCI creates a new UCI handler
func NewUCI() *UCI {
	InitZobrist()
	InitAttacks()
	InitEval()

	board := NewBoard()
	hashMB := 64
	tt := NewTransTable(hashMB)
	search := NewSearchInfo(board, tt)
	contempt := DefaultContempt
	search.Contempt = contempt

	return &UCI{
		Board:    board,
		TT:       tt,
		Search:   search,
		HashMB:   hashMB,
		Contempt: contempt,
	}
}

// Loop is the main UCI command loop
func (u *UCI) Loop() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
		cmd := tokens[0]

		switch cmd {
		case "uci":
			u.handleUCI()
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			u.handleNewGame()
		case "setoption":
			u.handleSetOption(tokens)
		case "position":
			u.handlePosition(tokens)
		case "go":
			u.handleGo(tokens)
		case "stop":
			u.handleStop()
		case "quit":
			return
		case "perft":
			u.handlePerft(tokens)
		case "d":
			u.handleDisplay()
		case "eval":
			fmt.Printf("Eval: %d cp\n", Evaluate(u.Board))
		}
	}
}

func (u *UCI) handleUCI() {
	fmt.Println("id name GoChess 1.0")
	fmt.Println("id author Chess Engine")
	fmt.Println("option name Hash type spin default 64 min 1 max 2048")
	fmt.Println("option name Contempt type spin default 12 min -100 max 100")
	fmt.Println("uciok")
}

func (u *UCI) handleNewGame() {
	u.Board = NewBoard()
	u.TT.Clear()
	u.Search = NewSearchInfo(u.Board, u.TT)
	u.Search.Contempt = u.Contempt
}

func (u *UCI) handleSetOption(tokens []string) {
	nameParts := []string{}
	value := ""
	inName := false
	inValue := false

	for i := 1; i < len(tokens); i++ {
		switch tokens[i] {
		case "name":
			inName = true
			inValue = false
		case "value":
			inName = false
			inValue = true
		default:
			if inName {
				nameParts = append(nameParts, tokens[i])
			} else if inValue {
				if value != "" {
					value += " "
				}
				value += tokens[i]
			}
		}
	}

	name := strings.ToLower(strings.Join(nameParts, " "))
	switch name {
	case "hash":
		if mb, err := strconv.Atoi(value); err == nil {
			if mb < 1 {
				mb = 1
			}
			if mb > 2048 {
				mb = 2048
			}
			u.HashMB = mb
			u.TT = NewTransTable(mb)
			u.Search = NewSearchInfo(u.Board, u.TT)
			u.Search.Contempt = u.Contempt
		}
	case "contempt":
		if cp, err := strconv.Atoi(value); err == nil {
			if cp < -100 {
				cp = -100
			}
			if cp > 100 {
				cp = 100
			}
			u.Contempt = cp
			u.Search.Contempt = cp
		}
	}
}

func (u *UCI) handlePosition(tokens []string) {
	idx := 1
	if idx >= len(tokens) {
		return
	}

	if tokens[idx] == "startpos" {
		u.Board = NewBoard()
		idx++
	} else if tokens[idx] == "fen" {
		idx++
		fenParts := []string{}
		for idx < len(tokens) && tokens[idx] != "moves" {
			fenParts = append(fenParts, tokens[idx])
			idx++
		}
		fen := strings.Join(fenParts, " ")
		u.Board = BoardFromFEN(fen)
	}

	// Apply moves
	if idx < len(tokens) && tokens[idx] == "moves" {
		idx++
		for idx < len(tokens) {
			m := u.Board.ParseMove(tokens[idx])
			if m != NullMove {
				u.Board.MakeMove(m)
			}
			idx++
		}
	}

	u.Search.Board = u.Board
}

func (u *UCI) handleGo(tokens []string) {
	u.Search.Board = u.Board

	var wtime, btime, winc, binc, movetime, depth, movestogo int
	infinite := false

	for i := 1; i < len(tokens); i++ {
		switch tokens[i] {
		case "wtime":
			if i+1 < len(tokens) {
				wtime, _ = strconv.Atoi(tokens[i+1])
				i++
			}
		case "btime":
			if i+1 < len(tokens) {
				btime, _ = strconv.Atoi(tokens[i+1])
				i++
			}
		case "winc":
			if i+1 < len(tokens) {
				winc, _ = strconv.Atoi(tokens[i+1])
				i++
			}
		case "binc":
			if i+1 < len(tokens) {
				binc, _ = strconv.Atoi(tokens[i+1])
				i++
			}
		case "movestogo":
			if i+1 < len(tokens) {
				movestogo, _ = strconv.Atoi(tokens[i+1])
				i++
			}
		case "movetime":
			if i+1 < len(tokens) {
				movetime, _ = strconv.Atoi(tokens[i+1])
				i++
			}
		case "depth":
			if i+1 < len(tokens) {
				depth, _ = strconv.Atoi(tokens[i+1])
				i++
			}
		case "infinite":
			infinite = true
		case "perft":
			if i+1 < len(tokens) {
				d, _ := strconv.Atoi(tokens[i+1])
				PerftDivide(u.Board, d)
				return
			}
		}
	}

	if infinite {
		u.Search.Timer.SetInfinite()
	} else if movetime > 0 {
		u.Search.Timer.SetMoveTime(movetime)
	} else if depth > 0 {
		u.Search.Timer.SetFixedDepth(depth)
	} else {
		// Time control
		timeLeft := wtime
		inc := winc
		if u.Board.Side == Black {
			timeLeft = btime
			inc = binc
		}
		if timeLeft > 0 {
			u.Search.Timer.SetTimeControl(timeLeft, inc, movestogo)
		} else {
			u.Search.Timer.SetFixedDepth(8) // fallback
		}
	}

	// Run search in goroutine so we can handle "stop"
	go func() {
		bestMove := u.Search.IterativeDeepening()
		fmt.Printf("bestmove %s\n", bestMove.String())
	}()
}

func (u *UCI) handleStop() {
	u.Search.StopMu.Lock()
	u.Search.StopSignal = true
	u.Search.StopMu.Unlock()
	u.Search.Timer.Stop()
}

func (u *UCI) handlePerft(tokens []string) {
	depth := 5
	if len(tokens) > 1 {
		depth, _ = strconv.Atoi(tokens[1])
	}
	PerftDivide(u.Board, depth)
}

func (u *UCI) handleDisplay() {
	pieces := [2][6]byte{
		{'P', 'N', 'B', 'R', 'Q', 'K'},
		{'p', 'n', 'b', 'r', 'q', 'k'},
	}

	fmt.Println("  +---+---+---+---+---+---+---+---+")
	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%d |", rank+1)
		for file := 0; file < 8; file++ {
			sq := rank*8 + file
			c, p := u.Board.PieceOn(sq)
			if p >= 0 {
				fmt.Printf(" %c |", pieces[c][p])
			} else {
				fmt.Print("   |")
			}
		}
		fmt.Println()
		fmt.Println("  +---+---+---+---+---+---+---+---+")
	}
	fmt.Println("    a   b   c   d   e   f   g   h")
	fmt.Printf("FEN: %s\n", u.Board.FEN())
	fmt.Printf("Hash: %016x\n", u.Board.Hash)
	if u.Board.Side == White {
		fmt.Println("Side: White")
	} else {
		fmt.Println("Side: Black")
	}
}
