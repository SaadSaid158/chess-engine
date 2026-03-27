package engine

// Evaluation constants
const (
	MateScore     = 30000
	MateThreshold = MateScore - 1000
	DrawScore     = 0
	Infinity      = 31000
)

// Material values (centipawns) — well-tuned
var PieceValues = [6]int{100, 325, 335, 500, 975, 20000}

// SEE piece values (simpler, for exchange evaluation)
var SEEValues = [7]int{100, 325, 335, 500, 975, 20000, 0}

// MVV-LVA table: [victim][attacker]
var MvvLva [6][6]int

func InitEval() {
	for victim := 0; victim < 6; victim++ {
		for attacker := 0; attacker < 6; attacker++ {
			MvvLva[victim][attacker] = PieceValues[victim]*10 - PieceValues[attacker]
		}
	}
}

// -------------------------------------------------------------------
// Piece-square tables — separate middlegame and endgame
// Index 0 = a1, from White's perspective
// -------------------------------------------------------------------

var PawnMgPST = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	-1, 4, 2, -10, -10, 2, 4, -1,
	-1, -4, -6, 4, 4, -6, -4, -1,
	0, 0, 0, 20, 20, 0, 0, 0,
	4, 4, 8, 24, 24, 8, 4, 4,
	8, 8, 16, 28, 28, 16, 8, 8,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}
var PawnEgPST = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	-8, -4, 0, 0, 0, 0, -4, -8,
	-6, -2, 4, 6, 6, 4, -2, -6,
	0, 4, 10, 14, 14, 10, 4, 0,
	8, 12, 18, 24, 24, 18, 12, 8,
	20, 24, 30, 36, 36, 30, 24, 20,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var KnightMgPST = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}
var KnightEgPST = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, -5, 0, 0, -5, -20, -40,
	-30, -5, 10, 15, 15, 10, -5, -30,
	-30, 0, 15, 25, 25, 15, 0, -30,
	-30, 0, 15, 25, 25, 15, 0, -30,
	-30, -5, 10, 15, 15, 10, -5, -30,
	-40, -20, -5, 0, 0, -5, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var BishopMgPST = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}
var BishopEgPST = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 0, 10, 15, 15, 10, 0, -10,
	-10, 0, 10, 15, 15, 10, 0, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var RookMgPST = [64]int{
	0, 0, 0, 5, 5, 0, 0, 0,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}
var RookEgPST = [64]int{
	0, 0, 0, 5, 5, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 5, 5, 5, 5, 5, 5, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var QueenMgPST = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}
var QueenEgPST = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 10, 10, 5, 0, -5,
	-5, 0, 5, 10, 10, 5, 0, -5,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var KingMgPST = [64]int{
	30, 40, 20, 0, 0, 10, 40, 30,
	20, 20, 0, -5, -5, 0, 20, 20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
}
var KingEgPST = [64]int{
	-50, -30, -30, -30, -30, -30, -30, -50,
	-30, -20, 0, 0, 0, 0, -20, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-50, -40, -30, -20, -20, -30, -40, -50,
}

// Passed pawn bonus by rank (from White's perspective)
var PassedPawnBonus = [8]int{0, 10, 15, 25, 45, 75, 120, 0}

// mirror square for Black
func mirrorSquare(sq int) int {
	return sq ^ 56
}

// -------------------------------------------------------------------
// Pawn hash table — caches pawn structure evaluation
// -------------------------------------------------------------------
const PawnHashSize = 1 << 16 // 64K entries

type PawnHashEntry struct {
	key uint64
	mg  int16
	eg  int16
}

var PawnHashTable [PawnHashSize]PawnHashEntry

func probePawnHash(key uint64) (int, int, bool) {
	idx := key & (PawnHashSize - 1)
	e := &PawnHashTable[idx]
	if e.key == key {
		return int(e.mg), int(e.eg), true
	}
	return 0, 0, false
}

func storePawnHash(key uint64, mg, eg int) {
	idx := key & (PawnHashSize - 1)
	PawnHashTable[idx] = PawnHashEntry{key: key, mg: int16(mg), eg: int16(eg)}
}

