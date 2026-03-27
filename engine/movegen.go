package engine

// Move representation: packed into uint32
// Bits 0-5:   from square
// Bits 6-11:  to square
// Bits 12-15: flags
// Bits 16-19: moved piece
// Bits 20-23: captured piece (6 = none)
type Move uint32

// Move flags
const (
	FlagNone       = 0
	FlagDoublePawn = 1
	FlagEnPassant  = 2
	FlagCastleK    = 3
	FlagCastleQ    = 4
	FlagPromoN     = 5
	FlagPromoB     = 6
	FlagPromoR     = 7
	FlagPromoQ     = 8
)

const NullMove Move = 0

func NewMove(from, to, flags, movedPiece, capturedPiece int) Move {
	return Move(from | (to << 6) | (flags << 12) | (movedPiece << 16) | (capturedPiece << 20))
}

func (m Move) From() int          { return int(m) & 0x3F }
func (m Move) To() int            { return (int(m) >> 6) & 0x3F }
func (m Move) Flags() int         { return (int(m) >> 12) & 0xF }
func (m Move) MovedPiece() int    { return (int(m) >> 16) & 0xF }
func (m Move) CapturedPiece() int { return (int(m) >> 20) & 0xF }

func (m Move) IsCapture() bool {
	return m.CapturedPiece() != 6 || m.Flags() == FlagEnPassant
}

func (m Move) IsPromotion() bool {
	f := m.Flags()
	return f >= FlagPromoN && f <= FlagPromoQ
}

func (m Move) PromoPiece() int {
	switch m.Flags() {
	case FlagPromoN:
		return Knight
	case FlagPromoB:
		return Bishop
	case FlagPromoR:
		return Rook
	case FlagPromoQ:
		return Queen
	}
	return -1
}

func (m Move) IsCastle() bool {
	return m.Flags() == FlagCastleK || m.Flags() == FlagCastleQ
}

// String returns UCI-format move string
func (m Move) String() string {
	if m == NullMove {
		return "0000"
	}
	from := m.From()
	to := m.To()
	s := squareToString(from) + squareToString(to)
	if m.IsPromotion() {
		promoChars := [4]byte{'n', 'b', 'r', 'q'}
		s += string(promoChars[m.Flags()-FlagPromoN])
	}
	return s
}

func squareToString(sq int) string {
	file := sq % 8
	rank := sq / 8
	return string([]byte{byte('a' + file), byte('1' + rank)})
}

func StringToSquare(s string) int {
	if len(s) < 2 {
		return NoSquare
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoSquare
	}
	return rank*8 + file
}

// MoveList is a fixed-size move list to avoid allocations
type MoveList struct {
	Moves [256]Move
	Count int
}

func (ml *MoveList) Add(m Move) {
	ml.Moves[ml.Count] = m
	ml.Count++
}

func (ml *MoveList) Clear() {
	ml.Count = 0
}

// Precomputed attack tables
var (
	PawnAttacks   [2][64]uint64
	KnightAttacks [64]uint64
	KingAttacks   [64]uint64

	// Magic bitboard tables for sliding pieces
	BishopMasks  [64]uint64
	RookMasks    [64]uint64
	BishopMagics [64]uint64
	RookMagics   [64]uint64
	BishopShifts [64]int
	RookShifts   [64]int
	BishopTable  [64][4096]uint64
	RookTable    [64][4096]uint64
)

