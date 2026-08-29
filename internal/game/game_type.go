package game

import (
	"fmt"
	"time"
)

type MatchType string

const (
	// Bullet Variants
	Bullet1Min   MatchType = "BULLET_1_0"
	Bullet1Min2S MatchType = "BULLET_1_2"
	Bullet2Min   MatchType = "BULLET_2_0"
	Bullet2Min1S MatchType = "BULLET_2_1"

	// Blitz Variants
	Blitz3Min   MatchType = "BLITZ_3_0"
	Blitz3Min2S MatchType = "BLITZ_3_2"
	Blitz5Min   MatchType = "BLITZ_5_0"

	// Classic
	Classic10Min MatchType = "CLASSIC_10_0"
	Classic20Min MatchType = "CLASSIC_20_0"
	Classic30Min MatchType = "CLASSIC_30_0"
)

func GetQueueName(mt MatchType) string {
	return fmt.Sprintf("queue:%s", mt)
}

type TimeConfig struct {
	Duration    time.Duration
	Increment   time.Duration
	MaxMoveTime time.Duration
}

func (mt MatchType) GetTimeConfig() TimeConfig {
	switch mt {
	case Bullet1Min:
		return TimeConfig{Duration: 1 * time.Minute, Increment: 0, MaxMoveTime: 15 * time.Second}
	case Bullet1Min2S:
		return TimeConfig{Duration: 1 * time.Minute, Increment: 2 * time.Second, MaxMoveTime: 15 * time.Second}
	case Bullet2Min:
		return TimeConfig{Duration: 2 * time.Minute, Increment: 0, MaxMoveTime: 20 * time.Second}
	case Bullet2Min1S:
		return TimeConfig{Duration: 2 * time.Minute, Increment: 1 * time.Second, MaxMoveTime: 20 * time.Second}
	case Blitz3Min:
		return TimeConfig{Duration: 3 * time.Minute, Increment: 0, MaxMoveTime: 30 * time.Second}
	case Blitz3Min2S:
		return TimeConfig{Duration: 3 * time.Minute, Increment: 2 * time.Second, MaxMoveTime: 30 * time.Second}
	case Blitz5Min:
		return TimeConfig{Duration: 5 * time.Minute, Increment: 0, MaxMoveTime: 45 * time.Second}
	case Classic10Min:
		return TimeConfig{Duration: 10 * time.Minute, Increment: 0, MaxMoveTime: 60 * time.Second}
	case Classic20Min:
		return TimeConfig{Duration: 20 * time.Minute, Increment: 0, MaxMoveTime: 90 * time.Second}
	case Classic30Min:
		return TimeConfig{Duration: 30 * time.Minute, Increment: 0, MaxMoveTime: 120 * time.Second}
	default:
		return TimeConfig{Duration: 5 * time.Minute, Increment: 0, MaxMoveTime: 30 * time.Second}
	}
}
