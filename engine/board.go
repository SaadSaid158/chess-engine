package engine

import (
	"fmt"
	"math/bits"
	"strings"
)

// Piece constants
const (
	Pawn   = 0
	Knight = 1
	Bishop = 2
	Rook   = 3
	Queen  = 4
	King   = 5
)

// Color constants
const (
	White = 0
	Black = 1
)

// Castling rights bits
const (
	WhiteKingSide  = 1
	WhiteQueenSide = 2
	BlackKingSide  = 4
	BlackQueenSide = 8
)

// Square constants
const (
	NoSquare = -1
)

// File and rank masks
var (
	FileMask [8]uint64
	RankMask [8]uint64
)

// Board represents the full chess position
type Board struct {
	Pieces       [2][6]uint64 // [color][piece] bitboards
	Occupied     [2]uint64    // occupancy per side
	AllOccupied  uint64       // all pieces
	Side         int          // side to move (0=White, 1=Black)
	CastleRights int          // castling rights bitmask
	EnPassant    int          // en passant target square (-1 if none)
	HalfMove     int          // half move clock
	FullMove     int          // full move number
	Hash         uint64       // Zobrist hash
	PawnHash     uint64       // pawn-specific hash

	// History for unmake
	History [1024]UndoInfo
	HistPly int
	// Repetition detection
	HashHistory [1024]uint64
	GamePly     int
}

// UndoInfo stores state needed to unmake a move
type UndoInfo struct {
	Move          Move
	CastleRights  int
	EnPassant     int
	HalfMove      int
	Hash          uint64
	PawnHash      uint64
	CapturedPiece int // -1 if none
}

func init() {
	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			FileMask[f] |= 1 << uint(r*8+f)
			RankMask[r] |= 1 << uint(r*8+f)
		}
	}
}