// -------------------------------------------------------------------
// Evaluate — full evaluation from side to move's perspective
// -------------------------------------------------------------------
func Evaluate(b *Board) int {
	score := evaluate(b)
	if b.Side == Black {
		score = -score
	}
	return score
}

// EvaluateFast — material + PST only, for quiescence / lazy eval cutoffs
func EvaluateFast(b *Board) int {
	score := evaluateFast(b)
	if b.Side == Black {
		score = -score
	}
	return score
}

// Phase weights
const (
	KnightPhase = 1
	BishopPhase = 1
	RookPhase   = 2
	QueenPhase  = 4
	TotalPhase  = 4*KnightPhase + 4*BishopPhase + 4*RookPhase + 2*QueenPhase // 24
)

// LazyMargin — if fast eval is this far from the window, skip slow eval terms
const LazyMargin = 400

// evaluate returns score from White's perspective
func evaluate(b *Board) int {
	var mgScore, egScore int
	var whiteMaterial, blackMaterial int
	phase := 0

	// ------ Material + PST ------
	for piece := Pawn; piece <= King; piece++ {
		bb := b.Pieces[White][piece]
		for bb != 0 {
			sq := PopLSB(&bb)
			whiteMaterial += PieceValues[piece]
			mgScore += pstMg(piece, sq)
			egScore += pstEg(piece, sq)
			switch piece {
			case Knight:
				phase += KnightPhase
			case Bishop:
				phase += BishopPhase
			case Rook:
				phase += RookPhase
			case Queen:
				phase += QueenPhase
			}
		}
		bb = b.Pieces[Black][piece]
		for bb != 0 {
			sq := PopLSB(&bb)
			blackMaterial += PieceValues[piece]
			msq := mirrorSquare(sq)
			mgScore -= pstMg(piece, msq)
			egScore -= pstEg(piece, msq)
			switch piece {
			case Knight:
				phase += KnightPhase
			case Bishop:
				phase += BishopPhase
			case Rook:
				phase += RookPhase
			case Queen:
				phase += QueenPhase
			}
		}
	}

	mgScore += whiteMaterial - blackMaterial
	egScore += whiteMaterial - blackMaterial

	if phase > TotalPhase {
		phase = TotalPhase
	}

	// ------ Lazy evaluation cutoff ------
	// If material+PST is far outside the likely search window, skip expensive terms
	fastScore := (mgScore*phase + egScore*(TotalPhase-phase)) / TotalPhase
	// We don't have alpha/beta here, so we use a simpler heuristic:
	// If the position is very lopsided, the slow terms won't change the outcome
	// This is checked by the caller via EvaluateFast comparison

	// ------ Bishop pair ------
	if PopCount(b.Pieces[White][Bishop]) >= 2 {
		mgScore += 35
		egScore += 55
	}
	if PopCount(b.Pieces[Black][Bishop]) >= 2 {
		mgScore -= 35
		egScore -= 55
	}

	// ------ Rook on open / semi-open files + 7th rank ------
	whitePawns := b.Pieces[White][Pawn]
	blackPawns := b.Pieces[Black][Pawn]

	bb := b.Pieces[White][Rook]
	for bb != 0 {
		sq := PopLSB(&bb)
		f := sq % 8
		if whitePawns&FileMask[f] == 0 {
			if blackPawns&FileMask[f] == 0 {
				mgScore += 22
				egScore += 18
			} else {
				mgScore += 12
				egScore += 8
			}
		}
		if sq/8 == 6 {
			mgScore += 15
			egScore += 25
		}
	}
	bb = b.Pieces[Black][Rook]
	for bb != 0 {
		sq := PopLSB(&bb)
		f := sq % 8
		if blackPawns&FileMask[f] == 0 {
			if whitePawns&FileMask[f] == 0 {
				mgScore -= 22
				egScore -= 18
			} else {
				mgScore -= 12
				egScore -= 8
			}
		}
		if sq/8 == 1 {
			mgScore -= 15
			egScore -= 25
		}
	}

	// ------ Pawn structure (cached) ------
	pawnKey := b.PawnHash
	mgPawn, egPawn, hit := probePawnHash(pawnKey)
	if !hit {
		mgPawn, egPawn = evalPawnStructure(b, whitePawns, blackPawns)
		storePawnHash(pawnKey, mgPawn, egPawn)
	}
	mgScore += mgPawn
	egScore += egPawn

	// ------ Mobility ------
	mgMob, egMob := evalMobility(b)
	mgScore += mgMob
	egScore += egMob

	// ------ King safety (middlegame only) ------
	mgKS := evalKingSafety(b, whitePawns, blackPawns)
	mgScore += mgKS

	// ------ Tempo ------
	if b.Side == White {
		mgScore += 14
	} else {
		mgScore -= 14
	}

	// ------ Tapered eval ------
	_ = fastScore // used for lazy eval by caller
	score := (mgScore*phase + egScore*(TotalPhase-phase)) / TotalPhase
	return score
}