// Directions
var (
	knightOffsets = [8][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	kingOffsets   = [8][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
)

func InitAttacks() {
	initPawnAttacks()
	initKnightAttacks()
	initKingAttacks()
	initMagics()
}

func initPawnAttacks() {
	for sq := 0; sq < 64; sq++ {
		file := sq % 8
		rank := sq / 8
		// White pawn attacks
		if rank < 7 {
			if file > 0 {
				PawnAttacks[White][sq] |= 1 << uint(sq+7)
			}
			if file < 7 {
				PawnAttacks[White][sq] |= 1 << uint(sq+9)
			}
		}
		// Black pawn attacks
		if rank > 0 {
			if file > 0 {
				PawnAttacks[Black][sq] |= 1 << uint(sq-9)
			}
			if file < 7 {
				PawnAttacks[Black][sq] |= 1 << uint(sq-7)
			}
		}
	}
}

func initKnightAttacks() {
	for sq := 0; sq < 64; sq++ {
		file := sq % 8
		rank := sq / 8
		for _, off := range knightOffsets {
			r := rank + off[0]
			f := file + off[1]
			if r >= 0 && r < 8 && f >= 0 && f < 8 {
				KnightAttacks[sq] |= 1 << uint(r*8+f)
			}
		}
	}
}

func initKingAttacks() {
	for sq := 0; sq < 64; sq++ {
		file := sq % 8
		rank := sq / 8
		for _, off := range kingOffsets {
			r := rank + off[0]
			f := file + off[1]
			if r >= 0 && r < 8 && f >= 0 && f < 8 {
				KingAttacks[sq] |= 1 << uint(r*8+f)
			}
		}
	}
}

// Magic bitboard implementation for sliding pieces

func bishopMask(sq int) uint64 {
	var mask uint64
	rank, file := sq/8, sq%8
	for r, f := rank+1, file+1; r < 7 && f < 7; r, f = r+1, f+1 {
		mask |= 1 << uint(r*8+f)
	}
	for r, f := rank+1, file-1; r < 7 && f > 0; r, f = r+1, f-1 {
		mask |= 1 << uint(r*8+f)
	}
	for r, f := rank-1, file+1; r > 0 && f < 7; r, f = r-1, f+1 {
		mask |= 1 << uint(r*8+f)
	}
	for r, f := rank-1, file-1; r > 0 && f > 0; r, f = r-1, f-1 {
		mask |= 1 << uint(r*8+f)
	}
	return mask
}

func rookMask(sq int) uint64 {
	var mask uint64
	rank, file := sq/8, sq%8
	for r := rank + 1; r < 7; r++ {
		mask |= 1 << uint(r*8+file)
	}
	for r := rank - 1; r > 0; r-- {
		mask |= 1 << uint(r*8+file)
	}
	for f := file + 1; f < 7; f++ {
		mask |= 1 << uint(rank*8+f)
	}
	for f := file - 1; f > 0; f-- {
		mask |= 1 << uint(rank*8+f)
	}
	return mask
}

func bishopAttacksSlow(sq int, occ uint64) uint64 {
	var attacks uint64
	rank, file := sq/8, sq%8
	for r, f := rank+1, file+1; r < 8 && f < 8; r, f = r+1, f+1 {
		attacks |= 1 << uint(r*8+f)
		if occ&(1<<uint(r*8+f)) != 0 {
			break
		}
	}
	for r, f := rank+1, file-1; r < 8 && f >= 0; r, f = r+1, f-1 {
		attacks |= 1 << uint(r*8+f)
		if occ&(1<<uint(r*8+f)) != 0 {
			break
		}
	}
	for r, f := rank-1, file+1; r >= 0 && f < 8; r, f = r-1, f+1 {
		attacks |= 1 << uint(r*8+f)
		if occ&(1<<uint(r*8+f)) != 0 {
			break
		}
	}
	for r, f := rank-1, file-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
		attacks |= 1 << uint(r*8+f)
		if occ&(1<<uint(r*8+f)) != 0 {
			break
		}
	}
	return attacks
}

func rookAttacksSlow(sq int, occ uint64) uint64 {
	var attacks uint64
	rank, file := sq/8, sq%8
	for r := rank + 1; r < 8; r++ {
		attacks |= 1 << uint(r*8+file)
		if occ&(1<<uint(r*8+file)) != 0 {
			break
		}
	}
	for r := rank - 1; r >= 0; r-- {
		attacks |= 1 << uint(r*8+file)
		if occ&(1<<uint(r*8+file)) != 0 {
			break
		}
	}
	for f := file + 1; f < 8; f++ {
		attacks |= 1 << uint(rank*8+f)
		if occ&(1<<uint(rank*8+f)) != 0 {
			break
		}
	}
	for f := file - 1; f >= 0; f-- {
		attacks |= 1 << uint(rank*8+f)
		if occ&(1<<uint(rank*8+f)) != 0 {
			break
		}
	}
	return attacks
}

// Precomputed magic numbers (found by trial)
var precomputedRookMagics = [64]uint64{
	0x0080001020400080, 0x0040001000200040, 0x0080081000200080, 0x0080040800100080,
	0x0080020400080080, 0x0080010200040080, 0x0080008001000200, 0x0080002040800100,
	0x0000800020400080, 0x0000400020005000, 0x0000801000200080, 0x0000800800100080,
	0x0000800400080080, 0x0000800200040080, 0x0000800100020080, 0x0000800040800100,
	0x0000208000400080, 0x0000404000201000, 0x0000808010002000, 0x0000808008001000,
	0x0000808004000800, 0x0000808002000400, 0x0000010100020004, 0x0000020000408104,
	0x0000208080004000, 0x0000200040005000, 0x0000100080200080, 0x0000080080100080,
	0x0000040080080080, 0x0000020080040080, 0x0000010080800200, 0x0000800080004100,
	0x0000204000800080, 0x0000200040401000, 0x0000100080802000, 0x0000080080801000,
	0x0000040080800800, 0x0000020080800400, 0x0000020001010004, 0x0000800040800100,
	0x0000204000808000, 0x0000200040008080, 0x0000100020008080, 0x0000080010008080,
	0x0000040008008080, 0x0000020004008080, 0x0000010002008080, 0x0000004081020004,
	0x0000204000800080, 0x0000200040008080, 0x0000100020008080, 0x0000080010008080,
	0x0000040008008080, 0x0000020004008080, 0x0000800100020080, 0x0000800041000080,
	0x00FFFCDDFCED714A, 0x007FFCDDFCED714A, 0x003FFFCDFFD88096, 0x0000040810002101,
	0x0001000204080011, 0x0001000204000801, 0x0001000082000401, 0x0001FFFAABFAD1A2,
}

var precomputedBishopMagics = [64]uint64{
	0x0002020202020200, 0x0002020202020000, 0x0004010202000000, 0x0004040080000000,
	0x0001104000000000, 0x0000821040000000, 0x0000410410400000, 0x0000104104104000,
	0x0000040404040400, 0x0000020202020200, 0x0000040102020000, 0x0000040400800000,
	0x0000011040000000, 0x0000008210400000, 0x0000004104104000, 0x0000002082082000,
	0x0004000808080800, 0x0002000404040400, 0x0001000202020200, 0x0000800802004000,
	0x0000800400A00000, 0x0000200100884000, 0x0000400082082000, 0x0000200041041000,
	0x0002080010101000, 0x0001040008080800, 0x0000208004010400, 0x0000404004010200,
	0x0000840000802000, 0x0000404002011000, 0x0000808001041000, 0x0000404000820800,
	0x0001041000202000, 0x0000820800101000, 0x0000104400080800, 0x0000020080080080,
	0x0000404040040100, 0x0000808100020100, 0x0001010100020800, 0x0000808080010400,
	0x0000820820004000, 0x0000410410002000, 0x0000082088001000, 0x0000002011000800,
	0x0000080100400400, 0x0001010101000200, 0x0002020202000400, 0x0001010101000200,
	0x0000410410400000, 0x0000208208200000, 0x0000002084100000, 0x0000000020880000,
	0x0000001002020000, 0x0000040408020000, 0x0004040404040000, 0x0002020202020000,
	0x0000104104104000, 0x0000002082082000, 0x0000000020841000, 0x0000000000208800,
	0x0000000010020200, 0x0000000404080200, 0x0000040404040400, 0x0002020202020200,
}

func initMagics() {
	for sq := 0; sq < 64; sq++ {
		BishopMasks[sq] = bishopMask(sq)
		RookMasks[sq] = rookMask(sq)
		BishopMagics[sq] = precomputedBishopMagics[sq]
		RookMagics[sq] = precomputedRookMagics[sq]
		BishopShifts[sq] = 64 - PopCount(BishopMasks[sq])
		RookShifts[sq] = 64 - PopCount(RookMasks[sq])

		// Fill bishop table
		mask := BishopMasks[sq]
		n := PopCount(mask)
		for i := 0; i < (1 << uint(n)); i++ {
			occ := indexToOccupancy(i, n, mask)
			idx := (occ * BishopMagics[sq]) >> uint(BishopShifts[sq])
			BishopTable[sq][idx] = bishopAttacksSlow(sq, occ)
		}

		// Fill rook table
		mask = RookMasks[sq]
		n = PopCount(mask)
		for i := 0; i < (1 << uint(n)); i++ {
			occ := indexToOccupancy(i, n, mask)
			idx := (occ * RookMagics[sq]) >> uint(RookShifts[sq])
			RookTable[sq][idx] = rookAttacksSlow(sq, occ)
		}
	}
}

func indexToOccupancy(index, bits int, mask uint64) uint64 {
	var occ uint64
	for i := 0; i < bits; i++ {
		sq := PopLSB(&mask)
		if index&(1<<uint(i)) != 0 {
			occ |= 1 << uint(sq)
		}
	}
	return occ
}

// BishopAttacks returns bishop attacks for a square given occupancy
func BishopAttacks(sq int, occ uint64) uint64 {
	occ &= BishopMasks[sq]
	idx := (occ * BishopMagics[sq]) >> uint(BishopShifts[sq])
	return BishopTable[sq][idx]
}

// RookAttacks returns rook attacks for a square given occupancy
func RookAttacks(sq int, occ uint64) uint64 {
	occ &= RookMasks[sq]
	idx := (occ * RookMagics[sq]) >> uint(RookShifts[sq])
	return RookTable[sq][idx]
}

// QueenAttacks returns queen attacks
func QueenAttacks(sq int, occ uint64) uint64 {
	return BishopAttacks(sq, occ) | RookAttacks(sq, occ)
}

// GenerateMoves generates all legal moves
func (b *Board) GenerateMoves(ml *MoveList) {
	ml.Clear()
	b.generatePseudoLegalMoves(ml)
	b.filterLegal(ml)
}

// GenerateCaptures generates only capture moves (for quiescence)
func (b *Board) GenerateCaptures(ml *MoveList) {
	ml.Clear()
	b.generatePseudoLegalCaptures(ml)
	b.filterLegal(ml)
}

func (b *Board) generatePseudoLegalMoves(ml *MoveList) {
	side := b.Side
	opp := side ^ 1
	friendly := b.Occupied[side]
	enemy := b.Occupied[opp]

	// Pawns
	b.generatePawnMoves(ml, side, friendly, enemy)

	// Knights
	bb := b.Pieces[side][Knight]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := KnightAttacks[from] & ^friendly
		for attacks != 0 {
			to := PopLSB(&attacks)
			captured := 6
			if enemy&(1<<uint(to)) != 0 {
				_, captured = b.PieceOn(to)
			}
			ml.Add(NewMove(from, to, FlagNone, Knight, captured))
		}
	}

	// Bishops
	bb = b.Pieces[side][Bishop]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := BishopAttacks(from, b.AllOccupied) & ^friendly
		for attacks != 0 {
			to := PopLSB(&attacks)
			captured := 6
			if enemy&(1<<uint(to)) != 0 {
				_, captured = b.PieceOn(to)
			}
			ml.Add(NewMove(from, to, FlagNone, Bishop, captured))
		}
	}

	// Rooks
	bb = b.Pieces[side][Rook]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := RookAttacks(from, b.AllOccupied) & ^friendly
		for attacks != 0 {
			to := PopLSB(&attacks)
			captured := 6
			if enemy&(1<<uint(to)) != 0 {
				_, captured = b.PieceOn(to)
			}
			ml.Add(NewMove(from, to, FlagNone, Rook, captured))
		}
	}

	// Queens
	bb = b.Pieces[side][Queen]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := QueenAttacks(from, b.AllOccupied) & ^friendly
		for attacks != 0 {
			to := PopLSB(&attacks)
			captured := 6
			if enemy&(1<<uint(to)) != 0 {
				_, captured = b.PieceOn(to)
			}
			ml.Add(NewMove(from, to, FlagNone, Queen, captured))
		}
	}

	// King
	bb = b.Pieces[side][King]
	if bb != 0 {
		from := LSB(bb)
		attacks := KingAttacks[from] & ^friendly
		for attacks != 0 {
			to := PopLSB(&attacks)
			captured := 6
			if enemy&(1<<uint(to)) != 0 {
				_, captured = b.PieceOn(to)
			}
			ml.Add(NewMove(from, to, FlagNone, King, captured))
		}

		// Castling
		if side == White {
			if b.CastleRights&WhiteKingSide != 0 {
				if b.AllOccupied&((1<<5)|(1<<6)) == 0 {
					if !b.IsSquareAttacked(4, Black) && !b.IsSquareAttacked(5, Black) && !b.IsSquareAttacked(6, Black) {
						ml.Add(NewMove(4, 6, FlagCastleK, King, 6))
					}
				}
			}
			if b.CastleRights&WhiteQueenSide != 0 {
				if b.AllOccupied&((1<<1)|(1<<2)|(1<<3)) == 0 {
					if !b.IsSquareAttacked(4, Black) && !b.IsSquareAttacked(3, Black) && !b.IsSquareAttacked(2, Black) {
						ml.Add(NewMove(4, 2, FlagCastleQ, King, 6))
					}
				}
			}
		} else {
			if b.CastleRights&BlackKingSide != 0 {
				if b.AllOccupied&((1<<61)|(1<<62)) == 0 {
					if !b.IsSquareAttacked(60, White) && !b.IsSquareAttacked(61, White) && !b.IsSquareAttacked(62, White) {
						ml.Add(NewMove(60, 62, FlagCastleK, King, 6))
					}
				}
			}
			if b.CastleRights&BlackQueenSide != 0 {
				if b.AllOccupied&((1<<57)|(1<<58)|(1<<59)) == 0 {
					if !b.IsSquareAttacked(60, White) && !b.IsSquareAttacked(59, White) && !b.IsSquareAttacked(58, White) {
						ml.Add(NewMove(60, 58, FlagCastleQ, King, 6))
					}
				}
			}
		}
	}
}

