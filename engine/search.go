package engine

import (
	"fmt"
	"math"
	"sync"
)

const (
	MaxPly          = 128
	HistoryLimit    = 16384
	DefaultContempt = 12
)

// SearchInfo holds search state and results.
type SearchInfo struct {
	Board       *Board
	TT          *TransTable
	Timer       *TimeManager
	BestMove    Move
	RootSide    int
	Contempt    int
	Nodes       uint64
	KillerMoves [MaxPly][2]Move
	History     [2][64][64]int  // [side][from][to]
	CounterMove [2][64][64]Move // [side][from][to] -> counter move
	StaticEval  [MaxPly]int
	PV          [MaxPly][MaxPly]Move
	PVLength    [MaxPly]int

	StopMu     sync.Mutex
	StopSignal bool
}

// NewSearchInfo creates a new search info.
func NewSearchInfo(b *Board, tt *TransTable) *SearchInfo {
	return &SearchInfo{
		Board:    b,
		TT:       tt,
		Timer:    NewTimeManager(),
		RootSide: b.Side,
		Contempt: DefaultContempt,
	}
}

// ClearForSearch resets search state.
func (si *SearchInfo) ClearForSearch() {
	si.Nodes = 0
	si.BestMove = NullMove
	si.RootSide = si.Board.Side
	si.StopSignal = false
	for i := range si.KillerMoves {
		si.KillerMoves[i] = [2]Move{}
	}
	for i := range si.PVLength {
		si.PVLength[i] = 0
	}
	// Age history scores instead of clearing them completely.
	for s := 0; s < 2; s++ {
		for f := 0; f < 64; f++ {
			for t := 0; t < 64; t++ {
				si.History[s][f][t] /= 2
			}
		}
	}
}

func (si *SearchInfo) shouldStop() bool {
	si.StopMu.Lock()
	s := si.StopSignal
	si.StopMu.Unlock()
	if s {
		return true
	}
	return si.Timer.ShouldStop()
}

// IterativeDeepening performs iterative deepening with aspiration windows.
func (si *SearchInfo) IterativeDeepening() Move {
	si.ClearForSearch()
	si.TT.NewSearch()
	si.Timer.Start()

	var bestMove Move
	var bestScore int
	prevScore := 0
	prevBestMove := NullMove
	stableBestCount := 0

	for depth := 1; depth <= si.Timer.MaxDepth; depth++ {
		var score int

		if depth >= 5 {
			delta := 25
			alpha := prevScore - delta
			beta := prevScore + delta

			for {
				score = si.alphaBeta(alpha, beta, depth, 0, true)

				if si.shouldStop() && depth > 1 {
					goto done
				}

				if score <= alpha {
					alpha = max(score-delta, -Infinity)
					delta *= 2
				} else if score >= beta {
					beta = min(score+delta, Infinity)
					delta *= 2
				} else {
					break
				}

				if delta > 1000 {
					score = si.alphaBeta(-Infinity, Infinity, depth, 0, true)
					break
				}
			}
		} else {
			score = si.alphaBeta(-Infinity, Infinity, depth, 0, true)
		}

		if si.shouldStop() && depth > 1 {
			break
		}

		bestScore = score
		if si.PVLength[0] > 0 {
			bestMove = si.PV[0][0]
		}
		si.BestMove = bestMove

		elapsed := si.Timer.Elapsed()
		if elapsed == 0 {
			elapsed = 1
		}
		nps := si.Nodes * 1000 / uint64(elapsed)

		pvStr := ""
		for i := 0; i < si.PVLength[0]; i++ {
			pvStr += si.PV[0][i].String() + " "
		}

		scoreStr := ""
		if bestScore > MateThreshold {
			mateIn := (MateScore - bestScore + 1) / 2
			scoreStr = fmt.Sprintf("score mate %d", mateIn)
		} else if bestScore < -MateThreshold {
			mateIn := -(MateScore + bestScore + 1) / 2
			scoreStr = fmt.Sprintf("score mate %d", mateIn)
		} else {
			scoreStr = fmt.Sprintf("score cp %d", bestScore)
		}

		fmt.Printf("info depth %d %s nodes %d nps %d time %d hashfull %d pv %s\n",
			depth, scoreStr, si.Nodes, nps, elapsed, si.TT.Hashfull(), pvStr)

		if bestScore > MateThreshold || bestScore < -MateThreshold {
			break
		}

		scoreSwing := 0
		bestMoveChanged := prevBestMove != NullMove && bestMove != prevBestMove
		if prevBestMove != NullMove {
			scoreSwing = absInt(bestScore - prevScore)
		}
		if bestMoveChanged {
			stableBestCount = 0
		} else {
			stableBestCount++
		}
		if si.Timer.ShouldStopAfterIteration(bestMoveChanged, scoreSwing, stableBestCount, depth) {
			break
		}

		prevBestMove = bestMove
		prevScore = bestScore
	}

done:
	return bestMove
}

