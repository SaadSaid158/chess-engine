package engine

// Transposition table entry types
const (
	TTExact = 0
	TTAlpha = 1 // upper bound (fail-low)
	TTBeta  = 2 // lower bound (fail-high)
	TTNone  = 3
)

// TTEntry is a single transposition table entry
type TTEntry struct {
	Hash  uint64
	Depth int8
	Score int16
	Flag  uint8
	Move  Move
}

// TransTable is the transposition table
type TransTable struct {
	entries    []TTEntry
	ages       []uint8
	size       uint64
	mask       uint64
	generation uint8
}

// NewTransTable creates a new transposition table with the given size in MB
func NewTransTable(sizeMB int) *TransTable {
	entries := uint64(sizeMB) * 1024 * 1024 / 16 // sizeof TTEntry ~16 bytes
	// Round down to power of 2
	size := uint64(1)
	for size*2 <= entries {
		size *= 2
	}
	tt := &TransTable{
		entries: make([]TTEntry, size),
		ages:    make([]uint8, size),
		size:    size,
		mask:    size - 1,
	}
	return tt
}

// NewSearch advances the generation so older entries become easier to replace.
func (tt *TransTable) NewSearch() {
	tt.generation++
}

// Probe looks up a position in the transposition table
func (tt *TransTable) Probe(hash uint64) (TTEntry, bool) {
	idx := hash & tt.mask
	entry := tt.entries[idx]
	if entry.Hash == hash && entry.Flag != TTNone {
		return entry, true
	}
	return TTEntry{}, false
}

// Store saves a position in the transposition table
func (tt *TransTable) Store(hash uint64, depth int, score int, flag int, m Move) {
	idx := hash & tt.mask
	entry := &tt.entries[idx]
	age := int(uint8(tt.generation - tt.ages[idx]))
	replaceScore := depth + age*4

	// Blend freshness and depth so stale deep entries don't clog single-slot buckets.
	if entry.Flag == TTNone || entry.Hash == 0 || entry.Hash == hash || replaceScore >= int(entry.Depth) {
		entry.Hash = hash
		entry.Depth = int8(depth)
		entry.Score = int16(score)
		entry.Flag = uint8(flag)
		entry.Move = m
		tt.ages[idx] = tt.generation
	}
}

// Clear resets the transposition table
func (tt *TransTable) Clear() {
	for i := range tt.entries {
		tt.entries[i] = TTEntry{}
		tt.ages[i] = 0
	}
	tt.generation = 0
}

// Hashfull returns the permille of entries used
func (tt *TransTable) Hashfull() int {
	count := 0
	sample := 1000
	if tt.size < uint64(sample) {
		sample = int(tt.size)
	}
	for i := 0; i < sample; i++ {
		if tt.entries[i].Flag != TTNone && tt.entries[i].Hash != 0 {
			count++
		}
	}
	return count * 1000 / sample
}
