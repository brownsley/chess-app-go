package game

import (
	"sync"
	"time"
)

type PlayerTimer struct {
	Duration      time.Duration
	Increment     time.Duration
	Remaining     time.Duration
	MaxMoveTime   time.Duration
	lastTick      time.Time
	moveStartTime time.Time
	isRunning     bool
	mu            sync.Mutex
}

type MatchTimer struct {
	WhiteTimer  *PlayerTimer
	BlackTimer  *PlayerTimer
	IsWhiteTurn bool
	onTimeout   func(isWhite bool)
	stopChan    chan struct{}
	mu          sync.Mutex
}

func NewMatchTimer(config TimeConfig, onTimeout func(isWhite bool)) *MatchTimer {
	mt := &MatchTimer{
		WhiteTimer: &PlayerTimer{
			Duration:    config.Duration,
			Increment:   config.Increment,
			Remaining:   config.Duration,
			MaxMoveTime: config.MaxMoveTime,
		},
		BlackTimer: &PlayerTimer{
			Duration:    config.Duration,
			Increment:   config.Increment,
			Remaining:   config.Duration,
			MaxMoveTime: config.MaxMoveTime,
		},
		IsWhiteTurn: true,
		onTimeout:   onTimeout,
		stopChan:    make(chan struct{}),
	}

	go mt.watchTimeout()

	return mt
}

func (pt *PlayerTimer) Start() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if !pt.isRunning {
		now := time.Now()
		pt.lastTick = now
		pt.moveStartTime = now
		pt.isRunning = true
	}
}

func (pt *PlayerTimer) StopAndApplyIncrement() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.isRunning {
		elapsed := time.Since(pt.lastTick)
		pt.Remaining -= elapsed
		pt.Remaining += pt.Increment
		pt.isRunning = false
	}
}

func (pt *PlayerTimer) ResetMoveTimer() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.isRunning {
		pt.moveStartTime = time.Now()
		pt.lastTick = time.Now()
	}
}

func (pt *PlayerTimer) GetRemaining() time.Duration {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.isRunning {
		return pt.Remaining - time.Since(pt.lastTick)
	}
	return pt.Remaining
}

func (pt *PlayerTimer) IsMoveTimedOut() bool {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.isRunning && pt.MaxMoveTime > 0 {
		return time.Since(pt.moveStartTime) > pt.MaxMoveTime
	}
	return false
}

func (mt *MatchTimer) StartWhite() {
	mt.IsWhiteTurn = true
	mt.BlackTimer.StopAndApplyIncrement()
	mt.WhiteTimer.Start()
}

func (mt *MatchTimer) SwitchTurn() {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.IsWhiteTurn {
		mt.WhiteTimer.StopAndApplyIncrement()
		mt.IsWhiteTurn = false
		mt.BlackTimer.Start()
	} else {
		mt.BlackTimer.StopAndApplyIncrement()
		mt.IsWhiteTurn = true
		mt.WhiteTimer.Start()
	}
}

func (mt *MatchTimer) Stop() {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	select {
	case <-mt.stopChan:
	default:
		close(mt.stopChan)
	}

	mt.WhiteTimer.StopAndApplyIncrement()
	mt.BlackTimer.StopAndApplyIncrement()
}

func (mt *MatchTimer) StopChan() <-chan struct{} {
	return mt.stopChan
}

func (mt *MatchTimer) watchTimeout() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-mt.stopChan:
			return
		case <-ticker.C:
			mt.mu.Lock()
			isWhite := mt.IsWhiteTurn
			mt.mu.Unlock()

			if isWhite {
				if mt.WhiteTimer.GetRemaining() <= 0 || mt.WhiteTimer.IsMoveTimedOut() {
					mt.WhiteTimer.StopAndApplyIncrement()
					mt.Stop()
					if mt.onTimeout != nil {
						mt.onTimeout(true)
					}
					return
				}
			} else {
				if mt.BlackTimer.GetRemaining() <= 0 || mt.BlackTimer.IsMoveTimedOut() {
					mt.BlackTimer.StopAndApplyIncrement()
					mt.Stop()
					if mt.onTimeout != nil {
						mt.onTimeout(false)
					}
					return
				}
			}
		}
	}
}