// NewBoard creates a board from the starting position
func NewBoard() *Board {
	return BoardFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

// BoardFromFEN creates a board from a FEN string
func BoardFromFEN(fen string) *Board {
	b := &Board{}
	b.EnPassant = NoSquare

	parts := strings.Fields(fen)
	if len(parts) < 4 {
		return b
	}

	// Parse piece placement
	rank := 7
	file := 0
	for _, ch := range parts[0] {
		switch {
		case ch == '/':
			rank--
			file = 0
		case ch >= '1' && ch <= '8':
			file += int(ch - '0')
		default:
			sq := rank*8 + file
			color, piece := charToPiece(ch)
			if piece >= 0 {
				b.Pieces[color][piece] |= 1 << uint(sq)
			}
			file++
		}
	}

	// Side to move
	if parts[1] == "b" {
		b.Side = Black
	} else {
		b.Side = White
	}

	// Castling rights
	for _, ch := range parts[2] {
		switch ch {
		case 'K':
			b.CastleRights |= WhiteKingSide
		case 'Q':
			b.CastleRights |= WhiteQueenSide
		case 'k':
			b.CastleRights |= BlackKingSide
		case 'q':
			b.CastleRights |= BlackQueenSide
		}
	}

	// En passant
	if parts[3] != "-" {
		file := int(parts[3][0] - 'a')
		rank := int(parts[3][1] - '1')
		b.EnPassant = rank*8 + file
	}

	// Half/full move
	if len(parts) >= 5 {
		fmt.Sscanf(parts[4], "%d", &b.HalfMove)
	}
	if len(parts) >= 6 {
		fmt.Sscanf(parts[5], "%d", &b.FullMove)
	}

	b.updateOccupancy()
	b.Hash = b.computeHash()
	b.PawnHash = b.computePawnHash()
	b.HashHistory[0] = b.Hash
	b.GamePly = 0

	return b
}

func charToPiece(ch rune) (int, int) {
	color := White
	if ch >= 'a' && ch <= 'z' {
		color = Black
		ch = ch - 32 // to upper
	}
	switch ch {
	case 'P':
		return color, Pawn
	case 'N':
		return color, Knight
	case 'B':
		return color, Bishop
	case 'R':
		return color, Rook
	case 'Q':
		return color, Queen
	case 'K':
		return color, King
	}
	return 0, -1
}

func (b *Board) updateOccupancy() {
	b.Occupied[White] = 0
	b.Occupied[Black] = 0
	for p := Pawn; p <= King; p++ {
		b.Occupied[White] |= b.Pieces[White][p]
		b.Occupied[Black] |= b.Pieces[Black][p]
	}
	b.AllOccupied = b.Occupied[White] | b.Occupied[Black]
}

func (b *Board) computeHash() uint64 {
	var h uint64
	for color := 0; color < 2; color++ {
		for piece := 0; piece < 6; piece++ {
			bb := b.Pieces[color][piece]
			for bb != 0 {
				sq := PopLSB(&bb)
				h ^= ZobristPieces[color][piece][sq]
			}
		}
	}
	h ^= ZobristCastling[b.CastleRights]
	if b.EnPassant != NoSquare {
		h ^= ZobristEnPassant[b.EnPassant%8]
	}
	if b.Side == Black {
		h ^= ZobristSide
	}
	return h
}

func (b *Board) computePawnHash() uint64 {
	var h uint64
	for color := 0; color < 2; color++ {
		bb := b.Pieces[color][Pawn]
		for bb != 0 {
			sq := PopLSB(&bb)
			h ^= ZobristPieces[color][Pawn][sq]
		}
	}
	return h
}

// PieceOn returns (color, piece) at a given square, or (-1,-1) if empty
func (b *Board) PieceOn(sq int) (int, int) {
	bit := uint64(1) << uint(sq)
	for color := 0; color < 2; color++ {
		if b.Occupied[color]&bit == 0 {
			continue
		}
		for piece := 0; piece < 6; piece++ {
			if b.Pieces[color][piece]&bit != 0 {
				return color, piece
			}
		}
	}
	return -1, -1
}

// IsSquareAttacked checks if a square is attacked by the given side
func (b *Board) IsSquareAttacked(sq int, bySide int) bool {
	// Pawn attacks
	if PawnAttacks[bySide^1][sq]&b.Pieces[bySide][Pawn] != 0 {
		return true
	}
	// Knight attacks
	if KnightAttacks[sq]&b.Pieces[bySide][Knight] != 0 {
		return true
	}
	// King attacks
	if KingAttacks[sq]&b.Pieces[bySide][King] != 0 {
		return true
	}
	// Bishop/Queen attacks (diagonals)
	bishopAttacks := BishopAttacks(sq, b.AllOccupied)
	if bishopAttacks&(b.Pieces[bySide][Bishop]|b.Pieces[bySide][Queen]) != 0 {
		return true
	}
	// Rook/Queen attacks (files/ranks)
	rookAttacks := RookAttacks(sq, b.AllOccupied)
	if rookAttacks&(b.Pieces[bySide][Rook]|b.Pieces[bySide][Queen]) != 0 {
		return true
	}
	return false
}

// InCheck returns true if the side to move is in check
func (b *Board) InCheck() bool {
	kingBB := b.Pieces[b.Side][King]
	if kingBB == 0 {
		return false
	}
	kingSq := LSB(kingBB)
	return b.IsSquareAttacked(kingSq, b.Side^1)
}

// IsRepetition checks if the current position has occurred before
func (b *Board) IsRepetition() bool {
	for i := b.GamePly - 2; i >= 0 && i >= b.GamePly-b.HalfMove; i -= 2 {
		if b.HashHistory[i] == b.Hash {
			return true
		}
	}
	return false
}

// IsInsufficientMaterial reports dead positions with no mating material.
func (b *Board) IsInsufficientMaterial() bool {
	if b.Pieces[White][Pawn]|b.Pieces[Black][Pawn] != 0 {
		return false
	}
	if b.Pieces[White][Rook]|b.Pieces[Black][Rook] != 0 {
		return false
	}
	if b.Pieces[White][Queen]|b.Pieces[Black][Queen] != 0 {
		return false
	}

	wKnights := PopCount(b.Pieces[White][Knight])
	bKnights := PopCount(b.Pieces[Black][Knight])
	wBishops := PopCount(b.Pieces[White][Bishop])
	bBishops := PopCount(b.Pieces[Black][Bishop])
	totalMinors := wKnights + bKnights + wBishops + bBishops

	switch totalMinors {
	case 0:
		return true
	case 1:
		return true
	}

	if totalMinors == 2 && wBishops+bBishops == 0 &&
		((wKnights == 2 && bKnights == 0) || (bKnights == 2 && wKnights == 0)) {
		return true
	}

	if wKnights+bKnights == 0 && wBishops == 1 && bBishops == 1 {
		wSq := LSB(b.Pieces[White][Bishop])
		bSq := LSB(b.Pieces[Black][Bishop])
		if wSq < 64 && bSq < 64 && squareColor(wSq) == squareColor(bSq) {
			return true
		}
	}

	return false
}

// FEN returns the FEN string for the current position
func (b *Board) FEN() string {
	var sb strings.Builder
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			sq := rank*8 + file
			c, p := b.PieceOn(sq)
			if p < 0 {
				empty++
			} else {
				if empty > 0 {
					sb.WriteByte(byte('0' + empty))
					empty = 0
				}
				sb.WriteByte(pieceToChar(c, p))
			}
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if rank > 0 {
			sb.WriteByte('/')
		}
	}

	if b.Side == White {
		sb.WriteString(" w ")
	} else {
		sb.WriteString(" b ")
	}

	castle := ""
	if b.CastleRights&WhiteKingSide != 0 {
		castle += "K"
	}
	if b.CastleRights&WhiteQueenSide != 0 {
		castle += "Q"
	}
	if b.CastleRights&BlackKingSide != 0 {
		castle += "k"
	}
	if b.CastleRights&BlackQueenSide != 0 {
		castle += "q"
	}
	if castle == "" {
		castle = "-"
	}
	sb.WriteString(castle)

	if b.EnPassant != NoSquare {
		f := b.EnPassant % 8
		r := b.EnPassant / 8
		sb.WriteString(fmt.Sprintf(" %c%c", 'a'+f, '1'+r))
	} else {
		sb.WriteString(" -")
	}

	sb.WriteString(fmt.Sprintf(" %d %d", b.HalfMove, b.FullMove))
	return sb.String()
}