// alphaBeta is the main alpha-beta search with all enhancements.
func (si *SearchInfo) alphaBeta(alpha, beta, depth, ply int, doNull bool) int {
	return si.alphaBetaWithExclusion(alpha, beta, depth, ply, doNull, NullMove)
}

func (si *SearchInfo) alphaBetaWithExclusion(alpha, beta, depth, ply int, doNull bool, excluded Move) int {
	if si.Nodes&2047 == 0 {
		if si.shouldStop() {
			return 0
		}
	}

	si.PVLength[ply] = ply

	if si.Board.IsInsufficientMaterial() {
		return si.drawScore()
	}
	if ply > 0 && (si.Board.IsRepetition() || si.Board.HalfMove >= 100) {
		return si.drawScore()
	}

	inCheck := si.Board.InCheck()
	if inCheck && depth < MaxPly-1 {
		depth++
	}

	if depth <= 0 {
		return si.quiescence(alpha, beta, ply, 0)
	}

	si.Nodes++

	isPV := beta-alpha > 1

	var ttMove Move
	var ttEntry TTEntry
	var ttHit bool
	var ttScore int
	if excluded == NullMove {
		if entry, ok := si.TT.Probe(si.Board.Hash); ok {
			ttEntry = entry
			ttHit = true
			ttMove = entry.Move
			ttScore = ttToSearchScore(int(entry.Score), ply)
			if int(entry.Depth) >= depth && !isPV {
				switch entry.Flag {
				case TTExact:
					return ttScore
				case TTAlpha:
					if ttScore <= alpha {
						return alpha
					}
				case TTBeta:
					if ttScore >= beta {
						return beta
					}
				}
			}
		}
	}

	staticEval := Evaluate(si.Board)
	si.StaticEval[ply] = staticEval
	improving := ply < 2 || staticEval > si.StaticEval[ply-2]

	singularExtension := 0
	if excluded == NullMove && isPV && depth >= 6 && ttHit && ttMove != NullMove &&
		int(ttEntry.Depth) >= depth-3 && (ttEntry.Flag == TTExact || ttEntry.Flag == TTBeta) &&
		absInt(ttScore) < MateThreshold {
		margin := 24 + 2*depth
		singularBeta := ttScore - margin
		verifyDepth := max(1, depth/2)
		score := si.alphaBetaWithExclusion(singularBeta-1, singularBeta, verifyDepth, ply, false, ttMove)
		if !si.shouldStop() && score < singularBeta {
			singularExtension = 1
		}
	}

	if !isPV && !inCheck && depth <= 3 {
		razorMargin := 200 + 100*depth
		if staticEval+razorMargin <= alpha {
			if depth <= 1 {
				return si.quiescence(alpha, beta, ply, 0)
			}
			score := si.quiescence(alpha, beta, ply, 0)
			if score <= alpha {
				return score
			}
		}
	}

	if !isPV && !inCheck && depth <= 6 && staticEval-80*depth >= beta {
		return staticEval - 80*depth
	}

	if doNull && !inCheck && !isPV && depth >= 3 && !si.isEndgame() {
		r := 3 + depth/6
		if r > depth-1 {
			r = depth - 1
		}

		si.Board.MakeNullMove()
		score := -si.alphaBeta(-beta, -beta+1, depth-r-1, ply+1, false)
		si.Board.UnmakeNullMove()

		if si.shouldStop() {
			return 0
		}

		if score >= beta {
			if score > MateThreshold {
				score = beta
			}
			return score
		}
	}

	if excluded == NullMove && isPV && ttMove == NullMove && depth >= 4 {
		si.alphaBeta(alpha, beta, depth-2, ply, false)
		if entry, ok := si.TT.Probe(si.Board.Hash); ok {
			ttMove = entry.Move
		}
	}

	var ml MoveList
	si.Board.GenerateMoves(&ml)
	if ml.Count == 0 {
		if inCheck {
			return -MateScore + ply
		}
		return si.drawScore()
	}

	var scores [256]int
	var prevMove Move
	if si.Board.HistPly > 0 {
		prevMove = si.Board.History[si.Board.HistPly-1].Move
	}
	si.scoreMoves(&ml, scores[:], ttMove, ply, prevMove)

	bestScore := -Infinity
	bestMove := NullMove
	oldAlpha := alpha
	movesSearched := 0
	quietMovesSearched := 0
	var quietsTried [256]Move
	quietCount := 0

	for i := 0; i < ml.Count; i++ {
		pickBest(&ml, scores[:], i)
		m := ml.Moves[i]
		if excluded != NullMove && m == excluded {
			continue
		}

		isQuiet := !m.IsCapture() && !m.IsPromotion()

		if m != ttMove && !isPV && !inCheck && depth <= 6 && isQuiet && bestScore > -MateThreshold {
			if staticEval+si.futilityMargin(depth, improving) <= alpha {
				continue
			}
		}

		if m != ttMove && !isPV && !inCheck && depth <= 6 && isQuiet &&
			quietMovesSearched >= si.lmpLimit(depth, improving) {
			continue
		}

		if !isPV && depth <= 4 && m.IsCapture() && movesSearched > 0 && !inCheck {
			if si.Board.SEE(m) < -50*depth {
				continue
			}
		}

		if !si.Board.MakeMove(m) {
			continue
		}
		movesSearched++
		if isQuiet {
			quietsTried[quietCount] = m
			quietCount++
			quietMovesSearched++
		}

		var score int
		givesCheck := si.Board.InCheck()
		childDepth := depth - 1 + singularExtensionForMove(m, ttMove, singularExtension)

		if movesSearched == 1 {
			score = -si.alphaBeta(-beta, -alpha, childDepth, ply+1, true)
		} else {
			reduction := 0
			if depth >= 3 && movesSearched > 3 && !inCheck && isQuiet {
				reduction = int(math.Log(float64(depth)) * math.Log(float64(movesSearched)) / 2.5)
				if reduction < 1 {
					reduction = 1
				}
				if ply < MaxPly && (m == si.KillerMoves[ply][0] || m == si.KillerMoves[ply][1]) {
					reduction--
				}
				if givesCheck {
					reduction--
				}
				if reduction < 0 {
					reduction = 0
				}
				if childDepth-reduction < 1 {
					reduction = childDepth - 1
				}
				if reduction < 0 {
					reduction = 0
				}
			}

			score = -si.alphaBeta(-alpha-1, -alpha, childDepth-reduction, ply+1, true)
			if score > alpha && reduction > 0 {
				score = -si.alphaBeta(-alpha-1, -alpha, childDepth, ply+1, true)
			}
			if score > alpha && score < beta {
				score = -si.alphaBeta(-beta, -alpha, childDepth, ply+1, true)
			}
		}

		si.Board.UnmakeMove()

		if si.shouldStop() {
			return 0
		}

		if score > bestScore {
			bestScore = score
			bestMove = m

			if score > alpha {
				alpha = score

				si.PV[ply][ply] = m
				for j := ply + 1; j < si.PVLength[ply+1]; j++ {
					si.PV[ply][j] = si.PV[ply+1][j]
				}
				si.PVLength[ply] = si.PVLength[ply+1]

				if score >= beta {
					if isQuiet {
						si.recordQuietBetaCutoff(ply, depth, si.Board.Side, m, prevMove, quietsTried[:quietCount])
					}
					break
				}
			}
		}
	}

	if movesSearched == 0 {
		if excluded != NullMove {
			if inCheck {
				return -MateScore + ply
			}
			return si.drawScore()
		}
		return alpha
	}

	flag := TTExact
	if bestScore <= oldAlpha {
		flag = TTAlpha
	} else if bestScore >= beta {
		flag = TTBeta
	}

	ttStoreScore := bestScore
	if ttStoreScore > MateThreshold {
		ttStoreScore += ply
	} else if ttStoreScore < -MateThreshold {
		ttStoreScore -= ply
	}
	if excluded == NullMove {
		si.TT.Store(si.Board.Hash, depth, ttStoreScore, flag, bestMove)
	}

	return bestScore
}