func (b *Board) generatePawnMoves(ml *MoveList, side int, friendly, enemy uint64) {
	pawns := b.Pieces[side][Pawn]
	if side == White {
		// Single push
		push1 := (pawns << 8) & ^b.AllOccupied
		push2 := ((push1 & RankMask[2]) << 8) & ^b.AllOccupied

		tmp := push1
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to - 8
			if to >= 56 { // promotion
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, 6))
				ml.Add(NewMove(from, to, FlagPromoR, Pawn, 6))
				ml.Add(NewMove(from, to, FlagPromoB, Pawn, 6))
				ml.Add(NewMove(from, to, FlagPromoN, Pawn, 6))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, 6))
			}
		}

		tmp = push2
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to - 16
			ml.Add(NewMove(from, to, FlagDoublePawn, Pawn, 6))
		}

		// Captures
		captL := (pawns << 7) & enemy & ^FileMask[7]
		captR := (pawns << 9) & enemy & ^FileMask[0]

		tmp = captL
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to - 7
			_, captured := b.PieceOn(to)
			if to >= 56 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoR, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoB, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoN, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}

		tmp = captR
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to - 9
			_, captured := b.PieceOn(to)
			if to >= 56 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoR, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoB, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoN, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}

		// En passant
		if b.EnPassant != NoSquare {
			epBB := uint64(1) << uint(b.EnPassant)
			epL := (pawns << 7) & epBB & ^FileMask[7]
			epR := (pawns << 9) & epBB & ^FileMask[0]
			if epL != 0 {
				ml.Add(NewMove(b.EnPassant-7, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
			if epR != 0 {
				ml.Add(NewMove(b.EnPassant-9, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
		}
	} else {
		// Black pawns
		push1 := (pawns >> 8) & ^b.AllOccupied
		push2 := ((push1 & RankMask[5]) >> 8) & ^b.AllOccupied

		tmp := push1
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to + 8
			if to < 8 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, 6))
				ml.Add(NewMove(from, to, FlagPromoR, Pawn, 6))
				ml.Add(NewMove(from, to, FlagPromoB, Pawn, 6))
				ml.Add(NewMove(from, to, FlagPromoN, Pawn, 6))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, 6))
			}
		}

		tmp = push2
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to + 16
			ml.Add(NewMove(from, to, FlagDoublePawn, Pawn, 6))
		}

		captL := (pawns >> 9) & enemy & ^FileMask[7]
		captR := (pawns >> 7) & enemy & ^FileMask[0]

		tmp = captL
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to + 9
			_, captured := b.PieceOn(to)
			if to < 8 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoR, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoB, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoN, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}

		tmp = captR
		for tmp != 0 {
			to := PopLSB(&tmp)
			from := to + 7
			_, captured := b.PieceOn(to)
			if to < 8 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoR, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoB, Pawn, captured))
				ml.Add(NewMove(from, to, FlagPromoN, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}

		if b.EnPassant != NoSquare {
			epBB := uint64(1) << uint(b.EnPassant)
			epL := (pawns >> 9) & epBB & ^FileMask[7]
			epR := (pawns >> 7) & epBB & ^FileMask[0]
			if epL != 0 {
				ml.Add(NewMove(b.EnPassant+9, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
			if epR != 0 {
				ml.Add(NewMove(b.EnPassant+7, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
		}
	}
}

func (b *Board) generatePseudoLegalCaptures(ml *MoveList) {
	side := b.Side
	opp := side ^ 1
	_ = b.Occupied[side]
	enemy := b.Occupied[opp]

	// Pawn captures + promotions
	pawns := b.Pieces[side][Pawn]
	if side == White {
		captL := (pawns << 7) & enemy & ^FileMask[7]
		captR := (pawns << 9) & enemy & ^FileMask[0]
		push1Promo := (pawns << 8) & ^b.AllOccupied & RankMask[7]

		for captL != 0 {
			to := PopLSB(&captL)
			from := to - 7
			_, captured := b.PieceOn(to)
			if to >= 56 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}
		for captR != 0 {
			to := PopLSB(&captR)
			from := to - 9
			_, captured := b.PieceOn(to)
			if to >= 56 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}
		for push1Promo != 0 {
			to := PopLSB(&push1Promo)
			from := to - 8
			ml.Add(NewMove(from, to, FlagPromoQ, Pawn, 6))
		}
		if b.EnPassant != NoSquare {
			epBB := uint64(1) << uint(b.EnPassant)
			epL := (pawns << 7) & epBB & ^FileMask[7]
			epR := (pawns << 9) & epBB & ^FileMask[0]
			if epL != 0 {
				ml.Add(NewMove(b.EnPassant-7, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
			if epR != 0 {
				ml.Add(NewMove(b.EnPassant-9, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
		}
	} else {
		captL := (pawns >> 9) & enemy & ^FileMask[7]
		captR := (pawns >> 7) & enemy & ^FileMask[0]
		push1Promo := (pawns >> 8) & ^b.AllOccupied & RankMask[0]

		for captL != 0 {
			to := PopLSB(&captL)
			from := to + 9
			_, captured := b.PieceOn(to)
			if to < 8 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}
		for captR != 0 {
			to := PopLSB(&captR)
			from := to + 7
			_, captured := b.PieceOn(to)
			if to < 8 {
				ml.Add(NewMove(from, to, FlagPromoQ, Pawn, captured))
			} else {
				ml.Add(NewMove(from, to, FlagNone, Pawn, captured))
			}
		}
		for push1Promo != 0 {
			to := PopLSB(&push1Promo)
			from := to + 8
			ml.Add(NewMove(from, to, FlagPromoQ, Pawn, 6))
		}
		if b.EnPassant != NoSquare {
			epBB := uint64(1) << uint(b.EnPassant)
			epL := (pawns >> 9) & epBB & ^FileMask[7]
			epR := (pawns >> 7) & epBB & ^FileMask[0]
			if epL != 0 {
				ml.Add(NewMove(b.EnPassant+9, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
			if epR != 0 {
				ml.Add(NewMove(b.EnPassant+7, b.EnPassant, FlagEnPassant, Pawn, Pawn))
			}
		}
	}

	// Knights
	bb := b.Pieces[side][Knight]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := KnightAttacks[from] & enemy
		for attacks != 0 {
			to := PopLSB(&attacks)
			_, captured := b.PieceOn(to)
			ml.Add(NewMove(from, to, FlagNone, Knight, captured))
		}
	}

	// Bishops
	bb = b.Pieces[side][Bishop]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := BishopAttacks(from, b.AllOccupied) & enemy
		for attacks != 0 {
			to := PopLSB(&attacks)
			_, captured := b.PieceOn(to)
			ml.Add(NewMove(from, to, FlagNone, Bishop, captured))
		}
	}

	// Rooks
	bb = b.Pieces[side][Rook]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := RookAttacks(from, b.AllOccupied) & enemy
		for attacks != 0 {
			to := PopLSB(&attacks)
			_, captured := b.PieceOn(to)
			ml.Add(NewMove(from, to, FlagNone, Rook, captured))
		}
	}

	// Queens
	bb = b.Pieces[side][Queen]
	for bb != 0 {
		from := PopLSB(&bb)
		attacks := QueenAttacks(from, b.AllOccupied) & enemy
		for attacks != 0 {
			to := PopLSB(&attacks)
			_, captured := b.PieceOn(to)
			ml.Add(NewMove(from, to, FlagNone, Queen, captured))
		}
	}

	// King captures
	bb = b.Pieces[side][King]
	if bb != 0 {
		from := LSB(bb)
		attacks := KingAttacks[from] & enemy
		for attacks != 0 {
			to := PopLSB(&attacks)
			_, captured := b.PieceOn(to)
			ml.Add(NewMove(from, to, FlagNone, King, captured))
		}
	}
}

func (b *Board) filterLegal(ml *MoveList) {
	write := 0
	for i := 0; i < ml.Count; i++ {
		m := ml.Moves[i]
		if b.MakeMove(m) {
			b.UnmakeMove()
			ml.Moves[write] = m
			write++
		}
	}
	ml.Count = write
}

// MakeMove makes a move on the board. Returns false if the move leaves the king in check.
func (b *Board) MakeMove(m Move) bool {
	side := b.Side
	opp := side ^ 1
	from := m.From()
	to := m.To()
	flags := m.Flags()
	movedPiece := m.MovedPiece()
	capturedPiece := m.CapturedPiece()

	// Save undo info
	undo := &b.History[b.HistPly]
	undo.Move = m
	undo.CastleRights = b.CastleRights
	undo.EnPassant = b.EnPassant
	undo.HalfMove = b.HalfMove
	undo.Hash = b.Hash
	undo.PawnHash = b.PawnHash
	undo.CapturedPiece = capturedPiece

	fromBit := uint64(1) << uint(from)
	toBit := uint64(1) << uint(to)

	// Remove piece from source
	b.Pieces[side][movedPiece] ^= fromBit
	b.Hash ^= ZobristPieces[side][movedPiece][from]
	if movedPiece == Pawn {
		b.PawnHash ^= ZobristPieces[side][Pawn][from]
	}

	// Handle capture
	if capturedPiece != 6 && flags != FlagEnPassant {
		b.Pieces[opp][capturedPiece] ^= toBit
		b.Hash ^= ZobristPieces[opp][capturedPiece][to]
		if capturedPiece == Pawn {
			b.PawnHash ^= ZobristPieces[opp][Pawn][to]
		}
	}

	// Place piece on target
	placePiece := movedPiece
	if m.IsPromotion() {
		placePiece = m.PromoPiece()
		b.PawnHash ^= ZobristPieces[side][Pawn][from] // already XORed above, will cancel
		// Actually we already removed the pawn hash above, no need to re-XOR
		// The pawn is gone, promoted piece is not a pawn
	}
	b.Pieces[side][placePiece] |= toBit
	b.Hash ^= ZobristPieces[side][placePiece][to]
	if placePiece == Pawn {
		b.PawnHash ^= ZobristPieces[side][Pawn][to]
	}

	// If promotion, we already put the promo piece, but also need to remove
	// the pawn entry we just placed (we placed movedPiece=Pawn but want promoPiece)
	if m.IsPromotion() {
		// We placed placePiece (promo) above. But we also need to NOT have the pawn there.
		// movedPiece is Pawn, placePiece is the promoted piece.
		// We removed Pawn from 'from', placed placePiece at 'to'. Good.
		// But wait - we did Pieces[side][movedPiece] ^= fromBit (removed pawn from source) - good
		// Then Pieces[side][placePiece] |= toBit (added promoted piece at dest) - good
		// Nothing more needed for the piece arrays.
	}

	// En passant capture
	if flags == FlagEnPassant {
		var epCapSq int
		if side == White {
			epCapSq = to - 8
		} else {
			epCapSq = to + 8
		}
		epCapBit := uint64(1) << uint(epCapSq)
		b.Pieces[opp][Pawn] ^= epCapBit
		b.Hash ^= ZobristPieces[opp][Pawn][epCapSq]
		b.PawnHash ^= ZobristPieces[opp][Pawn][epCapSq]
	}

	// Remove old en passant from hash
	if b.EnPassant != NoSquare {
		b.Hash ^= ZobristEnPassant[b.EnPassant%8]
	}

	// Set new en passant
	b.EnPassant = NoSquare
	if flags == FlagDoublePawn {
		if side == White {
			b.EnPassant = from + 8
		} else {
			b.EnPassant = from - 8
		}
		b.Hash ^= ZobristEnPassant[b.EnPassant%8]
	}

	// Castling - move the rook
	if flags == FlagCastleK {
		if side == White {
			b.Pieces[White][Rook] ^= (1 << 7) | (1 << 5)
			b.Hash ^= ZobristPieces[White][Rook][7] ^ ZobristPieces[White][Rook][5]
		} else {
			b.Pieces[Black][Rook] ^= (1 << 63) | (1 << 61)
			b.Hash ^= ZobristPieces[Black][Rook][63] ^ ZobristPieces[Black][Rook][61]
		}
	} else if flags == FlagCastleQ {
		if side == White {
			b.Pieces[White][Rook] ^= (1 << 0) | (1 << 3)
			b.Hash ^= ZobristPieces[White][Rook][0] ^ ZobristPieces[White][Rook][3]
		} else {
			b.Pieces[Black][Rook] ^= (1 << 56) | (1 << 59)
			b.Hash ^= ZobristPieces[Black][Rook][56] ^ ZobristPieces[Black][Rook][59]
		}
	}

	// Update castling rights
	b.Hash ^= ZobristCastling[b.CastleRights]
	b.CastleRights &= CastleRightsUpdate[from] & CastleRightsUpdate[to]
	b.Hash ^= ZobristCastling[b.CastleRights]

	// Update half move clock
	if movedPiece == Pawn || capturedPiece != 6 || flags == FlagEnPassant {
		b.HalfMove = 0
	} else {
		b.HalfMove++
	}

	// Update full move number
	if side == Black {
		b.FullMove++
	}

	// Switch side
	b.Side = opp
	b.Hash ^= ZobristSide

	b.updateOccupancy()

	b.HistPly++
	b.GamePly++
	b.HashHistory[b.GamePly] = b.Hash

	// Check if own king is in check (illegal move)
	kingSq := LSB(b.Pieces[side][King])
	if b.IsSquareAttacked(kingSq, opp) {
		b.UnmakeMove()
		return false
	}

	return true
}

// UnmakeMove undoes the last move
func (b *Board) UnmakeMove() {
	b.HistPly--
	b.GamePly--

	undo := &b.History[b.HistPly]
	m := undo.Move

	b.Side ^= 1
	side := b.Side
	opp := side ^ 1

	from := m.From()
	to := m.To()
	flags := m.Flags()
	movedPiece := m.MovedPiece()

	fromBit := uint64(1) << uint(from)
	toBit := uint64(1) << uint(to)

	// Remove piece from destination
	if m.IsPromotion() {
		promoPiece := m.PromoPiece()
		b.Pieces[side][promoPiece] ^= toBit
	} else {
		b.Pieces[side][movedPiece] ^= toBit
	}

	// Restore piece to source
	b.Pieces[side][movedPiece] |= fromBit

	// Restore captured piece
	if undo.CapturedPiece != 6 && flags != FlagEnPassant {
		b.Pieces[opp][undo.CapturedPiece] |= toBit
	}

	// Restore en passant capture
	if flags == FlagEnPassant {
		var epCapSq int
		if side == White {
			epCapSq = to - 8
		} else {
			epCapSq = to + 8
		}
		b.Pieces[opp][Pawn] |= 1 << uint(epCapSq)
	}

	// Restore castling rook
	if flags == FlagCastleK {
		if side == White {
			b.Pieces[White][Rook] ^= (1 << 7) | (1 << 5)
		} else {
			b.Pieces[Black][Rook] ^= (1 << 63) | (1 << 61)
		}
	} else if flags == FlagCastleQ {
		if side == White {
			b.Pieces[White][Rook] ^= (1 << 0) | (1 << 3)
		} else {
			b.Pieces[Black][Rook] ^= (1 << 56) | (1 << 59)
		}
	}

	b.CastleRights = undo.CastleRights
	b.EnPassant = undo.EnPassant
	b.HalfMove = undo.HalfMove
	b.Hash = undo.Hash
	b.PawnHash = undo.PawnHash

	if side == Black {
		b.FullMove--
	}

	b.updateOccupancy()
}

// MakeNullMove makes a null move (pass)
func (b *Board) MakeNullMove() {
	undo := &b.History[b.HistPly]
	undo.Move = NullMove
	undo.CastleRights = b.CastleRights
	undo.EnPassant = b.EnPassant
	undo.HalfMove = b.HalfMove
	undo.Hash = b.Hash
	undo.PawnHash = b.PawnHash
	undo.CapturedPiece = 6

	if b.EnPassant != NoSquare {
		b.Hash ^= ZobristEnPassant[b.EnPassant%8]
	}
	b.EnPassant = NoSquare
	b.Side ^= 1
	b.Hash ^= ZobristSide
	b.HalfMove = 0

	b.HistPly++
	b.GamePly++
	b.HashHistory[b.GamePly] = b.Hash
}

// UnmakeNullMove undoes a null move
func (b *Board) UnmakeNullMove() {
	b.HistPly--
	b.GamePly--

	undo := &b.History[b.HistPly]
	b.Side ^= 1
	b.EnPassant = undo.EnPassant
	b.HalfMove = undo.HalfMove
	b.Hash = undo.Hash
	b.PawnHash = undo.PawnHash
}

// ParseMove parses a UCI move string and returns the matching legal move
func (b *Board) ParseMove(s string) Move {
	if len(s) < 4 {
		return NullMove
	}
	from := StringToSquare(s[0:2])
	to := StringToSquare(s[2:4])
	promo := -1
	if len(s) == 5 {
		switch s[4] {
		case 'n':
			promo = Knight
		case 'b':
			promo = Bishop
		case 'r':
			promo = Rook
		case 'q':
			promo = Queen
		}
	}

	var ml MoveList
	b.GenerateMoves(&ml)
	for i := 0; i < ml.Count; i++ {
		m := ml.Moves[i]
		if m.From() == from && m.To() == to {
			if promo >= 0 {
				if m.IsPromotion() && m.PromoPiece() == promo {
					return m
				}
			} else if !m.IsPromotion() {
				return m
			}
		}
	}
	return NullMove
}