func pieceToChar(color, piece int) byte {
	chars := [6]byte{'P', 'N', 'B', 'R', 'Q', 'K'}
	ch := chars[piece]
	if color == Black {
		ch += 32 // lowercase
	}
	return ch
}

// Bit manipulation helpers
func LSB(bb uint64) int {
	if bb == 0 {
		return 64
	}
	return bits.TrailingZeros64(bb)
}

func PopLSB(bb *uint64) int {
	sq := LSB(*bb)
	*bb &= *bb - 1
	return sq
}

func PopCount(bb uint64) int {
	return bits.OnesCount64(bb)
}

// Castling update table
var CastleRightsUpdate [64]int

func init() {
	for i := 0; i < 64; i++ {
		CastleRightsUpdate[i] = 15 // all rights preserved by default
	}
	CastleRightsUpdate[0] = ^WhiteQueenSide & 15                    // a1 rook
	CastleRightsUpdate[7] = ^WhiteKingSide & 15                     // h1 rook
	CastleRightsUpdate[4] = ^(WhiteKingSide | WhiteQueenSide) & 15  // e1 king
	CastleRightsUpdate[56] = ^BlackQueenSide & 15                   // a8 rook
	CastleRightsUpdate[63] = ^BlackKingSide & 15                    // h8 rook
	CastleRightsUpdate[60] = ^(BlackKingSide | BlackQueenSide) & 15 // e8 king
}

func squareColor(sq int) int {
	return (sq/8 + sq%8) & 1
}
