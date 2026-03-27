package engine

import (
	"time"
)

// TimeManager handles time allocation for the search
type TimeManager struct {
	StartTime  time.Time
	TimeLimit  time.Duration
	SoftLimit  time.Duration
	MaxDepth   int
	Infinite   bool
	Stopped    bool
	NodeCount  uint64
	CheckEvery uint64
}

// NewTimeManager creates a new time manager
func NewTimeManager() *TimeManager {
	return &TimeManager{
		MaxDepth:   64,
		CheckEvery: 2048,
	}
}

// SetFixedDepth sets the search to a fixed depth
func (tm *TimeManager) SetFixedDepth(depth int) {
	tm.MaxDepth = depth
	tm.Infinite = false
	tm.TimeLimit = 0
	tm.SoftLimit = 0
}

// SetTimeControl sets the search to time-based mode
func (tm *TimeManager) SetTimeControl(timeLeftMs, incMs, movesToGo int) {
	tm.MaxDepth = 64

	if movesToGo <= 0 {
		movesToGo = 30 // default move budget
	}

	// Use a soft target for normal cases and a harder cap for unstable positions.
	softMs := timeLeftMs/movesToGo + incMs*3/4

	// Don't spend more than ~35% of the clock as the normal target.
	maxSoft := timeLeftMs * 35 / 100
	if maxSoft <= 0 {
		maxSoft = timeLeftMs
	}
	if softMs > maxSoft {
		softMs = maxSoft
	}

	if softMs < 25 {
		softMs = 25
	}
	if softMs > timeLeftMs {
		softMs = timeLeftMs
	}

	hardMs := softMs * 2
	maxHard := timeLeftMs / 2
	if maxHard < softMs {
		maxHard = softMs
	}
	if hardMs > maxHard {
		hardMs = maxHard
	}

	tm.SoftLimit = time.Duration(softMs) * time.Millisecond
	tm.TimeLimit = time.Duration(hardMs) * time.Millisecond
}

// SetMoveTime sets fixed time per move
func (tm *TimeManager) SetMoveTime(ms int) {
	tm.MaxDepth = 64
	tm.TimeLimit = time.Duration(ms) * time.Millisecond
	tm.SoftLimit = tm.TimeLimit
}

// SetInfinite sets the search to infinite mode
func (tm *TimeManager) SetInfinite() {
	tm.MaxDepth = 64
	tm.Infinite = true
	tm.TimeLimit = 0
	tm.SoftLimit = 0
}

// Start begins the timer
func (tm *TimeManager) Start() {
	tm.StartTime = time.Now()
	tm.Stopped = false
	tm.NodeCount = 0
}

// ShouldStop checks if the search should stop
func (tm *TimeManager) ShouldStop() bool {
	if tm.Stopped {
		return true
	}
	if tm.Infinite {
		return false
	}
	if tm.TimeLimit > 0 {
		elapsed := time.Since(tm.StartTime)
		if elapsed >= tm.TimeLimit {
			tm.Stopped = true
			return true
		}
	}
	return false
}

// ShouldStopAfterIteration decides whether the next root iteration is worth the time.
func (tm *TimeManager) ShouldStopAfterIteration(bestMoveChanged bool, scoreSwing int, stableCount int, depth int) bool {
	if tm.Stopped {
		return true
	}
	if tm.Infinite || tm.TimeLimit == 0 {
		return false
	}

	elapsed := time.Since(tm.StartTime)
	if elapsed >= tm.TimeLimit {
		tm.Stopped = true
		return true
	}

	target := tm.SoftLimit
	if target == 0 {
		target = tm.TimeLimit
	}

	if tm.SoftLimit < tm.TimeLimit {
		switch {
		case bestMoveChanged:
			target = minDuration(tm.TimeLimit, tm.SoftLimit+tm.SoftLimit/2)
		case stableCount >= 2 && scoreSwing <= 12:
			target = maxDuration(25*time.Millisecond, tm.SoftLimit*3/4)
		case scoreSwing >= 60:
			target = minDuration(tm.TimeLimit, tm.SoftLimit+tm.SoftLimit/3)
		}
	}

	if depth <= 4 {
		target = maxDuration(target, tm.SoftLimit)
	}

	if elapsed >= target {
		tm.Stopped = true
		return true
	}
	return false
}

// CheckTime is called periodically during search
func (tm *TimeManager) CheckTime() bool {
	tm.NodeCount++
	if tm.NodeCount&(tm.CheckEvery-1) == 0 {
		return tm.ShouldStop()
	}
	return false
}

// Elapsed returns elapsed time in milliseconds
func (tm *TimeManager) Elapsed() int64 {
	return time.Since(tm.StartTime).Milliseconds()
}

// Stop forces the search to stop
func (tm *TimeManager) Stop() {
	tm.Stopped = true
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