// evaluateFast returns material + PST score only (cheap, for quiescence / lazy cutoffs)
func evaluateFast(b *Board) int {
	var mgScore, egScore int
	var whiteMaterial, blackMaterial int
	phase := 0

	for piece := Pawn; piece <= King; piece++ {
		bb := b.Pieces[White][piece]
		for bb != 0 {
			sq := PopLSB(&bb)
			whiteMaterial += PieceValues[piece]
			mgScore += pstMg(piece, sq)
			egScore += pstEg(piece, sq)
			switch piece {
			case Knight:
				phase += KnightPhase
			case Bishop:
				phase += BishopPhase
			case Rook:
				phase += RookPhase
			case Queen:
				phase += QueenPhase
			}
		}
		bb = b.Pieces[Black][piece]
		for bb != 0 {
			sq := PopLSB(&bb)
			blackMaterial += PieceValues[piece]
			msq := mirrorSquare(sq)
			mgScore -= pstMg(piece, msq)
			egScore -= pstEg(piece, msq)
			switch piece {
			case Knight:
				phase += KnightPhase
			case Bishop:
				phase += BishopPhase
			case Rook:
				phase += RookPhase
			case Queen:
				phase += QueenPhase
			}
		}
	}

	mgScore += whiteMaterial - blackMaterial
	egScore += whiteMaterial - blackMaterial

	if phase > TotalPhase {
		phase = TotalPhase
	}

	score := (mgScore*phase + egScore*(TotalPhase-phase)) / TotalPhase
	return score
}

func pstMg(piece, sq int) int {
	switch piece {
	case Pawn:
		return PawnMgPST[sq]
	case Knight:
		return KnightMgPST[sq]
	case Bishop:
		return BishopMgPST[sq]
	case Rook:
		return RookMgPST[sq]
	case Queen:
		return QueenMgPST[sq]
	case King:
		return KingMgPST[sq]
	}
	return 0
}

func pstEg(piece, sq int) int {
	switch piece {
	case Pawn:
		return PawnEgPST[sq]
	case Knight:
		return KnightEgPST[sq]
	case Bishop:
		return BishopEgPST[sq]
	case Rook:
		return RookEgPST[sq]
	case Queen:
		return QueenEgPST[sq]
	case King:
		return KingEgPST[sq]
	}
	return 0
}

