package engine

import (
	"math/rand"
)

// Zobrist hash keys
var (
	ZobristPieces    [2][6][64]uint64 // [color][piece][square]
	ZobristCastling  [16]uint64
	ZobristEnPassant [8]uint64 // file-based
	ZobristSide      uint64
)

func InitZobrist() {
	rng := rand.New(rand.NewSource(0x1234567890ABCDEF))
	for c := 0; c < 2; c++ {
		for p := 0; p < 6; p++ {
			for sq := 0; sq < 64; sq++ {
				ZobristPieces[c][p][sq] = rng.Uint64()
			}
		}
	}
	for i := 0; i < 16; i++ {
		ZobristCastling[i] = rng.Uint64()
	}
	for i := 0; i < 8; i++ {
		ZobristEnPassant[i] = rng.Uint64()
	}
	ZobristSide = rng.Uint64()
}
