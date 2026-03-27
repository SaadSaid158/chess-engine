package engine

import "fmt"

// Perft runs a perft test and returns the node count
func Perft(b *Board, depth int) uint64 {
	if depth == 0 {
		return 1
	}

	var ml MoveList
	b.GenerateMoves(&ml)

	if depth == 1 {
		return uint64(ml.Count)
	}

	var nodes uint64
	for i := 0; i < ml.Count; i++ {
		m := ml.Moves[i]
		if b.MakeMove(m) {
			nodes += Perft(b, depth-1)
			b.UnmakeMove()
		}
	}

	return nodes
}

// PerftDivide runs perft with divide output
func PerftDivide(b *Board, depth int) uint64 {
	var ml MoveList
	b.GenerateMoves(&ml)

	var total uint64
	for i := 0; i < ml.Count; i++ {
		m := ml.Moves[i]
		if b.MakeMove(m) {
			count := Perft(b, depth-1)
			total += count
			fmt.Printf("%s: %d\n", m.String(), count)
			b.UnmakeMove()
		}
	}

	fmt.Printf("\nTotal: %d\n", total)
	return total
}