// -------------------------------------------------------------------
// Pawn structure
// -------------------------------------------------------------------
func evalPawnStructure(b *Board, whitePawns, blackPawns uint64) (int, int) {
	var mg, eg int

	// White pawns
	bb := whitePawns
	for bb != 0 {
		sq := PopLSB(&bb)
		file := sq % 8
		rank := sq / 8

		// Doubled (only penalize once, when another pawn is ahead)
		above := whitePawns & FileMask[file] & ^(uint64(1) << uint(sq))
		if above != 0 {
			hasAhead := false
			tmp := above
			for tmp != 0 {
				s := PopLSB(&tmp)
				if s/8 > rank {
					hasAhead = true
					break
				}
			}
			if hasAhead {
				mg -= 8
				eg -= 16
			}
		}

		// Isolated
		isolated := true
		if file > 0 && whitePawns&FileMask[file-1] != 0 {
			isolated = false
		}
		if file < 7 && whitePawns&FileMask[file+1] != 0 {
			isolated = false
		}
		if isolated {
			mg -= 12
			eg -= 18
		}

		// Backward pawn
		if !isolated {
			backward := true
			for r := rank; r >= 0; r-- {
				mask := uint64(0)
				if file > 0 {
					mask |= 1 << uint(r*8+file-1)
				}
				if file < 7 {
					mask |= 1 << uint(r*8+file+1)
				}
				if whitePawns&mask != 0 {
					backward = false
					break
				}
			}
			if backward && rank+1 < 8 {
				stopSq := (rank+1)*8 + file
				if PawnAttacks[White][stopSq]&blackPawns == 0 {
					backward = false
				}
			}
			if backward {
				mg -= 8
				eg -= 10
			}
		}

		// Passed pawn
		passed := true
		for r := rank + 1; r < 8; r++ {
			blockMask := uint64(1) << uint(r*8+file)
			if file > 0 {
				blockMask |= 1 << uint(r*8+file-1)
			}
			if file < 7 {
				blockMask |= 1 << uint(r*8+file+1)
			}
			if blackPawns&blockMask != 0 {
				passed = false
				break
			}
		}
		if passed {
			mg += PassedPawnBonus[rank]
			eg += PassedPawnBonus[rank] * 2
			// Bonus if path to promotion is clear
			blocked := false
			for r := rank + 1; r < 8; r++ {
				if b.AllOccupied&(1<<uint(r*8+file)) != 0 {
					blocked = true
					break
				}
			}
			if !blocked {
				eg += PassedPawnBonus[rank]
			}
			// King proximity to passed pawn (endgame)
			bKingSq := LSB(b.Pieces[Black][King])
			promoSq := 7*8 + file
			dist := chebyshevDist(bKingSq, promoSq)
			eg += dist * 5 // enemy king far = good
			wKingSq := LSB(b.Pieces[White][King])
			dist2 := chebyshevDist(wKingSq, promoSq)
			eg -= dist2 * 3 // own king close = good
		}

		// Connected pawns
		if file > 0 && rank > 0 && whitePawns&(1<<uint((rank-1)*8+file-1)) != 0 {
			mg += 5
			eg += 5
		}
		if file < 7 && rank > 0 && whitePawns&(1<<uint((rank-1)*8+file+1)) != 0 {
			mg += 5
			eg += 5
		}
	}

	// Black pawns
	bb = blackPawns
	for bb != 0 {
		sq := PopLSB(&bb)
		file := sq % 8
		rank := sq / 8

		below := blackPawns & FileMask[file] & ^(uint64(1) << uint(sq))
		if below != 0 {
			hasBelow := false
			tmp := below
			for tmp != 0 {
				s := PopLSB(&tmp)
				if s/8 < rank {
					hasBelow = true
					break
				}
			}
			if hasBelow {
				mg += 8
				eg += 16
			}
		}

		isolated := true
		if file > 0 && blackPawns&FileMask[file-1] != 0 {
			isolated = false
		}
		if file < 7 && blackPawns&FileMask[file+1] != 0 {
			isolated = false
		}
		if isolated {
			mg += 12
			eg += 18
		}

		if !isolated {
			backward := true
			for r := rank; r < 8; r++ {
				mask := uint64(0)
				if file > 0 {
					mask |= 1 << uint(r*8+file-1)
				}
				if file < 7 {
					mask |= 1 << uint(r*8+file+1)
				}
				if blackPawns&mask != 0 {
					backward = false
					break
				}
			}
			if backward && rank-1 >= 0 {
				stopSq := (rank-1)*8 + file
				if PawnAttacks[Black][stopSq]&whitePawns == 0 {
					backward = false
				}
			}
			if backward {
				mg += 8
				eg += 10
			}
		}

		passed := true
		for r := rank - 1; r >= 0; r-- {
			blockMask := uint64(1) << uint(r*8+file)
			if file > 0 {
				blockMask |= 1 << uint(r*8+file-1)
			}
			if file < 7 {
				blockMask |= 1 << uint(r*8+file+1)
			}
			if whitePawns&blockMask != 0 {
				passed = false
				break
			}
		}
		if passed {
			bRank := 7 - rank
			mg -= PassedPawnBonus[bRank]
			eg -= PassedPawnBonus[bRank] * 2
			blocked := false
			for r := rank - 1; r >= 0; r-- {
				if b.AllOccupied&(1<<uint(r*8+file)) != 0 {
					blocked = true
					break
				}
			}
			if !blocked {
				eg -= PassedPawnBonus[bRank]
			}
			wKingSq := LSB(b.Pieces[White][King])
			promoSq := file // rank 0
			dist := chebyshevDist(wKingSq, promoSq)
			eg -= dist * 5
			bKingSq := LSB(b.Pieces[Black][King])
			dist2 := chebyshevDist(bKingSq, promoSq)
			eg += dist2 * 3
		}

		if file > 0 && rank < 7 && blackPawns&(1<<uint((rank+1)*8+file-1)) != 0 {
			mg -= 5
			eg -= 5
		}
		if file < 7 && rank < 7 && blackPawns&(1<<uint((rank+1)*8+file+1)) != 0 {
			mg -= 5
			eg -= 5
		}
	}

	return mg, eg
}