// quiescence search searches captures and checks at depth 0.
func (si *SearchInfo) quiescence(alpha, beta, ply, qsDepth int) int {
	si.Nodes++

	if si.Nodes&2047 == 0 {
		if si.shouldStop() {
			return 0
		}
	}

	inCheck := si.Board.InCheck()
	standPat := EvaluateFast(si.Board)

	if !inCheck {
		if standPat >= beta {
			return beta
		}

		bigDelta := PieceValues[Queen] + 200
		if standPat+bigDelta < alpha {
			return alpha
		}

		if standPat > alpha {
			alpha = standPat
		}
	} else {
		standPat = -Infinity
	}

	var ml MoveList
	if inCheck {
		si.Board.GenerateMoves(&ml)
	} else {
		si.Board.GenerateCaptures(&ml)
	}

	if inCheck && ml.Count == 0 {
		return -MateScore + ply
	}

	var scores [256]int
	si.scoreMoves(&ml, scores[:], NullMove, ply, NullMove)

	for i := 0; i < ml.Count; i++ {
		pickBest(&ml, scores[:], i)
		m := ml.Moves[i]

		if !inCheck {
			if !m.IsPromotion() {
				captVal := 0
				if m.CapturedPiece() < 6 {
					captVal = PieceValues[m.CapturedPiece()]
				}
				if standPat+captVal+200 < alpha {
					continue
				}
			}

			if m.IsCapture() && si.Board.SEE(m) < 0 {
				continue
			}
		}

		if !si.Board.MakeMove(m) {
			continue
		}

		score := -si.quiescence(-beta, -alpha, ply+1, qsDepth+1)
		si.Board.UnmakeMove()

		if si.shouldStop() {
			return 0
		}

		if score > alpha {
			alpha = score
			if score >= beta {
				return beta
			}
		}
	}

	return alpha
}