// -------------------------------------------------------------------
// Mobility — safe squares (not attacked by enemy pawns)
// -------------------------------------------------------------------
func evalMobility(b *Board) (int, int) {
	var mg, eg int
	occ := b.AllOccupied
	wPawnAtt := pawnAttackSpan(b.Pieces[White][Pawn], White)
	bPawnAtt := pawnAttackSpan(b.Pieces[Black][Pawn], Black)

	// White
	bb := b.Pieces[White][Knight]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(KnightAttacks[sq] & ^b.Occupied[White] & ^bPawnAtt)
		mg += (mob - 4) * 5
		eg += (mob - 4) * 5
	}
	bb = b.Pieces[White][Bishop]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(BishopAttacks(sq, occ) & ^b.Occupied[White] & ^bPawnAtt)
		mg += (mob - 6) * 4
		eg += (mob - 6) * 4
	}
	bb = b.Pieces[White][Rook]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(RookAttacks(sq, occ) & ^b.Occupied[White])
		mg += (mob - 7) * 2
		eg += (mob - 7) * 4
	}
	bb = b.Pieces[White][Queen]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(QueenAttacks(sq, occ) & ^b.Occupied[White] & ^bPawnAtt)
		mg += (mob - 14) * 1
		eg += (mob - 14) * 2
	}

	// Black
	bb = b.Pieces[Black][Knight]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(KnightAttacks[sq] & ^b.Occupied[Black] & ^wPawnAtt)
		mg -= (mob - 4) * 5
		eg -= (mob - 4) * 5
	}
	bb = b.Pieces[Black][Bishop]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(BishopAttacks(sq, occ) & ^b.Occupied[Black] & ^wPawnAtt)
		mg -= (mob - 6) * 4
		eg -= (mob - 6) * 4
	}
	bb = b.Pieces[Black][Rook]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(RookAttacks(sq, occ) & ^b.Occupied[Black])
		mg -= (mob - 7) * 2
		eg -= (mob - 7) * 4
	}
	bb = b.Pieces[Black][Queen]
	for bb != 0 {
		sq := PopLSB(&bb)
		mob := PopCount(QueenAttacks(sq, occ) & ^b.Occupied[Black] & ^wPawnAtt)
		mg -= (mob - 14) * 1
		eg -= (mob - 14) * 2
	}

	return mg, eg
}

func pawnAttackSpan(pawns uint64, side int) uint64 {
	if side == White {
		return ((pawns << 7) & ^FileMask[7]) | ((pawns << 9) & ^FileMask[0])
	}
	return ((pawns >> 9) & ^FileMask[7]) | ((pawns >> 7) & ^FileMask[0])
}

// -------------------------------------------------------------------
// King safety — attack-count-based with safety table
// -------------------------------------------------------------------