// scoreMoves scores moves for move ordering.
func (si *SearchInfo) scoreMoves(ml *MoveList, scores []int, ttMove Move, ply int, prevMove Move) {
	for i := 0; i < ml.Count; i++ {
		m := ml.Moves[i]
		if m == ttMove {
			scores[i] = 10000000
		} else if m.IsCapture() {
			seeVal := si.Board.SEE(m)
			if seeVal >= 0 {
				if m.Flags() == FlagEnPassant {
					scores[i] = 1050000
				} else {
					scores[i] = 1000000 + MvvLva[m.CapturedPiece()][m.MovedPiece()]
				}
			} else {
				scores[i] = -100000 + MvvLva[m.CapturedPiece()][m.MovedPiece()]
			}
		} else if m.IsPromotion() {
			scores[i] = 950000
		} else if ply < MaxPly && m == si.KillerMoves[ply][0] {
			scores[i] = 900000
		} else if ply < MaxPly && m == si.KillerMoves[ply][1] {
			scores[i] = 800000
		} else if prevMove != NullMove {
			side := si.Board.Side
			cm := si.CounterMove[side][prevMove.From()][prevMove.To()]
			if m == cm {
				scores[i] = 700000
			} else {
				scores[i] = si.History[si.Board.Side][m.From()][m.To()]
			}
		} else {
			scores[i] = si.History[si.Board.Side][m.From()][m.To()]
		}
	}
}

func (si *SearchInfo) futilityMargin(depth int, improving bool) int {
	margin := 110 + 125*depth + 12*depth*depth
	if improving {
		margin += 40
	}
	return margin
}

func (si *SearchInfo) lmpLimit(depth int, improving bool) int {
	limits := [7]int{0, 0, 4, 8, 13, 19, 28}
	if depth < len(limits) {
		limit := limits[depth]
		if improving {
			limit += 2
		}
		return limit
	}
	limit := 28 + depth*depth/2
	if improving {
		limit += 2
	}
	return limit
}

func (si *SearchInfo) recordQuietBetaCutoff(ply, depth, side int, move Move, prevMove Move, quiets []Move) {
	if si.KillerMoves[ply][0] != move {
		si.KillerMoves[ply][1] = si.KillerMoves[ply][0]
		si.KillerMoves[ply][0] = move
	}

	bonus := min(HistoryLimit/2, 32*depth*depth)
	si.updateHistory(side, move, bonus)
	for _, quiet := range quiets {
		if quiet != move {
			si.updateHistory(side, quiet, -bonus)
		}
	}

	if prevMove != NullMove {
		si.CounterMove[side][prevMove.From()][prevMove.To()] = move
	}
}

func (si *SearchInfo) updateHistory(side int, move Move, bonus int) {
	entry := &si.History[side][move.From()][move.To()]
	*entry += bonus - (*entry*absInt(bonus))/HistoryLimit
	if *entry > HistoryLimit {
		*entry = HistoryLimit
	}
	if *entry < -HistoryLimit {
		*entry = -HistoryLimit
	}
}

func ttToSearchScore(score, ply int) int {
	if score > MateThreshold {
		return score - ply
	}
	if score < -MateThreshold {
		return score + ply
	}
	return score
}

func singularExtensionForMove(move, ttMove Move, extension int) int {
	if extension > 0 && move == ttMove {
		return extension
	}
	return 0
}

func (si *SearchInfo) drawScore() int {
	if si.Contempt == 0 {
		return DrawScore
	}
	if si.Board.Side == si.RootSide {
		return -si.Contempt
	}
	return si.Contempt
}

func pickBest(ml *MoveList, scores []int, start int) {
	bestIdx := start
	bestScore := scores[start]
	for i := start + 1; i < ml.Count; i++ {
		if scores[i] > bestScore {
			bestScore = scores[i]
			bestIdx = i
		}
	}
	if bestIdx != start {
		ml.Moves[start], ml.Moves[bestIdx] = ml.Moves[bestIdx], ml.Moves[start]
		scores[start], scores[bestIdx] = scores[bestIdx], scores[start]
	}
}

func (si *SearchInfo) isEndgame() bool {
	return si.Board.Pieces[si.Board.Side][Queen] == 0
}