var SafetyTable = [100]int{
	0, 0, 1, 2, 3, 5, 7, 9, 12, 15,
	18, 22, 26, 30, 35, 40, 45, 51, 57, 63,
	70, 77, 84, 92, 100, 108, 116, 125, 134, 143,
	152, 162, 172, 182, 192, 203, 214, 225, 236, 248,
	260, 272, 284, 297, 310, 323, 336, 350, 364, 378,
	392, 407, 422, 437, 452, 468, 484, 500, 516, 533,
	550, 567, 584, 602, 620, 638, 656, 675, 694, 713,
	732, 752, 772, 792, 812, 833, 854, 875, 896, 918,
	940, 940, 940, 940, 940, 940, 940, 940, 940, 940,
	940, 940, 940, 940, 940, 940, 940, 940, 940, 940,
}

func evalKingSafety(b *Board, whitePawns, blackPawns uint64) int {
	score := 0
	occ := b.AllOccupied

	// White king safety (attacks by black pieces)
	if b.Pieces[Black][Queen] != 0 {
		wKingSq := LSB(b.Pieces[White][King])
		score -= kingDanger(b, wKingSq, White, Black, whitePawns, occ)
	}

	// Black king safety (attacks by white pieces)
	if b.Pieces[White][Queen] != 0 {
		bKingSq := LSB(b.Pieces[Black][King])
		score += kingDanger(b, bKingSq, Black, White, blackPawns, occ)
	}

	return score
}

func kingDanger(b *Board, kingSq, kingSide, attackSide int, friendlyPawns uint64, occ uint64) int {
	// King zone: king square + surrounding squares + forward extension
	kingZone := KingAttacks[kingSq] | (uint64(1) << uint(kingSq))
	if kingSide == White {
		kingZone |= kingZone << 8
	} else {
		kingZone |= kingZone >> 8
	}

	attackCount := 0
	attackWeight := 0

	// Knights attacking king zone
	bb := b.Pieces[attackSide][Knight]
	for bb != 0 {
		sq := PopLSB(&bb)
		if KnightAttacks[sq]&kingZone != 0 {
			attackCount++
			attackWeight += 2
		}
	}

	// Bishops attacking king zone
	bb = b.Pieces[attackSide][Bishop]
	for bb != 0 {
		sq := PopLSB(&bb)
		if BishopAttacks(sq, occ)&kingZone != 0 {
			attackCount++
			attackWeight += 2
		}
	}

	// Rooks attacking king zone
	bb = b.Pieces[attackSide][Rook]
	for bb != 0 {
		sq := PopLSB(&bb)
		if RookAttacks(sq, occ)&kingZone != 0 {
			attackCount++
			attackWeight += 3
		}
	}

	// Queens attacking king zone
	bb = b.Pieces[attackSide][Queen]
	for bb != 0 {
		sq := PopLSB(&bb)
		if QueenAttacks(sq, occ)&kingZone != 0 {
			attackCount++
			attackWeight += 5
		}
	}

	danger := 0
	if attackCount >= 2 {
		idx := attackWeight
		if idx >= 100 {
			idx = 99
		}
		danger = SafetyTable[idx] / 10 // Scale down to centipawns
	}

	// Pawn shield evaluation
	kingFile := kingSq % 8
	kingRank := kingSq / 8
	shieldBonus := 0
	isCastled := false
	if kingSide == White {
		isCastled = kingRank <= 1 && (kingFile <= 2 || kingFile >= 5)
	} else {
		isCastled = kingRank >= 6 && (kingFile <= 2 || kingFile >= 5)
	}

	if isCastled {
		for f := max(0, kingFile-1); f <= min(7, kingFile+1); f++ {
			if kingSide == White {
				if friendlyPawns&(1<<uint((kingRank+1)*8+f)) != 0 {
					shieldBonus += 12
				} else if kingRank+2 < 8 && friendlyPawns&(1<<uint((kingRank+2)*8+f)) != 0 {
					shieldBonus += 5
				} else {
					shieldBonus -= 15
				}
			} else {
				if friendlyPawns&(1<<uint((kingRank-1)*8+f)) != 0 {
					shieldBonus += 12
				} else if kingRank-2 >= 0 && friendlyPawns&(1<<uint((kingRank-2)*8+f)) != 0 {
					shieldBonus += 5
				} else {
					shieldBonus -= 15
				}
			}
		}
	}

	// Open files near king penalty
	for f := max(0, kingFile-1); f <= min(7, kingFile+1); f++ {
		if friendlyPawns&FileMask[f] == 0 {
			danger += 15
		}
	}

	return danger - shieldBonus
}

// -------------------------------------------------------------------
// Real Static Exchange Evaluation (iterative, with x-ray)
// -------------------------------------------------------------------
func (b *Board) SEE(m Move) int {
	if !m.IsCapture() {
		return 0
	}

	to := m.To()
	from := m.From()

	capturedVal := 0
	if m.Flags() == FlagEnPassant {
		capturedVal = SEEValues[Pawn]
	} else {
		cp := m.CapturedPiece()
		if cp >= 6 {
			return 0
		}
		capturedVal = SEEValues[cp]
	}

	// Build gain array for the exchange sequence
	var gain [32]int
	depth := 0
	gain[0] = capturedVal

	movedPiece := m.MovedPiece()
	movedVal := SEEValues[movedPiece]
	side := b.Side ^ 1 // opponent to respond

	// X-ray occupancy tracking
	occ := b.AllOccupied
	occ ^= 1 << uint(from)
	if m.Flags() == FlagEnPassant {
		epCapSq := to - 8
		if b.Side == Black {
			epCapSq = to + 8
		}
		occ ^= 1 << uint(epCapSq)
	}

	attackers := b.allAttackers(to, occ) & occ

	for attackers != 0 {
		depth++
		gain[depth] = movedVal - gain[depth-1]

		// Pruning: if score can't improve, stop
		if max(-gain[depth-1], gain[depth]) < 0 {
			break
		}

		// Find least valuable attacker for current side
		piece, sq := b.leastValuableAttacker(attackers, side, occ)
		if piece < 0 {
			break
		}

		occ ^= 1 << uint(sq) // remove this attacker
		movedVal = SEEValues[piece]
		side ^= 1

		// Add x-ray discovered attackers through the removed piece
		if piece == Pawn || piece == Bishop || piece == Queen {
			attackers |= BishopAttacks(to, occ) & (b.Pieces[White][Bishop] | b.Pieces[White][Queen] |
				b.Pieces[Black][Bishop] | b.Pieces[Black][Queen])
		}
		if piece == Pawn || piece == Rook || piece == Queen {
			attackers |= RookAttacks(to, occ) & (b.Pieces[White][Rook] | b.Pieces[White][Queen] |
				b.Pieces[Black][Rook] | b.Pieces[Black][Queen])
		}
		attackers &= occ
	}

	// Negamax the gain array
	for depth > 0 {
		gain[depth-1] = -max(-gain[depth-1], gain[depth])
		depth--
	}

	return gain[0]
}

func (b *Board) allAttackers(sq int, occ uint64) uint64 {
	return (PawnAttacks[White][sq] & b.Pieces[Black][Pawn]) |
		(PawnAttacks[Black][sq] & b.Pieces[White][Pawn]) |
		(KnightAttacks[sq] & (b.Pieces[White][Knight] | b.Pieces[Black][Knight])) |
		(BishopAttacks(sq, occ) & (b.Pieces[White][Bishop] | b.Pieces[White][Queen] |
			b.Pieces[Black][Bishop] | b.Pieces[Black][Queen])) |
		(RookAttacks(sq, occ) & (b.Pieces[White][Rook] | b.Pieces[White][Queen] |
			b.Pieces[Black][Rook] | b.Pieces[Black][Queen])) |
		(KingAttacks[sq] & (b.Pieces[White][King] | b.Pieces[Black][King]))
}

func (b *Board) leastValuableAttacker(attackers uint64, side int, occ uint64) (int, int) {
	for piece := Pawn; piece <= King; piece++ {
		subset := attackers & b.Pieces[side][piece] & occ
		if subset != 0 {
			sq := LSB(subset)
			return piece, sq
		}
	}
	return -1, -1
}

// Utility
func chebyshevDist(sq1, sq2 int) int {
	dr := absInt(sq1/8 - sq2/8)
	df := absInt(sq1%8 - sq2%8)
	if dr > df {
		return dr
	}
	return df
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
